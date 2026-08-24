# RailVolt

RailVolt coordinates traction-power maintenance permits, isolation sequences, grounding evidence, insulation certificates and staged re-energization.

Run locally with Go 1.26.2:

```text
go run ./cmd/railvolt -listen 127.0.0.1:19699 -data data -web web
```

The service exposes four operational pages at `/permits`, `/isolation`, `/energize` and `/events`. Health status is available at `/healthz`. The repository includes vendored dependencies and can be built offline with `go build -mod=vendor ./...`.
