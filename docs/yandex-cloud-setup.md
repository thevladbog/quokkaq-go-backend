# Yandex Cloud Setup for Deployment

This document provides instructions for setting up Yandex Cloud infrastructure to support the automated deployment pipeline.

## Overview

The deployment pipeline targets a specific Virtual Machine (VM) in Yandex Cloud with ID: `fhmf3i36jq46rgl67sme`. This document describes the required setup for this VM to support automated deployments.

## Prerequisites

Before setting up the deployment infrastructure, ensure you have:

1. A Yandex Cloud account with appropriate permissions
2. A created VM with ID: `fhmf3i36jq46rgl67sme`
3. Network access to the VM
4. A service account with necessary permissions
5. A Container Registry set up in Yandex Cloud

## Container Registry Setup

### Create Container Registry

1. In the Yandex Cloud Console, navigate to the Container Registry section
2. Create a new container registry
3. Note the registry ID for use in the CI/CD pipeline

### Configure Registry Access

1. Create an IAM token for the service account with registry access:
   ```bash
   yc iam key create --service-account-name <service-account-name> --output key.json
   ```

2. Store the full JSON content of the key file and provide it to the CI/CD pipeline as `YC_REGISTRY_PASSWORD`.

## VM Configuration

### Base Setup

The VM should be configured with the following specifications:

- **Operating System**: Ubuntu 20.04 LTS or later
- **CPU**: 2 vCPUs or more
- **RAM**: 4 GB or more
- **Disk**: 20 GB or more of persistent storage
- **Network**: Public IP address or access through Load Balancer

### Required Software

Install the following software on the VM:

1. **Docker Engine**:
   ```bash
   # Update package index
   sudo apt-get update
   
   # Install prerequisites
   sudo apt-get install \
       ca-certificates \
       curl \
       gnupg \
       lsb-release
   
   # Add Docker's official GPG key
   sudo mkdir -p /etc/apt/keyrings
   curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
   
   # Set up the repository
   echo \
     "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
     $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
   
   # Install Docker Engine
   sudo apt-get update
   sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin
   ```

2. **Yandex Cloud CLI** (for registry authentication):
   ```bash
   curl https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash
   ```

3. **Docker Compose** (if not included with Docker Engine):
   ```bash
   # Install Docker Compose
   sudo apt install docker-compose-plugin
   ```

4. **SSH Access**:
   Ensure SSH access is configured for the CI/CD pipeline to connect to the VM.

### User Permissions

Create a dedicated user for deployment operations:

```bash
# Create deployment user
sudo adduser deploy

# Add user to docker group
sudo usermod -aG docker deploy

# Set up SSH key authentication for the deploy user
sudo mkdir -p /home/deploy/.ssh
sudo cp /path/to/public/key /home/deploy/.ssh/authorized_keys
sudo chown -R deploy:deploy /home/deploy/.ssh
sudo chmod 700 /home/deploy/.ssh
sudo chmod 600 /home/deploy/.ssh/authorized_keys
```

## Service Account Setup

### Create Service Account

1. In the Yandex Cloud Console, navigate to the IAM section
2. Create a new service account with the following roles:
   - `container-registry.images.puller` - for pulling images from Container Registry
   - `compute.admin` - for VM management
   - `iam.serviceAccounts.user` - for service account management
   - `vpc.publicAdmin` - for network management (if needed)

### Generate Authentication Key

1. Create an authorized key for the service account:
   ```bash
   yc iam key create --service-account-name <service-account-name> --output key.json
   ```

2. Store the full JSON content of the key file and provide it to the CI/CD pipeline as `YC_SERVICE_ACCOUNT_KEY`.

## Environment Configuration

### Environment Variables

Set up the following environment variables on the VM:

```bash
# Database Configuration
export DATABASE_URL=postgresql://postgres:password@localhost:5432/quokkaq

# Server Configuration
export PORT=3001
export APP_BASE_URL=https://your-domain.com

# MinIO / AWS S3 Configuration
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin
export AWS_REGION=us-east-1
export AWS_S3_BUCKET=quokkaq-materials
export AWS_ENDPOINT=http://localhost:9000

# SMTP Configuration
export SMTP_HOST=smtp.yandex.ru
export SMTP_PORT=587
export SMTP_USER=your-email@yandex.ru
export SMTP_PASS=your-password
export SMTP_FROM=noreply@your-domain.com
export SMTP_SECURE=true
```

These variables can be stored in a `.env` file in the deployment directory.

### Docker Registry Authentication

Configure Docker to authenticate with Yandex Cloud Container Registry:

```bash
# Login to Yandex Cloud Container Registry
echo $YC_REGISTRY_PASSWORD | docker login --username $YC_REGISTRY_USERNAME --password-stdin cr.yandex
```

This should be done as the deploy user.

## Deployment Directory Structure

Set up the following directory structure on the VM:

```
/home/deploy/
├── quokkaq/
│   ├── .env              # Environment variables
│   ├── docker-compose.yml # Production Docker Compose configuration
│   └── logs/              # Application logs
└── scripts/
    ├── deploy.sh         # Deployment script
    └── health-check.sh   # Health check script
```

### Deployment Script

Create a deployment script at `/home/deploy/scripts/deploy.sh`:

```bash
#!/bin/bash

# Load environment variables
source /home/deploy/quokkaq/.env

# Authenticate with Yandex Cloud Container Registry
echo $YC_REGISTRY_PASSWORD | docker login --username $YC_REGISTRY_USERNAME --password-stdin cr.yandex

# Pull the latest Docker image
docker pull cr.yandex/$YC_REGISTRY_ID/quokkaq-backend:$1

# Stop current services
docker compose -f /home/deploy/quokkaq/docker-compose.yml down

# Start new services
TAG=$1 docker compose -f /home/deploy/quokkaq/docker-compose.yml up -d

# Run database migrations
docker compose -f /home/deploy/quokkaq/docker-compose.yml exec backend ./migrate

# Check service health
sleep 30
curl -f http://localhost:3001/health || exit 1

echo "Deployment completed successfully"
```

### Health Check Script

Create a health check script at `/home/deploy/scripts/health-check.sh`:

```bash
#!/bin/bash

# Check if required services are running
if ! docker compose -f /home/deploy/quokkaq/docker-compose.yml ps | grep -q "running"; then
  echo "Services are not running"
  exit 1
fi

# Check application health endpoint
if ! curl -f http://localhost:3001/health; then
  echo "Application health check failed"
  exit 1
fi

echo "All services are healthy"
```

## Docker Compose Configuration

Create a production Docker Compose configuration at `/home/deploy/quokkaq/docker-compose.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:16-alpine
    container_name: quokkaq-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres}
      POSTGRES_DB: quokkaq
    ports:
      - "127.0.0.1:5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - quokkaq-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis for background jobs
  redis:
    image: redis:7-alpine
    container_name: quokkaq-redis
    restart: unless-stopped
    ports:
      - "127.0.0.1:6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - quokkaq-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # MinIO for file storage
  minio:
    image: minio/minio:latest
    container_name: quokkaq-minio
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD:-minioadmin}
    ports:
      - "127.0.0.1:9000:9000"
      - "127.0.0.1:9001:9001"
    volumes:
      - minio_data:/data
    networks:
      - quokkaq-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

  # QuokkaQ Backend API
  backend:
    image: cr.yandex/${YC_REGISTRY_ID}/quokkaq-backend:${TAG:-latest}
    container_name: quokkaq-backend
    restart: unless-stopped
    ports:
      - "3001:3001"
    environment:
      DATABASE_URL: postgresql://postgres:${POSTGRES_PASSWORD:-postgres}@postgres:5432/quokkaq
      PORT: 3001
      APP_BASE_URL: ${APP_BASE_URL:-http://localhost:3000}
      
      # MinIO Configuration
      AWS_ACCESS_KEY_ID: ${MINIO_ROOT_USER:-minioadmin}
      AWS_SECRET_ACCESS_KEY: ${MINIO_ROOT_PASSWORD:-minioadmin}
      AWS_REGION: us-east-1
      AWS_S3_BUCKET: quokkaq-materials
      AWS_ENDPOINT: http://minio:9000
      
      # SMTP Configuration
      SMTP_HOST: ${SMTP_HOST}
      SMTP_PORT: ${SMTP_PORT}
      SMTP_USER: ${SMTP_USER}
      SMTP_PASS: ${SMTP_PASS}
      SMTP_FROM: ${SMTP_FROM}
      SMTP_SECURE: ${SMTP_SECURE:-false}
      
      # Redis Configuration
      REDIS_URL: redis://redis:6379/0
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      minio:
        condition: service_healthy
    networks:
      - quokkaq-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3001/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  quokkaq-network:
    driver: bridge

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local
  minio_data:
    driver: local
```

## Security Considerations

### Network Security

1. Restrict access to the VM using security groups:
   - Allow SSH access only from trusted IP addresses
   - Allow HTTP/HTTPS access from necessary sources
   - Restrict database access to internal network only

2. Use private networks where possible:
   - Place database and other internal services on private networks
   - Use load balancers for public access

### Data Security

1. Enable encryption at rest for persistent volumes
2. Use TLS for all network communications
3. Regularly rotate passwords and API keys
4. Implement proper backup and recovery procedures

### Access Control

1. Use separate service accounts for different services
2. Implement role-based access control
3. Regularly audit access logs
4. Use multi-factor authentication for administrative access

## Monitoring and Logging

### System Monitoring

Set up monitoring for the following metrics:

1. CPU and memory usage
2. Disk space utilization
3. Network traffic
4. Application response times
5. Error rates

### Log Management

Configure log rotation and retention:

```bash
# Create log rotation configuration
sudo tee /etc/logrotate.d/quokkaq << EOF
/home/deploy/quokkaq/logs/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 deploy deploy
}
EOF
```

## Backup and Recovery

### Database Backup

Set up regular database backups:

```bash
# Create backup script
sudo tee /home/deploy/scripts/backup-db.sh << EOF
#!/bin/bash
docker exec quokkaq-postgres pg_dump -U postgres quokkaq > /home/deploy/backups/quokkaq-$(date +%Y%m%d-%H%M%S).sql
EOF

# Make script executable
chmod +x /home/deploy/scripts/backup-db.sh

# Schedule backups with cron
echo "0 2 * * * /home/deploy/scripts/backup-db.sh" | crontab -
```

### File Backup

Set up regular file backups for MinIO data:

```bash
# Create backup script
sudo tee /home/deploy/scripts/backup-minio.sh << EOF
#!/bin/bash
docker exec quokkaq-minio mc cp --recursive local/quokkaq-materials /backups/quokkaq-materials-$(date +%Y%m%d-%H%M%S)
EOF

# Make script executable
chmod +x /home/deploy/scripts/backup-minio.sh

# Schedule backups with cron
echo "0 3 * * * /home/deploy/scripts/backup-minio.sh" | crontab -
```

## Testing the Setup

### Initial Deployment Test

1. Manually run the deployment script with a test version:
   ```bash
   /home/deploy/scripts/deploy.sh 1.0.0
   ```

2. Verify all services are running:
   ```bash
   docker compose -f /home/deploy/quokkaq/docker-compose.yml ps
   ```

3. Check application health:
   ```bash
   curl -f http://localhost:3001/health
   ```

### Rollback Test

1. Deploy a known good version:
   ```bash
   /home/deploy/scripts/deploy.sh 0.9.9
   ```

2. Verify the application is working correctly

## Maintenance

### Regular Updates

1. Update the base OS regularly:
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

2. Update Docker and related tools:
   ```bash
   sudo apt update && sudo apt install docker-ce docker-compose-plugin
   ```

3. Restart services after updates:
   ```bash
   sudo systemctl restart docker
   ```

### Performance Tuning

1. Monitor resource usage and adjust VM specifications as needed
2. Optimize database queries and indexes
3. Configure appropriate caching strategies
4. Implement connection pooling for database connections

## Troubleshooting

### Common Issues

1. **Docker daemon not starting**:
   - Check system logs: `journalctl -u docker.service`
   - Verify Docker installation: `docker --version`
   - Restart Docker service: `sudo systemctl restart docker`

2. **Insufficient disk space**:
   - Check disk usage: `df -h`
   - Clean up unused Docker data: `docker system prune -a`
   - Remove old backups if necessary

3. **Network connectivity issues**:
   - Check firewall rules
   - Verify security group settings
   - Test connectivity to required services

### Logs and Diagnostics

1. Check Docker logs:
   ```bash
   docker logs quokkaq-backend
   ```

2. Check system logs:
   ```bash
   journalctl -u docker.service
   ```

3. Check application logs:
   ```bash
   tail -f /home/deploy/quokkaq/logs/application.log
   ```

## Conclusion

This setup provides a robust foundation for deploying the QuokkaQ Backend application to Yandex Cloud. The configuration supports automated deployments, monitoring, and maintenance operations. Regular review and updates to this setup will ensure continued reliability and security of the deployed application.