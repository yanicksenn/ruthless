terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0, < 6.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0"
    }
  }
}

resource "random_password" "db_password" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "google_sql_database_instance" "instance" {
  project          = var.project_id
  name             = "${var.project_name}-db-${var.environment}"
  region           = var.region
  database_version = "POSTGRES_15"

  settings {
    tier = var.tier

    # We leave backup and other settings as default or configure lightly for cost
    disk_type = "PD_HDD"
    disk_size = 10

    ip_configuration {
      ipv4_enabled = true
      # Remove public IP later and use private IP if in a VPC, but for simplicity public IP + Cloud SQL Auth Proxy is used from Cloud Run
    }
  }

  deletion_protection = var.environment == "prod" ? true : false
}

resource "google_sql_database" "database" {
  project  = var.project_id
  name     = "ruthless_${var.environment}"
  instance = google_sql_database_instance.instance.name
}

resource "google_sql_user" "users" {
  project  = var.project_id
  name     = "ruthless_user"
  instance = google_sql_database_instance.instance.name
  password = random_password.db_password.result
}
