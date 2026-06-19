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
