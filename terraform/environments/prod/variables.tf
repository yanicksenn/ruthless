variable "project_name" {
  type        = string
  description = "The name of the project."
  default     = "ruthless"
}

variable "project_id" {
  type        = string
  description = "The GCP project ID."
}

variable "environment" {
  type        = string
  description = "The environment (e.g., prod, prod)."
  default     = "prod"
}

variable "region" {
  type        = string
  description = "The GCP region."
  default     = "us-central1"
}

variable "db_tier" {
  type        = string
  description = "The machine tier for the Postgres instance."
  default     = "db-f1-micro"
}



variable "google_client_id" {
  type        = string
  description = "Google OAuth Client ID."
}

variable "google_client_secret" {
  type        = string
  description = "Google OAuth Client Secret."
  sensitive   = true
}

variable "auth_secret" {
  type        = string
  description = "Secret used for signing internal tokens."
  sensitive   = true
}

variable "domain" {
  type        = string
  description = "The domain name for the application."
}
