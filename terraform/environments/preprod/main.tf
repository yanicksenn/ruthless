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

data "external" "repo_hash" {
  program = ["bash", "-c", "echo \"{\\\"hash\\\": \\\"$(git rev-parse --short HEAD)-$(git status --porcelain | shasum | awk '{print $1}' | cut -c1-8)\\\"}\""]
  working_dir = "${path.module}/../../../"
}

resource "null_resource" "build_and_push_images" {
  triggers = {
    image_tag = data.external.repo_hash.result["hash"]
  }

  depends_on = [module.artifact_registry]

  provisioner "local-exec" {
    working_dir = "${path.module}/../../../"
    command = <<EOT
      set -e
      gcloud auth configure-docker ${var.region}-docker.pkg.dev --quiet
      
      bazel run --platforms=@rules_go//go/toolchain:linux_amd64 //backend:tarball
      docker tag backend:latest ${module.artifact_registry.repository_url}/backend:${data.external.repo_hash.result["hash"]}
      docker push ${module.artifact_registry.repository_url}/backend:${data.external.repo_hash.result["hash"]}
      
      export VITE_API_BASE_URL="https://${var.domain}"
      export VITE_GOOGLE_CLIENT_ID="${var.google_client_id}"
      bazel run --define VITE_API_BASE_URL="$VITE_API_BASE_URL" --define VITE_GOOGLE_CLIENT_ID="$VITE_GOOGLE_CLIENT_ID" --platforms=@rules_go//go/toolchain:linux_amd64 //frontend:tarball
      docker tag frontend:latest ${module.artifact_registry.repository_url}/frontend:${data.external.repo_hash.result["hash"]}
      docker push ${module.artifact_registry.repository_url}/frontend:${data.external.repo_hash.result["hash"]}
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
  backend_image_url    = "${module.artifact_registry.repository_url}/backend:${data.external.repo_hash.result["hash"]}"
  frontend_image_url   = "${module.artifact_registry.repository_url}/frontend:${data.external.repo_hash.result["hash"]}"
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

