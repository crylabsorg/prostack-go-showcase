env "local" {
  url = yamldecode(file("apps/user/config.yaml")).DATABASE_URL
  migration {
    dir = "file://apps/user/migrations"
  }
}
