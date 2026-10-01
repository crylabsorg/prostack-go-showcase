@prostack-go/AGENTS.md

# Project conventions

- Follow YAGNI: choose the smallest solution that meets the current requirement. Do not add public APIs, configuration, or abstractions without a current caller or concrete need.

- Each service keeps its domain, model, repository, service, error, metrics, and transport controller files together in one Go package named after the service. For example, `apps/user/user/{domain,model,repository,service,errors,metrics,grpc}.go` all use `package user`. Keep NATS message handling in `nats.go`, Kafka message handling in `kafka.go`, and business logic in `service.go` without broker dependencies. Keep HTTP gateway request and response DTOs and their conversions in `dto.go`.
- Keep methods that convert a domain type to another representation in `domain.go` beside that type. For example, define `User.ToProto()` beside `User`.
- Limit API request bodies in Traefik, not Go handlers.
- Define service errors as `ps_error.PSError` with explicit error code, HTTP status, and gRPC status. Return them directly from gRPC handlers; `ps_grpc.ErrorInterceptor` converts them and handles unexpected errors. In HTTP gateways, use `controller.WriteError` to read the status from the error details instead of maintaining a separate status mapping.
- Use `edition = "2024";` in `.proto` files.
- Always pin dependency and tool versions to their complete release version everywhere, including Docker Compose, GitHub Actions, Go modules, and Makefile commands. Avoid major-only, minor-only, `latest`, and floating tags or refs.
- Keep configuration minimal. Use the default values supplied by a service or tool unless the task explicitly requires changing them. For PostgreSQL in Docker Compose, set only the required `POSTGRES_PASSWORD: postgres`; use the default `postgres` user and database.
- Do not silently discard errors from fallible operations (for example, `_ = msg.Term()` or `_ = msg.NakWithDelay(...)`). Return or log the error with enough context to identify the failed operation.
