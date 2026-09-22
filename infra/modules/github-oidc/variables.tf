variable "name" {
  type = string
}

variable "github_repository" {
  type = string
}

variable "github_oidc_subject" {
  type = string
}

variable "eks_cluster_arn" {
  type = string
}

variable "existing_github_oidc_provider_arn" {
  type = string
}
