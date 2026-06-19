resource "certcentral_certificate" "example" {
  primary = "example.com"
  sans    = [
    "*.example.com",
    "www.example.com"
  ]
}
