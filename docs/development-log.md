# mctl Development Log

## Product Direction

`mctl` is a Go CLI for managing mihomo subscription links and generating a cleaned local provider file. It is not a full profile manager. The stable mihomo config remains in `~/.config/mihomo/config.yaml`; `mctl` writes app state and generated provider data around it.

Current storage layout:

```text
~/.config/mihomo/
  config.yaml              # user-managed mihomo config
  mctl.json                # mctl subscription state
  providers/
    nodes.yaml             # generated provider payload
```

The provider path intentionally has no extra `mctl/` directory because the local `config.yaml` already references `providers/nodes.yaml`.

## Implemented Commands

Subscription commands live in `internal/cmd/sub.go`:

```text
mctl sub add <name> <url>
mctl sub remove <name-or-url>
mctl sub list
mctl sub update
mctl sub enable <name-or-url>
mctl sub disable <name-or-url>
```

`sub update` loads enabled subscriptions from `mctl.json`, fetches each URL, parses nodes, cleans them, renders provider YAML, and writes `~/.config/mihomo/providers/nodes.yaml` or `$XDG_CONFIG_HOME/mihomo/providers/nodes.yaml`.

## Implemented Packages

`internal/config` owns app state:

- `Load` returns an empty config when `mctl.json` is missing.
- `Save` creates `~/.config/mihomo` and writes indented JSON.
- `AddSub` rejects duplicate names or URLs.
- `RemoveSub` removes by name or URL.
- `SetSubEnabled` backs `sub enable` and `sub disable`.

`internal/sub` owns the subscription pipeline:

- `Fetch` downloads subscription bodies with a 30 second HTTP timeout.
- `DecodeBody` accepts base64 subscriptions only when decoded content contains proxy URLs.
- `Parse` supports multiline URI lists and dispatches `anytls://` and `vless://`.
- `Clean` filters ad/status nodes, dedupes by connection identity, and normalizes names like `HK 01` to `Hong Kong 01`.
- `RenderProvider` marshals cleaned proxies to mihomo provider YAML using `go.yaml.in/yaml/v4`.

Supported first-pass protocols:

- `anytls://<password>@<server>:<port>?sni=<sni>&insecure=1#<name>`
- `vless://<uuid>@<server>:<port>?type=<network>&security=<security>&sni=<sni>&insecure=1#<name>`

## Tests And Verification

Current coverage includes:

- config load/save, duplicate detection, remove by name or URL
- `sub update` with temp `XDG_CONFIG_HOME`, `httptest.Server`, enabled/disabled subscriptions, and generated provider path
- clean filtering, dedupe, region numbering, and `Pro` tier naming

Known test gap: `internal/sub/render_test.go` and parser edge-case tests are not present in the current tree.

Recent verification commands that passed during implementation:

```text
go test ./...
go test -race -shuffle=on -count=1 ./...
```

## Decisions

- Use JSON for `mctl` state because it is app-owned and covered by Go stdlib.
- Use YAML only for mihomo-facing provider output and future mihomo config editing.
- Do not add Viper; Cobra is only for CLI command parsing.
- Keep provider generation simple: no profile templating, no UI, no Clash Verge-style profile manager.
- Do not auto-patch `config.yaml` right now because the local mihomo config already points at the generated provider file.

## Next Work

Recommended next steps, in order:

1. Add missing parser/render tests if absent: `DecodeBody`, `ParseURL` for `anytls`, multiline `Parse`, unsupported scheme errors, and vless render fields.
2. Run a real CLI smoke test with temp `XDG_CONFIG_HOME`: `sub add`, `sub list`, `sub update`, inspect generated `providers/nodes.yaml`.
3. Add mihomo external-controller support for runtime commands:

```text
mctl node list
mctl node use <group> <name>
mctl node test <name>
mctl reload
```

Suggested files for runtime commands:

```text
internal/mihomo/client.go
internal/mihomo/proxy.go
internal/cmd/node.go
internal/cmd/reload.go
```

Minimal controller config can later extend `mctl.json`:

```json
{
  "subs": [],
  "controller": {
    "addr": "http://127.0.0.1:9090",
    "secret": ""
  }
}
```

External-controller endpoints to use:

- `GET /proxies`
- `PUT /proxies/{group}` with `{"name":"node-name"}`
- `GET /proxies/{name}/delay?url=https://www.gstatic.com/generate_204&timeout=5000`
- `PUT /configs?force=true`
- `Authorization: Bearer <secret>` when a secret is configured
