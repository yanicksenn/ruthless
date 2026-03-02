resource "google_sql_database_instance" "postgres" {
  name             = "ruthless-db"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"
    
    ip_configuration {
      ipv4_enabled    = true
      # In a real production setup, we'd use private IP
    }
  }
  
  deletion_protection = false # For easier teardown in dev
}

resource "google_sql_user" "db_user" {
  name     = "ruthless_user"
  instance = google_sql_database_instance.postgres.name
  password = var.db_password
}

resource "google_sql_database" "database" {
  name     = "ruthless_prod"
  instance = google_sql_database_instance.postgres.name
}
