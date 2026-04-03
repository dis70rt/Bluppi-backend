# Contributing to Bluppi Backend

Contributions are welcome. This document outlines how to get involved.

## Getting Started

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run tests: `go test ./internals/... ./cmd/...`
5. Open a pull request

## Code Style

This project follows standard Go conventions. Run `go vet` and `go fmt` before committing.

Each feature module lives under `internals/` and follows a consistent layout:

```
internals/<feature>/
    handler.go       # gRPC handler
    service.go       # Business logic
    repository.go    # Data access (Postgres)
    model.go         # Domain types
    mapper.go        # Proto to domain mapping
```

Some modules include additional files like `memgraph_repo.go`, `redis_repository.go`, `consumer.go`, or `reaper.go` depending on their requirements.

## Pull Requests

Keep PRs focused on a single change. Include a clear description of what the change does and why.

## Reporting Issues

Open a GitHub issue with:

1. A clear title
2. Steps to reproduce (if applicable)
3. Expected vs actual behavior

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
