# Project conventions

- Each service keeps its domain, model, repository, service, error, metrics, and gRPC implementation files together in one Go package named after the service. For example, `apps/user/user/{domain,model,repository,service,errors,metrics,grpc}.go` all use `package user`.
- Keep the executable entry point in `apps/<service>/main.go`. Keep configuration in its own directory and generate shared protocol code in `pkg/proto`.
- Do not put Go-specific `go_package` options in `.proto` files. Configure generated Go package paths centrally in `buf.gen.yaml`.
- Do not split these implementation files into layer directories such as `service/`, `repository/`, or `domain/`.
- Always pin dependency and tool versions to their complete release version everywhere, including Docker Compose, GitHub Actions, Go modules, and Makefile commands. Avoid major-only, minor-only, `latest`, and floating tags or refs.
- Keep configuration minimal. Use the default values supplied by a service or tool unless the task explicitly requires changing them. For PostgreSQL in Docker Compose, set only the required `POSTGRES_PASSWORD: postgres`; use the default `postgres` user and database.
