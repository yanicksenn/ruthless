terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0, < 6.0"
    }
  }
}

resource "google_artifact_registry_repository" "repo" {
  project       = var.project_id
  location      = var.region
  repository_id = "${var.project_name}-repo-${var.environment}"
  description   = "Docker repository for ${var.project_name} ${var.environment}"
  format        = "DOCKER"
}
