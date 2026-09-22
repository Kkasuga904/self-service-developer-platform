locals {
  tags = {
    Project     = "self-service-developer-platform"
    Environment = "dev"
    ManagedBy   = "terraform"
    CostScope   = "disposable-validation"
  }
}

module "network" {
  source = "./modules/network"

  name               = var.name
  availability_zones = ["${var.aws_region}a", "${var.aws_region}c"]
}

module "eks" {
  source = "./modules/eks"

  name                = var.name
  kubernetes_version  = var.kubernetes_version
  private_subnet_ids  = module.network.private_subnet_ids
  vpc_id              = module.network.vpc_id
  node_instance_types = var.node_instance_types
  public_access_cidrs = var.cluster_public_access_cidrs
}

module "github_oidc" {
  count  = var.github_repository == "" ? 0 : 1
  source = "./modules/github-oidc"

  name                              = var.name
  github_repository                 = var.github_repository
  github_oidc_subject               = var.github_oidc_subject
  eks_cluster_arn                   = module.eks.cluster_arn
  existing_github_oidc_provider_arn = var.existing_github_oidc_provider_arn
}

resource "aws_ecr_repository" "sample_app" {
  name                 = "${var.name}/payment-api"
  image_tag_mutability = "IMMUTABLE"
  force_delete         = true

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }
}

resource "aws_ecr_lifecycle_policy" "sample_app" {
  repository = aws_ecr_repository.sample_app.name
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep the ten most recent validation images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 10
      }
      action = { type = "expire" }
    }]
  })
}
