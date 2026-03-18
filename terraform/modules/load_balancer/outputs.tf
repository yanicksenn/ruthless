output "ip_address" {
  value       = google_compute_global_forwarding_rule.https_forwarding_rule.ip_address
  description = "The public IP address of the Load Balancer."
}
