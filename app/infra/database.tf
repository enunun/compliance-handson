resource "random_password" "db" {
  length  = 32
  special = false
}

resource "aws_secretsmanager_secret" "db" {
  name = "app/db-password"
}

resource "aws_secretsmanager_secret_version" "db" {
  secret_id     = aws_secretsmanager_secret.db.id
  secret_string = random_password.db.result
}

resource "aws_db_instance" "app" {
  identifier          = "app"
  engine              = "postgres"
  engine_version      = "15"
  instance_class      = "db.t3.micro"
  allocated_storage   = 20
  db_name             = "app"
  username            = "app"
  password            = random_password.db.result
  skip_final_snapshot = true

  lifecycle {
    # MiniStackは指定しない値にallocated_storageと同じ値を返すので，差分から外す．
    ignore_changes = [max_allocated_storage]
  }
}
