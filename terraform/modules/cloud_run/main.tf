terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0, < 6.0"
    }
  }
}

resource "google_cloud_run_service" "backend" {
  project  = var.project_id
  name     = "${var.project_name}-backend-${var.environment}"
  location = var.region

  template {
    spec {
      containers {
        image = var.backend_image_url
        env {
          name  = "STORAGE"
          value = "postgres"
        }
        env {
          name  = "DB_CONN_STR"
          value = "postgres://${var.db_user}:${var.db_password}@localhost:5432/${var.db_name}?sslmode=disable"
        }
        env {
          name  = "AUTH"
          value = var.auth_type
        }
        env {
          name  = "GOOGLE_CLIENT_ID"
          value = var.google_client_id
        }
        env {
          name  = "GOOGLE_CLIENT_SECRET"
          value = var.google_client_secret
        }
        env {
          name  = "GOOGLE_REDIRECT_URL"
          value = "https://${var.domain}/auth/google/callback"
        }
        env {
          name  = "UI_URL"
          value = "https://${var.domain}"
        }
        env {
          name  = "AUTH_SECRET"
          value = var.auth_secret
        }
      }
    }

    metadata {
      annotations = {
        "run.googleapis.com/cloudsql-instances" = var.db_connection_name
        # Allow requests to scale down to 0
        "autoscaling.knative.dev/minScale" = "0"
        "autoscaling.knative.dev/maxScale" = "3"
      }
    }
  }

  autogenerate_revision_name = true
}

# Allow unauthenticated invocation for backend
data "google_iam_policy" "noauth_backend" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth_backend_policy" {
  location    = google_cloud_run_service.backend.location
  project     = google_cloud_run_service.backend.project
  service     = google_cloud_run_service.backend.name
  policy_data = data.google_iam_policy.noauth_backend.policy_data
}

resource "google_cloud_run_service" "frontend" {
  project  = var.project_id
  name     = "${var.project_name}-frontend-${var.environment}"
  location = var.region

  template {
    spec {
      containers {
        image = var.frontend_image_url
        ports {
          container_port = 80
        }
        # Frontend is built via Vite, its environment vars are needed built-in, 
        # but if we run a custom Go server or Nginx container, we can pass them here.
        # Assuming Vite outputs static files served by Nginx.
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale" = "0"
        "autoscaling.knative.dev/maxScale" = "3"
      }
    }
  }

  autogenerate_revision_name = true
}

resource "google_cloud_run_service_iam_policy" "noauth_frontend_policy" {
  location    = google_cloud_run_service.frontend.location
  project     = google_cloud_run_service.frontend.project
  service     = google_cloud_run_service.frontend.name
  policy_data = data.google_iam_policy.noauth_backend.policy_data
}
