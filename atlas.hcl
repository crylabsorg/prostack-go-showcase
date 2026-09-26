variable "app" {
  type = string
}

env "local" {
  url = yamldecode(file("apps/${var.app}/config.yaml")).DATABASE_URL
  migration {
    dir = "file://apps/${var.app}/migrations"
  }
}
