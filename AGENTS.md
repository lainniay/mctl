# AGENTS.md

## Repo Snapshot

- Go CLI module: `github.com/lainniay/mctl`, entrypoint `cmd/mctl/main.go`.
- Cobra command wiring lives in `internal/cmd`; keep CLI parsing there.
- App-owned state lives in `internal/config` and writes JSON to `$XDG_CONFIG_HOME/mihomo/mctl.json` or `~/.config/mihomo/mctl.json`.
- Subscription fetching/parsing/cleaning/rendering lives in `internal/sub`; do not mix it into Cobra command definitions unless it is orchestration only.
- Project decisions and current roadmap are summarized in `docs/development-log.md`; update that file when direction changes.

## Commands

- Run all tests: `go test ./...`
- Race/shuffle verification: `go test -race -shuffle=on -count=1 ./...`
- Focus a package: `go test ./internal/sub` or `go test ./internal/cmd`
- Run CLI locally: `go run ./cmd/mctl <args>`
- Format touched Go files with `gofmt -w <files>`; there is no Makefile, Taskfile, or linter config yet.

## Current CLI Behavior

- Implemented subscription commands: `sub add`, `sub remove`, `sub list`, `sub update`, `sub enable`, `sub disable`.
- `sub update` fetches only enabled subscriptions, parses them, cleans/dedupes nodes, and writes provider YAML to `$XDG_CONFIG_HOME/mihomo/providers/nodes.yaml` or `~/.config/mihomo/providers/nodes.yaml`.
- Do not reintroduce an extra `mctl/` directory under `providers`; the intended generated provider path is `mihomo/providers/nodes.yaml`.
- `mctl.json` is JSON on purpose; generated mihomo provider output is YAML on purpose.

## Tests And Fixtures

- Tests that touch user config should set `XDG_CONFIG_HOME` with `t.Setenv` and use `t.TempDir`; never write to the real home config during tests.
- `internal/cmd/sub_test.go` uses `httptest.Server` for subscription update coverage and includes a disabled bad URL to prove disabled subs are skipped.
- Current known gap: parser edge-case tests and `internal/sub/render_test.go` are absent; add focused tests before changing parser/render behavior.

## Implementation Notes

- First supported subscription shape is base64 or plaintext multiline URI lists, currently `anytls://` and `vless://`.
- `Clean` filters ad/status node names before deduping, dedupes by `Proxy.Equal` connection identity rather than node name, then generates region-based display names.
- Keep dependencies minimal: Cobra is already used for CLI, `go.yaml.in/yaml/v4` for provider YAML; do not add Viper.
- Future `node` and `reload` commands should talk to mihomo external-controller, not read `providers/nodes.yaml` as runtime state.
