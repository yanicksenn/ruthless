variable "project_name" {
  type        = string
  description = "The name of the project."
}

variable "project_id" {
  type        = string
  description = "The GCP project ID."
}

variable "environment" {
  type        = string
  description = "The environment (e.g., preprod, prod)."
}

variable "region" {
  type        = string
  description = "The GCP region."
}

variable "tier" {
  type        = string
  description = "The machine tier for the Postgres instance. e.g. db-f1-micro"
}
