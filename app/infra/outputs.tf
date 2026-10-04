output "database_url" {
  value     = "postgres://${aws_db_instance.app.username}:${random_password.db.result}@${aws_db_instance.app.endpoint}/${aws_db_instance.app.db_name}"
  sensitive = true
}
