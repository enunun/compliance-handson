terraform {
  required_version = "~> 1.13"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

# 接続先と認証情報は環境変数(AWS_ENDPOINT_URLなど)から読む．
provider "aws" {
  # MiniStackはパス形式のS3にだけ対応する．
  s3_use_path_style = true
}
