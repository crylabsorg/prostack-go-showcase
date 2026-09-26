@prostack-go/AGENTS.md

# Project conventions

- Each service keeps its domain, model, repository, service, error, metrics, and gRPC implementation files together in one Go package named after the service. For example, `apps/user/user/{domain,model,repository,service,errors,metrics,grpc}.go` all use `package user`.
- Always pin dependency and tool versions to their complete release version everywhere, including Docker Compose, GitHub Actions, Go modules, and Makefile commands. Avoid major-only, minor-only, `latest`, and floating tags or refs.
- Keep configuration minimal. Use the default values supplied by a service or tool unless the task explicitly requires changing them. For PostgreSQL in Docker Compose, set only the required `POSTGRES_PASSWORD: postgres`; use the default `postgres` user and database.
