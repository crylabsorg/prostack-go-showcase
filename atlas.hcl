variable "app" {
  type = string
}

env "local" {
  url = yamldecode(file("apps/${var.app}/config.yaml")).DATABASE_URL
  src = "file://apps/${var.app}/schema.sql"
  dev = "docker://postgres/18.6/dev"
  migration {
    dir = "file://apps/${var.app}/migrations"
  }
}
