output "cluster_name" {
  description = "EKS cluster name used by aws eks update-kubeconfig."
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "EKS API endpoint."
  value       = module.eks.cluster_endpoint
}

output "github_validation_role_arn" {
  description = "Optional GitHub OIDC role for read-only cluster discovery."
  value       = try(module.github_oidc[0].role_arn, null)
}

output "sample_app_repository_url" {
  description = "ECR repository used for the sample payment-api validation image."
  value       = aws_ecr_repository.sample_app.repository_url
}

output "destroy_warning" {
  description = "Reminder that this environment is intentionally disposable."
  value       = "Run make destroy only after reviewing the destroy plan."
}
