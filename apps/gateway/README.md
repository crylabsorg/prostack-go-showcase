# Gateway API documentation

`swagger.json` is the source of truth for the HTTP API. Edit the
specification, then run from the repository root:

```sh
make build-gateway-swagger
```

The target installs oapi-codegen v2.8.0 and runs `oapi/user.yaml`, which selects
user operations with `include-tags` and generates `user/api.gen.go`. Commit
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

`asyncapi.json` describes the user service's actual outgoing events:
`user.created` on NATS JetStream and Kafka. The gateway embeds the specification
and serves it at `/asyncapi.json`, with AsyncAPI UI at `/asyncapi` (also
`/asyncapi/`). Names are optional in the event payload and omitted when absent.
The NATS `Nats-Msg-Id` header and Kafka record key contain the outbox event ID,
not the user ID. Consumers must account for at least once delivery.

The UI is the official
[AsyncAPI React standalone component v3.2.1](https://github.com/asyncapi/asyncapi-react/tree/v3.2.1),
with its JavaScript, CSS, and licenses vendored in
`prostack-go/ps_asyncapi/ui/`. The framework's `ps_asyncapi.SetupAsyncAPI`
serves the specification embedded by the gateway, following `ps_swagger`.
It works
without a CDN or a frontend build. To update it, download the exact npm release
and replace `browser/standalone/index.js`, its `index.js.LICENSE.txt`,
`styles/default.min.css`, and `LICENSE`; keep the local filenames used by
`index.html`. The specification uses
[AsyncAPI 3.0.0](https://www.asyncapi.com/docs/reference/specification/v3.0.0)
and Kafka binding 0.5.0.
