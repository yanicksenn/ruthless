output "backend_url" {
  value       = google_cloud_run_service.backend.status[0].url
  description = "URL of the backend Cloud Run service."
}

output "frontend_url" {
  value       = google_cloud_run_service.frontend.status[0].url
  description = "URL of the frontend Cloud Run service."
}

output "backend_service_name" {
  value       = google_cloud_run_service.backend.name
  description = "Name of the backend Cloud Run service."
}

output "frontend_service_name" {
  value       = google_cloud_run_service.frontend.name
  description = "Name of the frontend Cloud Run service."
}
