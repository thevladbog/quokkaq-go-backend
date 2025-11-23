# Traefik Reverse Proxy Setup Guide

## Overview

This guide covers the Traefik integration for QuokkaQ Backend, providing automatic SSL certificates via Let's Encrypt and reverse proxy functionality.

**Production Domain**: `https://api.quokkaq.v-b.tech`

---

## What is Traefik?

Traefik is a modern reverse proxy and load balancer that makes deploying microservices easy. Key features:

- ✅ **Automatic SSL**: Let's Encrypt integration for free SSL certificates
- ✅ **Auto-discovery**: Automatically detects Docker containers  
- ✅ **Load balancing**: Built-in load balancing support
- ✅ **Middleware**: CORS, authentication, rate limiting, etc.
- ✅ **Dashboard**: Web UI for monitoring

---

## Architecture

```
Internet
   ↓
Traefik (Port 80/443)
   ├→ api.quokkaq.v-b.tech → Backend API (Port 3001)
   ├→ s3.quokkaq.v-b.tech → MinIO API (Port 9000)
   ├→ minio.quokkaq.v-b.tech → MinIO Console (Port 9001)
   └→ traefik.quokkaq.v-b.tech → Traefik Dashboard
```

All HTTP traffic is automatically redirected to HTTPS.

---

## Quick Start

### 1. DNS Configuration

Point these domains to your server IP:

```
api.quokkaq.v-b.tech        → YOUR_SERVER_IP
s3.quokkaq.v-b.tech         → YOUR_SERVER_IP
minio.quokkaq.v-b.tech      → YOUR_SERVER_IP
traefik.quokkaq.v-b.tech    → YOUR_SERVER_IP
```

### 2. Create Traefik Network

```bash
docker network create traefik-public
```

### 3. Configure Environment

```bash
cp .env.prod.example .env.prod
nano .env.prod
```

Set `ACME_EMAIL` to your email for Let's Encrypt notifications.

### 4. Start Services

```bash
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

### 5. Verify SSL

Visit https://api.quokkaq.v-b.tech - you should see a valid SSL certificate 🔒

---

## Configuration Details

### Traefik Service Configuration

The Traefik service in `docker-compose.prod.yml` includes:

**Command-line arguments**:
- `--api.dashboard=true` - Enable dashboard
- `--providers.docker=true` - Enable Docker provider
- `--entrypoints.web.address=:80` - HTTP entrypoint
- `--entrypoints.websecure.address=:443` - HTTPS entrypoint
- `--certificatesresolvers.letsencrypt.acme.tlschallenge=true` - TLS challenge for SSL
- `--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}` - Let's Encrypt email

**Labels for Backend**:
```yaml
- "traefik.http.routers.quokkaq-api.rule=Host(`api.quokkaq.v-b.tech`)"
- "traefik.http.routers.quokkaq-api.tls.certresolver=letsencrypt"
```

### Middleware Configuration

#### CORS Middleware
Allows requests from frontend domain:
```yaml
- "traefik.http.middlewares.quokkaq-cors.headers.accesscontrolalloworiginlist=https://quokkaq.v-b.tech"
```

#### Security Headers
Adds HSTS and other security headers:
```yaml
- "traefik.http.middlewares.quokkaq-security.headers.stsSeconds=31536000"
- "traefik.http.middlewares.quokkaq-security.headers.forceSTSHeader=true"
```

---

## SSL Certificates

### Let's Encrypt Integration

**How it works**:
1. Traefik detects services with TLS configuration
2. Initiates ACME challenge with Let's Encrypt
3. Obtains and installs certificate
4. Automatically renews before expiration

**Certificate storage**: `/letsencrypt/acme.json` (inside Traefik container)

**Renewal**: Automatic, ~30 days before expiration

### Testing with Staging

To test without hitting Let's Encrypt rate limits, use staging server:

In `docker-compose.prod.yml`, uncomment:
```yaml
- "--certificatesresolvers.letsencrypt.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory"
```

**Note**: Staging certificates will show as invalid in browsers but are useful for testing.

### Rate Limits

Let's Encrypt limits:
- **50 certificates per registered domain per week**
- **5 duplicate certificates per week**

Use staging for testing to avoid hitting limits!

---

## Traefik Dashboard

### Accessing the Dashboard

URL: https://traefik.quokkaq.v-b.tech

Default credentials:
- Username: `admin`
- Password: (set via `htpasswd`, see below)

### Securing the Dashboard

Generate a new password:

```bash
# Install htpasswd
sudo apt install apache2-utils -y

# Generate password (replace 'your-password')
echo $(htpasswd -nb admin your-password) | sed -e s/\\$/\\$\\$/g
```

Update the label in `docker-compose.prod.yml`:
```yaml
- "traefik.http.middlewares.auth.basicauth.users=admin:$$apr1$$..."
```

### Disable Dashboard (Recommended for Production)

Remove or comment out dashboard labels in `docker-compose.prod.yml`:

```yaml
# Disable Traefik dashboard
# - "traefik.http.routers.traefik-dashboard.rule=Host(`traefik.quokkaq.v-b.tech`)"
# - ...
```

---

## Adding New Services

To add a new service behind Traefik:

1. **Add service to docker-compose.prod.yml**
2. **Connect to traefik-public network**
3. **Add Traefik labels**:

```yaml
myservice:
  image: myimage
  networks:
    - traefik-public
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.myservice.rule=Host(`myservice.quokkaq.v-b.tech`)"
    - "traefik.http.routers.myservice.entrypoints=websecure"
    - "traefik.http.routers.myservice.tls.certresolver=letsencrypt"
    - "traefik.http.services.myservice.loadbalancer.server.port=8080"
```

---

## Troubleshooting

### Certificate Not Issued

**Check Traefik logs**:
```bash
docker compose -f docker-compose.prod.yml logs traefik | grep -i "certificate\|acme"
```

**Common issues**:
- DNS not pointing to server
- Port 80/443 not accessible
- Hit Let's Encrypt rate limit (use staging)
- Invalid email in ACME_EMAIL

**Solution**:
```bash
# Remove certificates and retry
docker compose -f docker-compose.prod.yml down
docker volume rm quokkaq-go-backend_traefik_letsencrypt
docker compose -f docker-compose.prod.yml up -d
```

### Backend Not Accessible

**Check if Traefik can reach backend**:
```bash
docker compose -f docker-compose.prod.yml exec traefik wget -O- http://backend:3001
```

**Verify labels**:
```bash
docker inspect quokkaq-backend | grep -A 20 Labels
```

### HTTP Not Redirecting to HTTPS

**Check redirect config**:
```yaml
- "--entrypoints.web.http.redirections.entrypoint.to=websecure"
- "--entrypoints.web.http.redirections.entrypoint.scheme=https"
```

Test:
```bash
curl -I http://api.quokkaq.v-b.tech
# Should return 301/302 redirect to https://
```

---

## Advanced Configuration

### Custom Middleware

Add rate limiting:

```yaml
labels:
  - "traefik.http.middlewares.ratelimit.ratelimit.average=100"
  - "traefik.http.middlewares.ratelimit.ratelimit.burst=50"
  - "traefik.http.routers.quokkaq-api.middlewares=quokkaq-cors,quokkaq-security,ratelimit"
```

### IP Whitelist

Restrict access to specific IPs:

```yaml
labels:
  - "traefik.http.middlewares.ipwhitelist.ipwhitelist.sourcerange=1.2.3.4/32,5.6.7.8/32"
  - "traefik.http.routers.quokkaq-api.middlewares=ipwhitelist"
```

### Load Balancing Multiple Backends

```yaml
backend:
  deploy:
    replicas: 3
  # Traefik automatically load balances across replicas
```

---

## Monitoring

### View Active Routes

Dashboard: https://traefik.quokkaq.v-b.tech

Or via API:
```bash
curl http://localhost:8080/api/http/routers
```

### Check Certificate Expiry

```bash
echo | openssl s_client -servername api.quokkaq.v-b.tech -connect api.quokkaq.v-b.tech:443 2>/dev/null | openssl x509 -noout -dates
```

### Monitor Logs

```bash
# Real-time logs
docker compose -f docker-compose.prod.yml logs -f traefik

# Access logs only
docker compose -f docker-compose.prod.yml logs traefik | grep "GET\|POST"
```

---

## Security Best Practices

1. ✅ **Disable dashboard in production** or use strong authentication
2. ✅ **Use HTTPS only** (HTTP redirects configured)
3. ✅ **Enable security headers** (HSTS, etc.)
4. ✅ **Restrict Traefik dashboard access** (IP whitelist or VPN)
5. ✅ **Regular updates** - Keep Traefik image updated
6. ✅ **Monitor certificate expiry** - Set up alerts
7. ✅ **Backup acme.json** - Contains all SSL certificates

---

## Migration from Existing Setup

If you're migrating from nginx or another reverse proxy:

1. **Document current routes** and SSL certificates
2. **Start Traefik** alongside existing proxy (different ports)
3. **Test thoroughly** before switching DNS
4. **Update DNS** to point to Traefik server
5. **Decommission old proxy** after verification

---

## Resources

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Let's Encrypt Rate Limits](https://letsencrypt.org/docs/rate-limits/)
- [Docker Provider](https://doc.traefik.io/traefik/providers/docker/)
- [Middlewares](https://doc.traefik.io/traefik/middlewares/overview/)

---

## Quick Reference

```bash
# Start with Traefik
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

# View Traefik logs
docker compose -f docker-compose.prod.yml logs -f traefik

# Restart Traefik
docker compose -f docker-compose.prod.yml restart traefik

# Check certificate
openssl s_client -connect api.quokkaq.v-b.tech:443 -servername api.quokkaq.v-b.tech

# Access dashboard (if enabled)
https://traefik.quokkaq.v-b.tech
```
