# Repository Secrets Setup

This document provides instructions for setting up the required secrets in the repository for the CI/CD pipeline to function correctly.

## Overview

The CI/CD pipeline requires several secrets to be configured in the repository settings. These secrets are used for authentication with Yandex Cloud services and other external systems.

## Required Secrets

The following secrets must be configured in the repository settings:

| Secret Name | Description | How to Obtain |
|-------------|-------------|---------------|
| `YC_REGISTRY_USERNAME` | Username for Yandex Cloud Container Registry | Usually the ID of the service account |
| `YC_REGISTRY_PASSWORD` | Password for Yandex Cloud Container Registry | IAM token or API key for the service account |
| `YC_REGISTRY_ID` | Yandex Cloud Container Registry ID | From Yandex Cloud Console |
| `YC_SERVICE_ACCOUNT_KEY` | Yandex Cloud service account key | JSON key file for the service account |
| `VM_SSH_KEY` | SSH private key for accessing the Yandex Cloud VM | Generate SSH key pair for VM access |
| `ACME_EMAIL` | Email for Let's Encrypt SSL certificates | Your email address |
| `POSTGRES_USER` | PostgreSQL username | Choose a secure username |
| `POSTGRES_PASSWORD` | PostgreSQL password | Generate a strong password |
| `POSTGRES_DB` | PostgreSQL database name | Usually `quokkaq` |
| `REDIS_PASSWORD` | Redis password | Generate a strong password |
| `MINIO_ROOT_USER` | MinIO root username | Choose a secure username |
| `MINIO_ROOT_PASSWORD` | MinIO root password | Generate a strong password |
| `AWS_S3_BUCKET` | S3 bucket name | Usually `quokkaq-materials` |
| `SMTP_HOST` | SMTP server hostname | Your SMTP provider's hostname |
| `SMTP_PORT` | SMTP server port | Usually 587 or 465 |
| `SMTP_USER` | SMTP username | Your SMTP username |
| `SMTP_PASS` | SMTP password | Your SMTP password |
| `SMTP_FROM` | SMTP from address | Email address for sending emails |
| `SMTP_SECURE` | SMTP secure setting | `true` or `false` |
| `JWT_SECRET` | JWT secret for authentication | Generate a strong secret |
| `APP_BASE_URL` | Application base URL | Your frontend URL |

## Setting Up Secrets

### 1. Access Repository Settings

1. Navigate to the repository on SourceCraft
2. Click on "Settings" in the repository menu
3. Select "Secrets" from the settings menu

### 2. Add Required Secrets

For each required secret:

1. Click "New secret"
2. Enter the secret name (exactly as listed above)
3. Enter the secret value
4. Click "Add secret"

### 3. Yandex Cloud Container Registry Setup

#### Obtain Registry ID

1. Go to Yandex Cloud Console
2. Navigate to "Container Registry"
3. Copy the registry ID

#### Create Service Account for Registry Access

1. In Yandex Cloud Console, go to "Service Accounts"
2. Create a new service account
3. Assign the role `container-registry.images.pusher`
4. Create an authorized key for the service account

#### Configure Registry Credentials

- `YC_REGISTRY_USERNAME`: Use the service account ID
- `YC_REGISTRY_PASSWORD`: Use the authorized key content (full JSON)
- `YC_REGISTRY_ID`: Use the registry ID from Container Registry

### 4. Yandex Cloud VM Access Setup

#### Generate SSH Key Pair

```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 4096 -f ~/.ssh/quokkaq-deploy

# The private key will be used as VM_SSH_KEY secret
# The public key will be added to the VM
```

#### Configure VM Access

1. Add the public key to the VM's authorized_keys file
2. Store the private key content as the `VM_SSH_KEY` secret

### 5. Service Account for Yandex Cloud Resources

#### Create Service Account

1. In Yandex Cloud Console, go to "Service Accounts"
2. Create a new service account with the following roles:
   - `compute.admin` - for VM management
   - `container-registry.images.puller` - for pulling images
   - `iam.serviceAccounts.user` - for service account management

#### Generate Key

1. Create an authorized key for the service account:
   ```bash
   yc iam key create --service-account-name <service-account-name> --output key.json
   ```

2. Store the key content as the `YC_SERVICE_ACCOUNT_KEY` secret

### 6. Production Environment Variables

Set up the following production environment variables as secrets:

#### Database Configuration
- `POSTGRES_USER`: Choose a secure username (e.g., `quokkaq_user`)
- `POSTGRES_PASSWORD`: Generate a strong password
- `POSTGRES_DB`: Usually `quokkaq`

#### Redis Configuration
- `REDIS_PASSWORD`: Generate a strong password

#### MinIO Configuration
- `MINIO_ROOT_USER`: Choose a secure username (e.g., `quokkaq_minio`)
- `MINIO_ROOT_PASSWORD`: Generate a strong password
- `AWS_S3_BUCKET`: Usually `quokkaq-materials`

#### SMTP Configuration
- `SMTP_HOST`: Your SMTP provider's hostname (e.g., `smtp.yandex.ru`)
- `SMTP_PORT`: Usually 587 or 465
- `SMTP_USER`: Your SMTP username
- `SMTP_PASS`: Your SMTP password
- `SMTP_FROM`: Email address for sending emails (e.g., `noreply@quokkaq.v-b.tech`)
- `SMTP_SECURE`: `true` or `false`

#### Application Configuration
- `ACME_EMAIL`: Your email address for Let's Encrypt notifications
- `JWT_SECRET`: Generate a strong secret (at least 32 characters)
- `APP_BASE_URL`: Your frontend URL (e.g., `https://quokkaq.v-b.tech`)

## Secret Values Format

### YC_REGISTRY_PASSWORD

Этот секрет должен содержать полное содержимое JSON-файла авторизованного ключа сервисного аккаунта. В него входят как открытая, так и закрытая части ключа, необходимые для аутентификации в Container Registry.

Пример формата:
```json
{
  "id": "aje...example",
  "created_at": "2023-01-01T00:00:00Z",
  "key_algorithm": "RSA_2048",
  "public_key": "-----BEGIN PUBLIC KEY-----
...
-----END PUBLIC KEY-----",
  "private_key": "-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----"
}
```

### YC_SERVICE_ACCOUNT_KEY

Этот секрет должен содержать полное содержимое JSON-файла авторизованного ключа сервисного аккаунта, который используется для управления ресурсами Yandex Cloud.

Пример формата:
```json
{
  "id": "aje...example",
  "created_at": "2023-01-01T00:00:00Z",
  "key_algorithm": "RSA_2048",
  "public_key": "-----BEGIN PUBLIC KEY-----
...
-----END PUBLIC KEY-----",
  "private_key": "-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----"
}
```

### VM_SSH_KEY

Этот секрет должен содержать полное содержимое приватного SSH-ключа для доступа к виртуальной машине в Yandex Cloud.

Пример формата:
```
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

### Other Secrets

All other secrets should contain plain text values as appropriate for their use case.

## Testing Secret Configuration

After configuring all secrets, you can test the configuration by triggering the CI/CD pipeline:

1. Make a small change to the code
2. Create a pull request to the `prod-release` branch with a proper title format
3. Merge the pull request
4. Monitor the pipeline execution in the Actions tab

## Security Best Practices

### Secret Rotation

1. Regularly rotate all secrets (recommended every 90 days)
2. Update secrets in both Yandex Cloud and repository settings
3. Test the pipeline after secret rotation

### Access Control

1. Limit access to repository secrets to authorized personnel only
2. Use separate service accounts for different purposes
3. Regularly audit secret access logs

### Storage

1. Never store secrets in source code
2. Use encrypted storage for secrets
3. Use temporary credentials when possible

## Troubleshooting

### Common Issues

1. **Authentication failures**: Check that all secrets are correctly configured and not expired
2. **Permission denied**: Verify that service accounts have the required roles
3. **Connection failures**: Check network access and firewall rules
4. **Missing environment variables**: Ensure all required secrets are configured

### Debugging Steps

1. Check pipeline logs for detailed error messages
2. Verify secret values are correctly formatted
3. Test secret values outside the pipeline if possible
4. Ensure secrets are available to the workflow

## Conclusion

Properly configured secrets are essential for the CI/CD pipeline to function correctly. Following the instructions in this document will ensure that the pipeline can authenticate with Yandex Cloud services and deploy the application successfully.

**Важно**: Для секрета `YC_REGISTRY_PASSWORD` необходимо использовать полное содержимое JSON-файла авторизованного ключа, включая как открытую, так и закрытую части ключа. Именно закрытая часть ключа используется для аутентификации в Container Registry.