output "instance_connection_name" {
  value       = google_sql_database_instance.instance.connection_name
  description = "The connection name of the Cloud SQL instance to be used by Cloud Run"
}

output "db_user" {
  value = google_sql_user.users.name
}

output "db_password" {
  value     = google_sql_user.users.password
  sensitive = true
}

output "db_name" {
  value = google_sql_database.database.name
}
