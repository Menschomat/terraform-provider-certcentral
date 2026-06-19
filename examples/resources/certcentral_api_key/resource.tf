resource "certcentral_api_key" "web_client" {
  token           = "$argon2id$v=19$m=65536,t=3,p=2$..." # Argon2id hash of the token
  allowed_domains = ["example.com"]
  admin           = false
}
