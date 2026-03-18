

output "artifact_registry_url" {
  value = module.artifact_registry.repository_url
}

output "database_instance_connection_name" {
  value = module.database.instance_connection_name
}

output "load_balancer_ip" {
  value       = module.load_balancer.ip_address
  description = "The public IP address of the Load Balancer. Point your domain A record here."
}
