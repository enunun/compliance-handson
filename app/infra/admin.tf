# APIの最初の管理者のパスワード．
resource "random_password" "admin" {
  length  = 20
  special = false
}

resource "aws_secretsmanager_secret" "admin" {
  name = "app/admin-password"
}

resource "aws_secretsmanager_secret_version" "admin" {
  secret_id     = aws_secretsmanager_secret.admin.id
  secret_string = random_password.admin.result
}
