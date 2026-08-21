# AGENTS.md

## Repo Snapshot

- Go CLI module: `github.com/lainniay/mctl`, entrypoint `cmd/mctl/main.go`.
- Cobra command wiring lives in `internal/cmd`; keep CLI parsing there.
- App paths live in `internal/config`: editable configuration is under `$XDG_CONFIG_HOME/mihomo`, while generated files and runtime state are under `$XDG_STATE_HOME/mihomo`.
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
- `sub update` fetches only enabled subscriptions, parses them, cleans/dedupes nodes, and writes provider YAML to `$XDG_STATE_HOME/mihomo/providers/nodes.yaml` or `~/.local/state/mihomo/providers/nodes.yaml`.
- Do not reintroduce an extra `mctl/` directory under `providers`; the intended generated provider path is `mihomo/providers/nodes.yaml`.
- `base.yaml` and `mctl.json` belong in the config directory; generated `config.yaml`, providers, subscriptions, and selected-group state belong in the state directory.
- `mctl.json` is JSON on purpose; generated mihomo config and provider output are YAML on purpose.

## Tests And Fixtures

- Tests that touch user config or state should set the corresponding `XDG_CONFIG_HOME` and `XDG_STATE_HOME` with `t.Setenv` and use `t.TempDir`; never write to real user directories during tests.
- `internal/cmd/sub_test.go` uses `httptest.Server` for subscription update coverage and includes a disabled bad URL to prove disabled subs are skipped.
- Current known gap: parser edge-case tests and `internal/sub/render_test.go` are absent; add focused tests before changing parser/render behavior.

## Implementation Notes

- First supported subscription shape is base64 or plaintext multiline URI lists, currently `anytls://` and `vless://`.
- `Clean` filters ad/status node names before deduping, dedupes by `Proxy.Equal` connection identity rather than node name, then generates region-based display names.
- Keep dependencies minimal: Cobra is already used for CLI, `go.yaml.in/yaml/v4` for provider YAML; do not add Viper.
- Future `node` and `reload` commands should talk to mihomo external-controller, not read `providers/nodes.yaml` as runtime state.
