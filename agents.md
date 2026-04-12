# Umami MCP Server — Agent Guidelines

## Go Style

- Follow [Google's Go style guide](https://google.github.io/styleguide/go).
- Prefer the standard library. Only reach for third-party packages when the stdlib genuinely cannot do the job.
- No comments unless an LLM cannot understand what the code is doing from the code alone. If you need a comment, the code is probably too clever.

## MCP Protocol

- This server implements the [Model Context Protocol](https://modelcontextprotocol.io).
- Current spec version: **2025-11-25** — reference: https://modelcontextprotocol.io/specification/2025-11-25
- Before making protocol-level changes (transports, JSON-RPC handling, capability negotiation), check the latest spec to ensure compliance.

## Before Committing

- Run tests: `go test ./...`
- Run the linter **matching CI** (v1 config): `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run ./...`
- Fix any lint or test failures before pushing. Do not skip or suppress linter rules without justification.

## Project Layout

- `umami.go` — Umami API client
- `mcp.go` — MCP server, JSON-RPC dispatch, tool/prompt definitions
- `handlers.go` — Tool handler implementations
- `http.go` — Streamable HTTP transport
- `config.go` — Config loading (YAML + env vars)
- `tools.json` / `prompts.json` — Tool and prompt schemas (embedded at build)
- `main.go` — Entrypoint, transport selection
