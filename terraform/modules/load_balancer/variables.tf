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

variable "domain" {
  type        = string
  description = "The domain name for the SSL certificate and routing."
}

variable "frontend_service_name" {
  type        = string
  description = "The name of the frontend Cloud Run service."
}

variable "backend_service_name" {
  type        = string
  description = "The name of the backend Cloud Run service."
}

variable "project_id" {
  type        = string
  description = "The GCP project ID."
}
