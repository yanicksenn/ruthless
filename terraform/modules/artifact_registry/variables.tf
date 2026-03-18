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

variable "project_id" {
  type        = string
  description = "The GCP project ID."
}
