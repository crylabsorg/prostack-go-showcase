# Gateway OpenAPI

`swagger.json` is the source of truth for the HTTP API. Edit the
specification, then run from the repository root:

```sh
make build-gateway-swagger
```

The target installs oapi-codegen v2.8.0 and runs each config in `oapi/`. Each
config selects one group of operations with `include-tags` and generates that
package's `api.gen.go`: `user`, `organization`, or `project`. Commit
the generated files along with specification changes. The gateway embeds the same
specification and serves it at `/swagger.json`, with Swagger UI at `/swagger/`.

Previously, swaggo generated documentation from Go annotations. Now
[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/tree/v2.8.0) generates
request/response models, `ServerInterface`, route registration, and path parameter
binding from OpenAPI. Each package's `http.go` implements its own interface and
calls gRPC; `dto.go` keeps the conversions to and from protobuf. Assign each
operation to exactly one generation group to avoid duplicate route registration.

The `std-http-server` generator uses the existing `http.ServeMux`. JSON body
decoding stays in the handler with `encoding/json/v2`, and gRPC errors still go
through `controller.WriteError`. The generated router does not validate request
bodies against the schema: malformed JSON and unknown fields are rejected by the
handler, while business validation remains in the backend services.

The [demo walkthrough](../demo/README.md) shows organizations and nested projects
served by a separate app, with A/B/TEST/MOCK resources and one shared Swagger UI.
