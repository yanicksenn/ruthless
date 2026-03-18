variable "project_name" {
  type        = string
  description = "The name of the project."
}

variable "environment" {
  type        = string
  description = "The environment (e.g., preprod, prod)."
}

variable "region" {
  type        = string
  description = "The GCP region."
}

variable "backend_image_url" {
  type        = string
  description = "The URL of the backend Docker image in Artifact Registry."
}

variable "frontend_image_url" {
  type        = string
  description = "The URL of the frontend Docker image in Artifact Registry."
}

variable "db_connection_name" {
  type        = string
  description = "The connection name of the Cloud SQL instance."
}

variable "db_user" {
  type        = string
  description = "The database user."
}

variable "db_password" {
  type        = string
  description = "The database password."
  sensitive   = true
}

variable "db_name" {
  type        = string
  description = "The name of the database."
}



variable "auth_type" {
  type        = string
  description = "Authentication provider (e.g., google)."
  default     = "google"
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

variable "project_id" {
  type        = string
  description = "The GCP project ID."
}
