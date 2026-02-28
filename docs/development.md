# Development

## Build

```bash
go build ./...

# Build to an explicit output path
go build -o ziti-mcp .
```

## Unit tests

No external dependencies required.

```bash
go test ./internal/...
```

## Integration tests

Integration tests spin up a live [OpenZiti](https://openziti.io) network using `ziti edge quickstart` and run all tools against it. They are skipped automatically if the `ziti` binary is not in PATH.

Install the `ziti` binary from the [OpenZiti releases page](https://github.com/openziti/ziti/releases), then:

```bash
go test -v -timeout 5m ./test/integration/...
```

Run a single test by name:

```bash
go test -v -timeout 5m -run TestFullWorkflow ./test/integration/...
```

## Lint

```bash
golangci-lint run ./...
```

## Releases

Releases are automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions.

- **Tagged releases**: Push a version tag to create a full release with cross-platform binaries:

    ```bash
    git tag v0.2.0
    git push origin v0.2.0
    ```

- **Development builds**: Every push to `main` automatically updates a rolling `dev` pre-release with the latest binaries.
