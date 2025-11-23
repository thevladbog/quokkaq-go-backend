# Deployment Process

This document describes the automated deployment process for the QuokkaQ Backend.

## Overview

The deployment process is triggered automatically when changes are pushed to the `prod-release` branch. The CI/CD pipeline performs the following actions:

1. Automatically calculates the next semantic version
2. Updates the CHANGELOG.md with release information
3. Builds and pushes a Docker image to the registry
4. Creates a Git tag for the release
5. Deploys the new version to a Yandex Cloud VM
6. Creates a GitHub release

## Prerequisites

### Environment Variables

The following environment variables need to be configured in the CI/CD environment:

- `DOCKER_REGISTRY_USERNAME` - Username for the Docker registry
- `DOCKER_REGISTRY_PASSWORD` - Password for the Docker registry
- `YC_SERVICE_ACCOUNT_KEY` - Yandex Cloud service account key for deployment
- `VM_SSH_KEY` - SSH private key for accessing the Yandex Cloud VM

### Yandex Cloud Setup

The deployment targets a specific VM with ID: `fhmf3i36jq46rgl67sme`

## Deployment Process Details

### 1. Version Management

The system automatically determines the next version by:
1. Reading the current version from CHANGELOG.md
2. Incrementing the patch version number (e.g., 1.0.0 → 1.0.1)
3. Updating CHANGELOG.md with the release date
4. Adding a new "Unreleased" section for future changes

### 2. Docker Image Build

A Docker image is built using the multi-stage Dockerfile:
- Base image: golang:1.25.4-alpine for building
- Runtime image: alpine:latest for minimal size
- The final image contains only the compiled binary and necessary assets

The image is tagged with:
- The new version number (e.g., 1.0.1)
- `latest` tag

### 3. Deployment to Yandex Cloud

The deployment process:
1. Connects to the Yandex Cloud VM (ID: fhmf3i36jq46rgl67sme)
2. Pulls the new Docker image
3. Stops the current containers
4. Starts new containers with the updated image
5. Runs database migrations if needed

### 4. Release Management

After successful deployment:
- A Git tag is created (e.g., v1.0.1)
- A GitHub release is created with release notes
- The CHANGELOG.md is updated

## Manual Deployment

For manual deployment, follow these steps:

1. Ensure you have the required environment variables set
2. Build the Docker image:
   ```bash
   docker build -t quokkaq-backend:manual .
   ```
3. Push to registry:
   ```bash
   docker push quokkaq-backend:manual
   ```
4. Deploy to Yandex Cloud VM using your preferred method

## Troubleshooting

### Common Issues

1. **Version conflicts**: If the version in CHANGELOG.md is not properly formatted, the pipeline will fail.
   - Solution: Ensure the CHANGELOG.md follows the expected format with versions in brackets.

2. **Docker registry authentication**: If the Docker registry credentials are incorrect, the build will fail.
   - Solution: Verify the DOCKER_REGISTRY_USERNAME and DOCKER_REGISTRY_PASSWORD environment variables.

3. **Yandex Cloud deployment failures**: If the VM is not accessible or the credentials are incorrect, deployment will fail.
   - Solution: Verify the YC_SERVICE_ACCOUNT_KEY and VM_SSH_KEY environment variables.

### Rollback Process

To rollback to a previous version:
1. Identify the previous working version
2. Update the prod-release branch to point to the previous version's commit
3. The CI/CD pipeline will automatically deploy the previous version
4. Alternatively, manually deploy the previous Docker image to the VM

## Security Considerations

- All sensitive credentials are stored as encrypted environment variables
- Docker images are scanned for vulnerabilities before deployment
- SSH keys for VM access are rotated regularly
- All communication with Yandex Cloud uses encrypted connections

## Monitoring

After deployment, verify the application is running correctly:
1. Check the application health endpoint
2. Verify all services are responding
3. Check logs for any errors
4. Confirm database connectivity
5. Test critical user flows