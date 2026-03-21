terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.0, < 6.0"
    }
  }
}

resource "google_compute_region_network_endpoint_group" "frontend_neg" {
  project               = var.project_id
  name                  = "${var.project_name}-fe-neg-${var.environment}"
  network_endpoint_type = "SERVERLESS"
  region                = var.region
  cloud_run {
    service = var.frontend_service_name
  }
}

resource "google_compute_region_network_endpoint_group" "backend_neg" {
  project               = var.project_id
  name                  = "${var.project_name}-be-neg-${var.environment}"
  network_endpoint_type = "SERVERLESS"
  region                = var.region
  cloud_run {
    service = var.backend_service_name
  }
}

resource "google_compute_backend_service" "frontend_backend" {
  project               = var.project_id
  name                  = "${var.project_name}-fe-be-${var.environment}"
  protocol              = "HTTP"
  port_name             = "http"
  load_balancing_scheme = "EXTERNAL"
  backend {
    group = google_compute_region_network_endpoint_group.frontend_neg.id
  }
}

resource "google_compute_backend_service" "backend_backend" {
  project               = var.project_id
  name                  = "${var.project_name}-be-be-${var.environment}"
  protocol              = "HTTP"
  port_name             = "http"
  load_balancing_scheme = "EXTERNAL"
  backend {
    group = google_compute_region_network_endpoint_group.backend_neg.id
  }
}

resource "google_compute_url_map" "url_map" {
  project         = var.project_id
  name            = "${var.project_name}-url-map-${var.environment}"
  default_service = google_compute_backend_service.frontend_backend.id

  host_rule {
    hosts        = [var.domain]
    path_matcher = "allpaths"
  }

  path_matcher {
    name            = "allpaths"
    default_service = google_compute_backend_service.frontend_backend.id

    path_rule {
      paths   = [
        "/auth/*",
        "/ruthless.v1.CardService/*",
        "/ruthless.v1.DeckService/*",
        "/ruthless.v1.SessionService/*",
        "/ruthless.v1.UserService/*",
        "/ruthless.v1.GameService/*",
        "/ruthless.v1.FriendService/*",
        "/ruthless.v1.NotificationService/*",
        "/ruthless.v1.SessionInvitationService/*"
      ]
      service = google_compute_backend_service.backend_backend.id
    }
  }
}

resource "google_compute_managed_ssl_certificate" "cert" {
  project = var.project_id
  name    = "${var.project_name}-cert-${var.environment}"
  managed {
    domains = [var.domain]
  }
}

resource "google_compute_target_https_proxy" "https_proxy" {
  project          = var.project_id
  name             = "${var.project_name}-https-proxy-${var.environment}"
  url_map          = google_compute_url_map.url_map.id
  ssl_certificates = [google_compute_managed_ssl_certificate.cert.id]
}

resource "google_compute_global_forwarding_rule" "https_forwarding_rule" {
  project               = var.project_id
  name                  = "${var.project_name}-https-rule-${var.environment}"
  target                = google_compute_target_https_proxy.https_proxy.id
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL"
}

# Optional HTTP to HTTPS redirect
resource "google_compute_url_map" "http_redirect" {
  project = var.project_id
  name    = "${var.project_name}-http-redirect-${var.environment}"

  default_url_redirect {
    https_redirect = true
    strip_query    = false
  }
}

resource "google_compute_target_http_proxy" "http_proxy" {
  project = var.project_id
  name    = "${var.project_name}-http-proxy-${var.environment}"
  url_map = google_compute_url_map.http_redirect.id
}

resource "google_compute_global_forwarding_rule" "http_forwarding_rule" {
  project               = var.project_id
  name                  = "${var.project_name}-http-rule-${var.environment}"
  target                = google_compute_target_http_proxy.http_proxy.id
  port_range            = "80"
  load_balancing_scheme = "EXTERNAL"
}
