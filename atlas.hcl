variable "app" {
  type = string
}

env "local" {
  url = yamldecode(file("apps/${var.app}/config.yaml")).DATABASE_URL
  src = "file://apps/${var.app}/schema.sql"
  dev = "docker://postgres/17.11/dev?search_path=public"
  migration {
    dir = "file://apps/${var.app}/migrations"
  }
}
