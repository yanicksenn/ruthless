terraform {
  required_version = ">= 1.0.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0, < 6.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

locals {
  services = [
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "iam.googleapis.com"
  ]
}

resource "google_project_service" "services" {
  for_each           = toset(local.services)
  project            = var.project_id
  service            = each.key
  disable_on_destroy = false
}

module "artifact_registry" {
  source       = "../../modules/artifact_registry"
  depends_on   = [google_project_service.services]
  project_name = var.project_name
  project_id   = var.project_id
  environment  = var.environment
  region       = var.region
}

module "database" {
  source       = "../../modules/database"
  depends_on   = [google_project_service.services]
  project_name = var.project_name
  project_id   = var.project_id
  environment  = var.environment
  region       = var.region
  tier         = var.db_tier
}

resource "null_resource" "build_and_push_images" {
  triggers = {
    always_run = timestamp()
  }

  depends_on = [module.artifact_registry]

  provisioner "local-exec" {
    command = <<EOT
      set -e
      export DOCKER_BUILDKIT=1
      gcloud auth configure-docker ${var.region}-docker.pkg.dev --quiet
      
      docker build --platform linux/amd64 -t ${module.artifact_registry.repository_url}/backend:latest -f ../../../backend/Dockerfile ../../../
      docker push ${module.artifact_registry.repository_url}/backend:latest
      
      docker build --platform linux/amd64 \
        --build-arg VITE_API_BASE_URL="https://${var.domain}" \
        --build-arg VITE_GOOGLE_CLIENT_ID="${var.google_client_id}" \
        -t ${module.artifact_registry.repository_url}/frontend:latest \
        -f ../../../frontend/Dockerfile ../../../frontend
      docker push ${module.artifact_registry.repository_url}/frontend:latest
    EOT
  }
}

module "cloud_run" {
  source               = "../../modules/cloud_run"
  depends_on           = [google_project_service.services, null_resource.build_and_push_images]
  project_name         = var.project_name
  project_id           = var.project_id
  environment          = var.environment
  region               = var.region
  backend_image_url    = "${module.artifact_registry.repository_url}/backend:latest"
  frontend_image_url   = "${module.artifact_registry.repository_url}/frontend:latest"
  db_connection_name   = module.database.instance_connection_name
  db_user              = module.database.db_user
  db_password          = module.database.db_password
  db_name              = module.database.db_name
  domain               = var.domain
  google_client_id     = var.google_client_id
  google_client_secret = var.google_client_secret
  auth_secret          = var.auth_secret
}
module "load_balancer" {
  source                = "../../modules/load_balancer"
  depends_on            = [google_project_service.services]
  project_name          = var.project_name
  project_id            = var.project_id
  environment           = var.environment
  region                = var.region
  domain                = var.domain
  frontend_service_name = module.cloud_run.frontend_service_name
  backend_service_name  = module.cloud_run.backend_service_name
}

