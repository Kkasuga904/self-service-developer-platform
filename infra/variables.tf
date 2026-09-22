variable "aws_region" {
  description = "AWS Region for the disposable validation environment."
  type        = string
  default     = "ap-northeast-1"
}

variable "name" {
  description = "Name prefix for platform resources."
  type        = string
  default     = "developer-platform-dev"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{2,30}$", var.name))
    error_message = "name must be a lowercase, DNS-like value between 3 and 31 characters."
  }
}

variable "kubernetes_version" {
  description = "EKS Kubernetes version. Confirm standard support immediately before apply."
  type        = string
  default     = "1.35"
}

variable "node_instance_types" {
  description = "EC2 instance types for the small managed node group."
  type        = list(string)
  default     = ["t3.medium"]
}

variable "cluster_public_access_cidrs" {
  description = "Trusted operator CIDRs allowed to reach the public EKS API endpoint. 0.0.0.0/0 is rejected."
  type        = list(string)

  validation {
    condition = length(var.cluster_public_access_cidrs) > 0 && alltrue([
      for cidr in var.cluster_public_access_cidrs : cidr != "0.0.0.0/0" && can(cidrnetmask(cidr))
    ])
    error_message = "Provide at least one valid trusted CIDR; 0.0.0.0/0 is not allowed."
  }
}

variable "github_repository" {
  description = "GitHub repository in owner/name form allowed to assume the validation role; empty disables GitHub OIDC resources."
  type        = string
  default     = ""

  validation {
    condition     = var.github_repository == "" || can(regex("^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$", var.github_repository))
    error_message = "github_repository must be empty or use owner/name format."
  }
}

variable "existing_github_oidc_provider_arn" {
  description = "Existing account-level GitHub OIDC provider ARN. Leave empty to create one when github_repository is set."
  type        = string
  default     = ""
}

variable "github_oidc_subject" {
  description = "Optional exact GitHub OIDC sub claim for accounts using a customized subject template. Empty uses the standard repository/main subject."
  type        = string
  default     = ""
}
