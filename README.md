# kubeseal-encrypt-string

A simple CLI tool to encrypt strings using Bitnami's Sealed Secrets for Kubernetes. This tool provides a simple way to encrypt sensitive data that can be safely stored in Git repositories and decrypted only within your Kubernetes cluster.

## Prerequisites

Before using this tool, ensure you have the following installed:

- **kubectl** (v1.29+)

Installing [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

- **kubeseal** (v0.24.0+)

Installing [kubeseal](https://github.com/bitnami-labs/sealed-secrets?tab=readme-ov-file#kubeseal)

## Features

- Automatic validation of encrypted secrets
- Namespace-scoped encryption

## Installation

## Building from Source

```bash
git clone https://github.com/mataberat/kubeseal-encrypt-string.git
cd kubeseal-encrypt-string
make build
```

## Usage

```bash
# Basic usage with required flags
kubeseal-encrypt-string --key mysecret --value supersecret --namespace production

# Using custom controller namespace
kubeseal-encrypt-string --key mysecret --value supersecret --namespace production --controller-namespace sealed-secrets

# Using custom controller name and namespace
kubeseal-encrypt-string --key mysecret --value supersecret --namespace production --controller-namespace sealed-secrets --controller-name sealed-secrets

# Using environment variables for controller config
export SEALED_SECRETS_CONTROLLER_NAMESPACE=sealed-secrets
export SEALED_SECRETS_CONTROLLER_NAME=sealed-secrets
kubeseal-encrypt-string --key mysecret --value supersecret --namespace production
```
