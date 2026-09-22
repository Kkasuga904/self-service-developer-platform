data "tls_certificate" "github" {
  count = var.existing_github_oidc_provider_arn == "" ? 1 : 0
  url   = "https://token.actions.githubusercontent.com"
}

resource "aws_iam_openid_connect_provider" "github" {
  count = var.existing_github_oidc_provider_arn == "" ? 1 : 0

  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github[0].certificates[0].sha1_fingerprint]
}

locals {
  provider_arn = var.existing_github_oidc_provider_arn != "" ? var.existing_github_oidc_provider_arn : aws_iam_openid_connect_provider.github[0].arn
  oidc_subject = var.github_oidc_subject != "" ? var.github_oidc_subject : "repo:${var.github_repository}:ref:refs/heads/main"
}

resource "aws_iam_role" "github" {
  name = "${var.name}-github-validation"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = local.provider_arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        # Exact match only. The value carries the account-specific custom
        # subject (owner/repository IDs); StringLike without wildcards would
        # behave the same today but invites future wildcard widening, so the
        # operator is pinned to StringEquals and guarded by a CI test.
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" = local.oidc_subject
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "describe_cluster" {
  name = "describe-eks-cluster"
  role = aws_iam_role.github.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["eks:DescribeCluster"]
      Resource = var.eks_cluster_arn
    }]
  })
}
