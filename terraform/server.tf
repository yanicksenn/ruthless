resource "google_artifact_registry_repository" "repo" {
  location      = var.region
  repository_id = "ruthless-repo"
  description   = "Docker repository for Ruthless CAH Clone"
  format        = "DOCKER"
}

resource "google_cloud_run_v2_service" "api" {
  name     = "ruthless-api"
  location = var.region

  template {
    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.repo.repository_id}/ruthless-api:${var.image_tag}"
      
      env {
        name  = "RUTHLESS_STORAGE"
        value = "postgres"
      }
      
      env {
        name  = "RUTHLESS_AUTH"
        value = "oauth"
      }

      # DB Connection string
      env {
        name  = "DATABASE_URL"
        value = "postgres://${google_sql_user.db_user.name}:${var.db_password}@${google_sql_database_instance.postgres.public_ip_address}:5432/${google_sql_database.database.name}"
      }
    }
  }
}

resource "google_cloud_run_service_iam_member" "public_access" {
  location = google_cloud_run_v2_service.api.location
  project  = google_cloud_run_v2_service.api.project
  service  = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
