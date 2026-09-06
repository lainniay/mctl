package sub

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ShadowsocksOptions contains SS-specific connection settings. Keep it comparable:
// every setting participates in connection identity, not just cipher/password.
type ShadowsocksOptions struct {
	Cipher            string
	UDP               bool
	Plugin            string
	PluginOptions     SSPluginOptions
	UDPOverTCP        bool
	UDPOverTCPVersion int
}

type SSPluginOptions struct {
	Mode           string
	Host           string
	Path           string
	TLS            bool
	Mux            bool
	SkipCertVerify bool
}

func decodeSSBase64(value string) (string, error) {
	for _, encoding := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.StdEncoding} {
		if data, err := encoding.Strict().DecodeString(value); err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("invalid SS base64 encoding")
}

func parseSS(raw string) (Proxy, error) {
	// Legacy links encode the entire method:password@host:port, excluding tag.
	rest, tag, tagged := strings.Cut(strings.TrimPrefix(raw, "ss://"), "#")
	encoded := false
	if !strings.Contains(rest, "@") {
		decoded, err := decodeSSBase64(rest)
		if err != nil {
			return Proxy{}, err
		}
		credentials, endpoint, ok := strings.Cut(decoded, "@")
		// Passwords may contain @; the final @ separates the endpoint.
		if i := strings.LastIndex(decoded, "@"); i >= 0 {
			credentials, endpoint, ok = decoded[:i], decoded[i+1:], true
		}
		cipher, password, found := strings.Cut(credentials, ":")
		if !ok || !found {
			return Proxy{}, fmt.Errorf("invalid legacy SS URI")
		}
		raw = "ss://" + url.UserPassword(cipher, password).String() + "@" + endpoint
		if tagged {
			raw += "#" + tag
		}
		encoded = true
	} else if userinfo, endpoint, _ := strings.Cut(rest, "@"); !strings.Contains(userinfo, ":") {
		userinfo, err := url.PathUnescape(userinfo)
		if err != nil {
			return Proxy{}, fmt.Errorf("invalid SS userinfo")
		}
		credentials, err := decodeSSBase64(userinfo)
		if err != nil {
			return Proxy{}, err
		}
		cipher, password, ok := strings.Cut(credentials, ":")
		if !ok {
			return Proxy{}, fmt.Errorf("SS credentials require cipher and password")
		}
		raw = "ss://" + url.UserPassword(cipher, password).String() + "@" + endpoint
		if tagged {
			raw += "#" + tag
		}
		encoded = true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid SS URI")
	}
	if u.Scheme != "ss" || u.Opaque != "" || u.User == nil || u.Hostname() == "" || (u.Path != "" && u.Path != "/") {
		return Proxy{}, fmt.Errorf("SS URI requires credentials and server:port, with no path")
	}
	if _, _, err := net.SplitHostPort(u.Host); err != nil {
		return Proxy{}, fmt.Errorf("invalid SS server:port (IPv6 must be bracketed)")
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 {
		return Proxy{}, fmt.Errorf("invalid SS port")
	}
	cipher := u.User.Username()
	password, plain := u.User.Password()
	if !plain {
		credentials, err := decodeSSBase64(cipher)
		if err != nil {
			return Proxy{}, err
		}
		var ok bool
		cipher, password, ok = strings.Cut(credentials, ":")
		if !ok {
			return Proxy{}, fmt.Errorf("SS credentials require cipher and password")
		}
		encoded = true
	}
	if cipher == "" || password == "" {
		return Proxy{}, fmt.Errorf("SS cipher and password are required")
	}
	if err := validateSSCipher(cipher, password, encoded); err != nil {
		return Proxy{}, err
	}
	options := ShadowsocksOptions{Cipher: cipher, UDP: true}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid SS query")
	}
	for key, values := range query {
		if len(values) != 1 {
			return Proxy{}, fmt.Errorf("duplicate SS query parameter")
		}
		switch key {
		case "plugin":
			if err := parseSSPlugin(values[0], &options); err != nil {
				return Proxy{}, err
			}
		case "udp", "udp-over-tcp":
			value, err := parseSSBool(values[0])
			if err != nil {
				return Proxy{}, err
			}
			if key == "udp" {
				options.UDP = value
			} else {
				options.UDPOverTCP = value
			}
		case "udp-over-tcp-version":
			version, err := strconv.Atoi(values[0])
			if err != nil || (version != 1 && version != 2) {
				return Proxy{}, fmt.Errorf("SS UDP-over-TCP version must be 1 or 2")
			}
			options.UDPOverTCPVersion = version
		default:
			return Proxy{}, fmt.Errorf("unsupported SS query parameter")
		}
	}
	if options.UDPOverTCPVersion != 0 && !options.UDPOverTCP {
		return Proxy{}, fmt.Errorf("SS UDP-over-TCP version requires udp-over-tcp=true")
	}
	if options.UDPOverTCP && options.UDPOverTCPVersion == 0 {
		options.UDPOverTCPVersion = 2
	}
	name := u.Fragment
	if name == "" {
		name = net.JoinHostPort(u.Hostname(), strconv.Itoa(port))
	}
	return Proxy{Name: name, Type: "ss", Server: u.Hostname(), Port: port, Password: password, SS: options}, nil
}

func validateSSCipher(cipher, password string, encoded bool) error {
	size := 0
	switch cipher {
	case "2022-blake3-aes-128-gcm":
		size = 16
	case "2022-blake3-aes-256-gcm", "2022-blake3-chacha20-poly1305":
		size = 32
	default:
		if strings.HasPrefix(cipher, "2022-") {
			return fmt.Errorf("unsupported SS2022 cipher")
		}
		// Mihomo's cipher set evolves independently. Preserve ordinary cipher names;
		// config validate/apply delegates actual cipher availability to the installed core.
		for _, c := range cipher {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return fmt.Errorf("invalid SS cipher name")
			}
		}
		return nil
	}
	if encoded {
		return fmt.Errorf("SS2022 requires percent-encoded plain userinfo, not base64 userinfo")
	}
	keys := strings.Split(password, ":")
	if cipher == "2022-blake3-chacha20-poly1305" && len(keys) != 1 {
		return fmt.Errorf("SS2022 ChaCha20 requires a single key")
	}
	for _, key := range keys {
		decoded, err := base64.StdEncoding.Strict().DecodeString(key)
		if err != nil || len(decoded) != size {
			return fmt.Errorf("SS2022 password keys must be base64-encoded %d-byte keys", size)
		}
	}
	return nil
}

func parseSSBool(value string) (bool, error) {
	switch value {
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("SS boolean must be true, false, 1, or 0")
	}
}

func parseSSPlugin(value string, out *ShadowsocksOptions) error {
	if strings.Contains(value, `\`) {
		return fmt.Errorf("SS plugin escape sequences are not supported")
	}
	parts := strings.Split(value, ";")
	switch parts[0] {
	case "obfs-local", "simple-obfs", "obfs":
		out.Plugin = "obfs"
		out.PluginOptions.Host = "cloudfront.net"
	case "v2ray-plugin":
		out.Plugin = "v2ray-plugin"
		out.PluginOptions.Mode = "websocket"
		out.PluginOptions.Host = "cloudfront.com"
		out.PluginOptions.Path = "/"
		out.PluginOptions.Mux = true
	default:
		return fmt.Errorf("unsupported SS plugin (supported: obfs, v2ray-plugin)")
	}
	seen := map[string]bool{}
	for _, part := range parts[1:] {
		key, value, hasValue := strings.Cut(part, "=")
		if seen[key] {
			return fmt.Errorf("duplicate SS plugin option")
		}
		seen[key] = true
		if out.Plugin == "obfs" {
			if !hasValue || value == "" {
				return fmt.Errorf("invalid SS obfs option")
			}
			switch key {
			case "obfs":
				out.PluginOptions.Mode = value
			case "obfs-host":
				out.PluginOptions.Host = value
			default:
				return fmt.Errorf("unsupported SS obfs option")
			}
			continue
		}
		switch key {
		case "mode", "host", "path":
			if !hasValue || value == "" {
				return fmt.Errorf("missing SS plugin option value")
			}
			switch key {
			case "mode":
				out.PluginOptions.Mode = value
			case "host":
				out.PluginOptions.Host = value
			case "path":
				out.PluginOptions.Path = value
			}
		case "tls", "mux", "skip-cert-verify":
			if !hasValue {
				value = "true"
			}
			b, err := parseSSBool(value)
			if err != nil {
				return err
			}
			switch key {
			case "tls":
				out.PluginOptions.TLS = b
			case "mux":
				out.PluginOptions.Mux = b
			case "skip-cert-verify":
				out.PluginOptions.SkipCertVerify = b
			}
		default:
			return fmt.Errorf("unsupported SS v2ray-plugin option")
		}
	}
	if out.Plugin == "obfs" && out.PluginOptions.Mode != "http" && out.PluginOptions.Mode != "tls" {
		return fmt.Errorf("SS obfs mode must be http or tls")
	}
	if out.Plugin == "v2ray-plugin" && out.PluginOptions.Mode != "websocket" {
		return fmt.Errorf("only SS v2ray-plugin websocket mode is supported")
	}
	return nil
}
