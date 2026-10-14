# In-memory demo services

This app hosts two gRPC services in one process: organizations and projects.
They store resources in memory with generated UUIDs. Names are trimmed and must
be nonempty. Projects require an existing organization and can only be fetched
through that organization's URL. All data is lost on restart.

The app needs no database or message broker. Its gRPC listener is
`localhost:43003`; health and metrics use `localhost:33003`.

## Run the example

From the repository root, in separate terminals:

```sh
make run-demo
```

```sh
OTEL_TRACES_EXPORTER=none make run-gateway
```

The existing user app is needed only when calling `/v1/users`. The demo endpoints
and Swagger UI work with just demo and gateway running.

Open http://localhost:33000/swagger/ to see users, organizations, and projects
in **one Swagger UI**, backed by **one OpenAPI specification**.

| Method | Gateway path | Behavior |
| --- | --- | --- |
| POST | `/v1/organizations` | Create an organization with `{"name":"A"}` |
| GET | `/v1/organizations/{orgId}` | Fetch an organization |
| POST | `/v1/organizations/{orgId}/projects` | Create a project with `{"name":"TEST"}` |
| GET | `/v1/organizations/{orgId}/projects/{projectId}` | Fetch that organization's project |

Create an organization, copy its `id`, then create a project under it:

```sh
curl -i http://localhost:33000/v1/organizations \
  -H 'Content-Type: application/json' -d '{"name":"A"}'

curl -i http://localhost:33000/v1/organizations/ORG_ID/projects \
  -H 'Content-Type: application/json' -d '{"name":"TEST"}'

curl -i http://localhost:33000/v1/organizations/ORG_ID/projects/PROJECT_ID
```

## Follow the code

```text
apps/gateway/swagger.json                    one HTTP contract
  + oapi/organization.yaml (organizations) -> gateway/organization/api.gen.go
  + oapi/project.yaml      (projects)      -> gateway/project/api.gen.go
  + oapi/user.yaml         (users)         -> gateway/user/api.gen.go

gateway/organization/http.go -> OrganizationServiceClient -> demo/organization/grpc.go
gateway/project/http.go      -> ProjectServiceClient      -> demo/project/grpc.go
```

`apps/gateway/main.go` connects both demo clients to one gRPC connection and
registers their generated routes on the same HTTP mux. The nested project URL
belongs to the `project` package. Each package has its own generated
`ServerInterface`; the user handler does not implement demo operations.

In demo, `service.go` owns the in-memory behavior and `domain.go` converts domain
objects to protobuf. gRPC handlers return `ps_error.PSError` directly;
`ps_grpc.ErrorInterceptor` supplies the details used by gateway's
`controller.WriteError` to produce HTTP 400/404 responses.

Regenerate both contracts after edits:

```sh
make proto-gen
make build-gateway-swagger
```
