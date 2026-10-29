# Development log

## Direct VLESS sources and naming

- `MCTL_SUB_URL=... mctl sub add <name>` now accepts `vless://` node URIs in addition to HTTP(S) subscriptions and direct `ss://` nodes. Direct VLESS sources are validated when added, persisted in `subscriptions.json`, and parsed locally by `sub update` without an HTTP fetch.
- Region-mapped VLESS nodes retain their protocol in the generated display name, for example `JP-VLESS` becomes `Japan-01-VLESS`. Connection deduplication still uses the parsed VLESS settings rather than the display name.
- The currently rendered VLESS transport settings cover TCP/WebSocket, TLS, SNI, WebSocket path/Host, certificate verification, the URI `fp` client fingerprint (rendered as mihomo `client-fingerprint`), REALITY (`security=reality` with required `pbk` and `sni`, optional `sid`, rendered as `reality-opts`/`tls: true`), and the `flow=xtls-rprx-vision` XTLS flow (TCP only). `flow`, `pbk` and `sid` participate in deduplication. Other VLESS transports (gRPC, HTTP/2, xhttp, …) remain outside the current support boundary.

## SS sources and Shadowsocks support

- Keep the existing `MCTL_SUB_URL=... mctl sub add <name>` interface. HTTP(S) URLs remain remote subscriptions; `ss://` URIs are validated at add time and persisted as direct sources in the existing state-directory `subscriptions.json` (0600). No import command, temporary file, or separate registry is needed.
- `sub update` parses enabled direct SS sources without fetching them, merges them with remote subscriptions, and applies cleaning/deduplication/region naming. Untagged SS nodes use `host:port`; colliding display names receive deterministic numeric suffixes after deduplication, without clashing with existing suffixed names. Enable, disable, list and remove work unchanged. HTTP subscription bodies may also contain SS URI lists, either plaintext or base64-encoded.
- `Proxy.SS` models cipher, UDP, plugin/options and UDP-over-TCP/version as comparable values. All SS settings participate in connection identity; display names do not. Password remains in the shared `Proxy.Password` field.
- Supported URI forms: SIP002 percent-encoded credentials; padded/unpadded standard or URL-safe Base64 userinfo for non-2022 ciphers; legacy whole-link Base64 for non-2022 ciphers; bracketed IPv6; percent-encoded tags and passwords. SS2022 requires plain percent-encoded userinfo and validates decoded key lengths (AES-128: 16 bytes; AES-256/ChaCha20: 32 bytes). AES identity-key chains are preserved.
- SS plugin URI support covers `obfs-local`/`simple-obfs`/`obfs` (mapped to mihomo `obfs`, with `obfs=http|tls` and optional `obfs-host`) and `v2ray-plugin` websocket mode (`mode`, `host`, `path`, `tls`, `mux`, `skip-cert-verify`). Reference plugin defaults are materialized (obfs host `cloudfront.net`; v2ray-plugin host `cloudfront.com`, path `/`, mux enabled), rather than inheriting different mihomo defaults. Unknown plugins/options, duplicate parameters and invalid values fail explicitly rather than silently losing settings. Other mihomo plugins (e.g. shadow-tls/restls), arbitrary YAML options and SIP003 escape sequences are not yet supported.
- Local URI query extensions: `udp=true|false|1|0` (default true), `udp-over-tcp=true|false|1|0`, `udp-over-tcp-version=1|2` (requires UDP-over-TCP enabled; defaults to 2 when enabled). Unknown top-level query parameters are deliberately rejected as well; this is stricter than SIP002's recommendation to ignore unknown extensions, so links containing extra application metadata must have it removed first. Ordinary cipher names are preserved; actual availability is validated by the installed mihomo through `config validate`/`config apply`, avoiding a stale cipher allowlist in mctl.
- Generated provider files now use 0600 with a 0700 provider directory because they contain node credentials. The mihomo process must run as the same user (or otherwise have appropriate access).
- Validation errors for SS do not include URI/password values. `sub list` still prints names and enabled status only. Do not commit real node credentials in examples or fixtures.

### Usage

```sh
# Copy an ss:// link first; this avoids putting it literally in shell history.
MCTL_SUB_URL="$(pbpaste)" mctl sub add personal
mctl sub update
mctl config apply
```

No clipboard or temporary environment variable is needed after adding the source; the link itself remains persisted until the source is removed. Do not manually edit generated `providers/nodes.yaml`, which is overwritten on update. SS nodes use the same region naming rules and retain a protocol suffix (e.g. `JP-SS2022` becomes `Japan-01-SS`).
