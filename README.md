# Terraform Provider for Cert-Central

The Cert-Central Terraform provider allows you to manage SSL/TLS certificate configurations, provision access keys, and fetch issued certificates directly within your Terraform workflows.

This provider is designed to interface with **cert-central**, a custom, containerized certificate manager solution currently supporting Let's Encrypt and ZeroSSL. Built on top of the `lego` Go library, it has the capacity to support more ACME providers in the future. The cert-central control plane is scheduled to be released and open-sourced in the near future.

> [!IMPORTANT]
> **Disclaimer:** This provider is for testing purposes only and is currently **not usable** if you do not have access to the private `cert-central` backend. It will become fully functional for general use once the control plane is released and open-sourced.

---


## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 1.0+
- [Go](https://golang.org) 1.22+ (for building the provider)

---

## Installation

The Cert-Central provider is published on the official Terraform Registry at [Menschomat/certcentral](https://registry.terraform.io/providers/Menschomat/certcentral/latest).

To use the provider in your Terraform configuration, add it to your `required_providers` block:

```hcl
terraform {
  required_providers {
    certcentral = {
      source  = "Menschomat/certcentral"
      version = "~> 1.0"
    }
  }
}

provider "certcentral" {
  address = "http://localhost:8080"
  token   = "your_admin_api_token"
}
```

---

## Local Development & Private Registries (Optional / Dev Only)

For local development or custom deployments, you can use the provider through either a **Gitea Private Registry** or a **Local Plugin Directory**.

### Option A: Gitea Package Registry (Development)
Gitea has a built-in, Terraform-compliant package registry. When you write:
```hcl
terraform {
  required_providers {
    certcentral = {
      source  = "gitea.dmz.k8s.menscho.space/m0space/certcentral"
      version = "~> 1.0.0"
    }
  }
}
```
Terraform resolves this by contacting `gitea.dmz.k8s.menscho.space` to download the pre-compiled binary matching your platform (e.g., `linux_amd64` or `darwin_arm64`).

To publish your compiled provider to Gitea's registry:
1. Compile and zip the provider binary:
   ```bash
   cd terraform-provider-certcentral
   go build -o terraform-provider-certcentral
   zip terraform-provider-certcentral_1.0.0_darwin_arm64.zip terraform-provider-certcentral
   ```
2. Upload it to your Gitea instance using `curl` (replace `YOUR_TOKEN` with a Gitea personal access token):
   ```bash
   curl --header "Authorization: token YOUR_TOKEN" \
        --upload-file terraform-provider-certcentral_1.0.0_darwin_arm64.zip \
        https://gitea.dmz.k8s.menscho.space/api/packages/m0space/terraform-provider/certcentral/1.0.0/darwin_arm64.zip
   ```

*Note: Gitea automatically handles Terraform service discovery under the hood at your domain's `.well-known/terraform.json` endpoint.*

### Option B: Local Plugin Installation
For offline use or local development, you can compile the provider and place it directly into your local Terraform plugin directory:

1. Build and copy the binary to your local plugin mirror directory (adjusting architecture name as needed):
   ```bash
   cd terraform-provider-certcentral
   go build -o terraform-provider-certcentral
   
   # For macOS (Apple Silicon):
   mkdir -p ~/.terraform.d/plugins/gitea.dmz.k8s.menscho.space/m0space/certcentral/1.0.0/darwin_arm64/
   cp terraform-provider-certcentral ~/.terraform.d/plugins/gitea.dmz.k8s.menscho.space/m0space/certcentral/1.0.0/darwin_arm64/terraform-provider-certcentral_v1.0.0
   ```

2. Reference your local source in your Terraform configuration:
   ```hcl
   terraform {
     required_providers {
       certcentral = {
         source  = "gitea.dmz.k8s.menscho.space/m0space/certcentral"
         version = "1.0.0"
       }
     }
   }
   ```
   When running `terraform init`, it will resolve the provider directly from your local cache directory.

For either option, you must configure the provider block with your `cert-central` endpoint and admin API key:

```hcl
provider "certcentral" {
  address = "http://localhost:8080"
  token   = "your_admin_api_token"
}
```

---

## Resources

### 1. `certcentral_team`

Manages a team configuration in `cert-central`. On creation, the server automatically generates a unique UUID v7 identifier for the team.

```hcl
resource "certcentral_team" "example" {
  name        = "example-team"
  description = "Example Team Description"
}
```

#### Argument Reference
* `name` (String, Required) - The name of the team.
* `description` (String, Optional) - A description of the team.

#### Attribute Reference
* `id` (String, Computed) - The unique UUID identifier of the team.

---

### 2. `certcentral_certificate`

Manages a certificate configuration in the background renewal scheduler. When created, `cert-central` automatically schedules DNS-01/HTTP-01 ACME challenges to issue the certificate.

```hcl
resource "certcentral_certificate" "example" {
  primary = "example.com"
  team_id = certcentral_team.example.id
  sans    = [
    "*.example.com",
    "www.example.com"
  ]
}
```

#### Argument Reference
* `primary` (String, Required) - The primary domain name for the certificate. Changing this triggers resource replacement.
* `team_id` (String, Required) - The unique UUID identifier of the team that owns this certificate configuration.
* `sans` (List of String, Optional) - Subject Alternative Names (SANs) for the certificate.

---

### 3. `certcentral_api_key`

Manages client API keys and access scopes in `cert-central`. On creation, the server automatically generates a secure 32-byte token and returns it in cleartext.

```hcl
resource "certcentral_api_key" "web_client" {
  description     = "web-client-token"
  allowed_domains = ["example.com"]
  allowed_teams   = [certcentral_team.example.id]
  admin           = false
}

# The generated cleartext token can be retrieved from state:
output "web_client_token" {
  value     = certcentral_api_key.web_client.cleartext_token
  sensitive = true
}
```

#### Argument Reference
* `description` (String, Optional) - A description of the API key configuration.
* `allowed_domains` (List of String, Optional) - Domains this standard token is allowed to fetch certificates for (ignored for admin tokens).
* `allowed_teams` (List of String, Optional) - The list of team UUIDs this token is scoped to (scopes configuration management if admin=true, or certificate retrieval if admin=false).
* `admin` (Boolean, Required) - If `true`, this token is used for configuration management (control plane). If `false`, this is a standard fetch token used to pull raw certificate and private key files.

#### API Key Scoping Matrix

The combination of the `admin` flag and the `allowed_teams` list dictates the token scoping level:

| Key Type | `admin` | `allowed_teams` | Authorized Actions |
| :--- | :--- | :--- | :--- |
| **Root Admin** | `true` | Empty / Omitted | Full CRUD access to all configurations (`/api/v1/config/*`) across all teams. |
| **Scoped Admin** | `true` | List of Team UUIDs | Manage configurations (certificates, keys) ONLY within the specified teams. Cannot manage root admin keys. |
| **Fetch Key (Standard)** | `false` | List of Team UUIDs | Retrieves raw PEM certificate chain and private key assets (`/api/v1/certificates`) for domains in `allowed_domains` AND `team_id` in `allowed_teams`. Cannot access configuration. |

#### Attribute Reference
* `cleartext_token` (String, Computed, Sensitive) - The generated plaintext token value. This is only returned by the server on creation and cannot be retrieved on subsequent reads.

---

## Data Sources

### `certcentral_certificate_data`

Retrieves PEM-encoded certificate chains and private keys once they are successfully issued by `cert-central`. You can pass these to load balancers, CDN, or file structures.

```hcl
data "certcentral_certificate_data" "example" {
  certificate_id = certcentral_certificate.example.id
}

# Example: Output the certificate body
output "certificate_pem" {
  value     = data.certcentral_certificate_data.example.certificate
  sensitive = true
}

# Example: Pass details to another resource (e.g. AWS ACM)
resource "aws_acm_certificate" "imported" {
  private_key       = data.certcentral_certificate_data.example.private_key
  certificate_body  = data.certcentral_certificate_data.example.certificate
}
```

#### Attribute Reference
* `certificate_id` (String, Required) - The unique configuration UUID of the certificate to fetch.
* `domain` (String, Computed) - The primary domain of the certificate.
* `sans` (List of String, Computed) - SANs associated with the certificate.
* `issued` (Boolean, Computed) - `true` if the certificate has been successfully issued.
* `certificate` (String, Computed, Sensitive) - The PEM-encoded certificate chain.
* `private_key` (String, Computed, Sensitive) - The PEM-encoded private key.
* `cert_filename` (String, Computed) - The filename of the certificate stored on the server.
* `key_filename` (String, Computed) - The filename of the private key stored on the server.

---

## Local Development & Testing

### 1. Build the Provider
To build the provider binary locally, run:

```bash
cd terraform-provider-certcentral
go build -o terraform-provider-certcentral
```

### 2. Run Tests
To run provider and client unit tests, execute:

```bash
go test -v ./...
```

---

## Releasing & Publishing

This repository uses [GoReleaser](https://goreleaser.com/) and GitHub Actions to automate compiles, checksum signatures, and release generation.

### Automated Releases (GitHub Actions)

When you tag a commit on the `main` branch with a semantic version (e.g., `v1.0.0`), the [Release GitHub Action workflow](.github/workflows/release.yml) is automatically triggered.

This workflow will:
1. Checkout the code at the tagged commit.
2. Set up the Go environment based on the version in `go.mod`.
3. Import the private GPG key used to sign the release (using the `GPG_PRIVATE_KEY` and `PASSPHRASE` repository secrets).
4. Run GoReleaser to compile binaries for `freebsd`, `windows`, `linux`, and `darwin`, calculate SHA256 checksums, sign the checksums, and publish the release and its artifacts to GitHub (fully compliant with the Terraform Registry layout requirements).

### Local GoReleaser Snapshot Builds

To test the GoReleaser configuration locally without publishing or signing, install GoReleaser and run:

```bash
goreleaser build --snapshot --clean
```

