# Contributing to qrafter

Thanks for taking the time to improve qrafter.

Qrafter is a typed SQL query builder for Go.

## Ways to help

- Improve examples for common query patterns
- Add or refine dialect support
- Expand render and integration tests
- Improve package documentation and README clarity
- Report API rough edges from real usage

## Development

Run the root package checks:

```sh
go test ./...
```

Run the integration and compatibility tests:

```sh
cd tests
go test ./...
```

Run the linter if you have `golangci-lint` installed:

```sh
golangci-lint run ./...
```

## Pull requests

- Keep changes focused and easy to review.
- Add tests for behavior changes.
- Add examples when introducing user-facing APIs.
- Preserve existing query rendering behavior unless the change intentionally
  fixes a bug.
- Mention any dialect-specific behavior in the PR description.

For larger API changes, open an issue first so the design can be discussed
before implementation.
