---
name: Go HTTP service gold standard
description: Grafana blog article by Mat Ryer — the user's reference for how to write HTTP services in Go
type: reference
---

URL: https://grafana.com/blog/how-i-write-http-services-in-go-after-13-years/

Key patterns the user wants followed:
- **run() function** — logic lives in run(ctx, w, args), not main()
- **NewServer constructor** — accepts all dependencies as arguments, returns http.Handler
- **Handler maker functions** — handlers are functions returning http.HandlerFunc, not structs
- **routes.go** — all routes in one file
- **End-to-end tests** — call run() from tests, hit real HTTP endpoints; avoid unit testing individual handlers with mocks
- **Inline request/response types** — defined inside handler functions
- **Generic encode/decode helpers** — centralized JSON serialization
- **sync.Once** for deferred expensive setup
