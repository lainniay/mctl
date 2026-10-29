package sub

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

func parseVLESS(raw string) (Proxy, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid VLESS URI")
	}
	if u.Scheme != "vless" || u.Opaque != "" || u.User == nil || u.User.Username() == "" || u.Hostname() == "" || (u.Path != "" && u.Path != "/") {
		return Proxy{}, fmt.Errorf("VLESS URI requires UUID and server:port, with no path")
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		return Proxy{}, fmt.Errorf("VLESS URI userinfo must contain only a UUID")
	}
	if _, _, err := net.SplitHostPort(u.Host); err != nil {
		return Proxy{}, fmt.Errorf("invalid VLESS server:port (IPv6 must be bracketed)")
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 {
		return Proxy{}, fmt.Errorf("invalid VLESS port")
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return Proxy{}, fmt.Errorf("invalid VLESS query")
	}
	for key, values := range query {
		if len(values) != 1 {
			return Proxy{}, fmt.Errorf("duplicate VLESS query parameter")
		}
		switch key {
		case "type", "security", "sni", "fp", "path", "host", "insecure", "encryption", "headerType", "flow", "pbk", "sid":
		default:
			return Proxy{}, fmt.Errorf("unsupported VLESS query parameter %q", key)
		}
	}
	network := query.Get("type")
	if network != "" && network != "tcp" && network != "ws" {
		return Proxy{}, fmt.Errorf("unsupported VLESS network")
	}
	security := query.Get("security")
	if security != "" && security != "none" && security != "tls" && security != "reality" {
		return Proxy{}, fmt.Errorf("unsupported VLESS security")
	}
	if encryption := query.Get("encryption"); encryption != "" && encryption != "none" {
		return Proxy{}, fmt.Errorf("unsupported VLESS encryption")
	}
	if headerType := query.Get("headerType"); headerType != "" && headerType != "none" {
		return Proxy{}, fmt.Errorf("unsupported VLESS TCP header type")
	}
	fingerprint := query.Get("fp")
	if _, exists := query["fp"]; exists {
		if fingerprint == "" {
			return Proxy{}, fmt.Errorf("VLESS client fingerprint cannot be empty")
		}
		if fingerprint == "none" {
			fingerprint = ""
		} else if !isSupportedVLESSFingerprint(fingerprint) {
			return Proxy{}, fmt.Errorf("unsupported VLESS client fingerprint")
		}
	}
	if fingerprint != "" && security != "tls" && security != "reality" {
		return Proxy{}, fmt.Errorf("VLESS client fingerprint requires security=tls or security=reality")
	}
	flow := query.Get("flow")
	if _, exists := query["flow"]; exists {
		if flow == "" {
			return Proxy{}, fmt.Errorf("VLESS flow cannot be empty")
		}
		if flow != "xtls-rprx-vision" {
			return Proxy{}, fmt.Errorf("unsupported VLESS flow")
		}
	}
	if flow != "" && network == "ws" {
		return Proxy{}, fmt.Errorf("VLESS flow requires type=tcp")
	}
	realityPublicKey, err := parseVLESSRealityPublicKey(query)
	if err != nil {
		return Proxy{}, err
	}
	realityShortID, err := parseVLESSRealityShortID(query)
	if err != nil {
		return Proxy{}, err
	}
	if security == "reality" {
		if realityPublicKey == "" {
			return Proxy{}, fmt.Errorf("VLESS REALITY requires pbk")
		}
		if query.Get("sni") == "" {
			return Proxy{}, fmt.Errorf("VLESS REALITY requires sni")
		}
	}
	if network != "ws" && (query.Get("path") != "" || query.Get("host") != "") {
		return Proxy{}, fmt.Errorf("VLESS path and host require type=ws")
	}
	skipCertVerify := false
	if values, exists := query["insecure"]; exists {
		skipCertVerify, err = parseVLESSBool(values[0])
		if err != nil {
			return Proxy{}, err
		}
	}
	name := u.Fragment
	if name == "" {
		name = net.JoinHostPort(u.Hostname(), strconv.Itoa(port))
	}
	return Proxy{
		Name:              name,
		Type:              "vless",
		Server:            u.Hostname(),
		Port:              port,
		UUID:              u.User.Username(),
		SNI:               query.Get("sni"),
		Network:           network,
		Security:          security,
		ClientFingerprint: fingerprint,
		WSPath:            query.Get("path"),
		WSHost:            query.Get("host"),
		Flow:              flow,
		RealityPublicKey:  realityPublicKey,
		RealityShortID:    realityShortID,
		SkipCertVerify:    skipCertVerify,
	}, nil
}

func parseVLESSRealityPublicKey(query url.Values) (string, error) {
	values, exists := query["pbk"]
	if !exists {
		return "", nil
	}
	if query.Get("security") != "reality" {
		return "", fmt.Errorf("VLESS pbk requires security=reality")
	}
	value := values[0]
	if value == "" {
		return "", fmt.Errorf("VLESS REALITY public key cannot be empty")
	}
	for _, encoding := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.StdEncoding} {
		if decoded, err := encoding.DecodeString(value); err == nil {
			if len(decoded) != 32 {
				return "", fmt.Errorf("VLESS REALITY public key must decode to 32 bytes")
			}
			return value, nil
		}
	}
	return "", fmt.Errorf("VLESS REALITY public key must be base64")
}

func parseVLESSRealityShortID(query url.Values) (string, error) {
	values, exists := query["sid"]
	if !exists {
		return "", nil
	}
	if query.Get("security") != "reality" {
		return "", fmt.Errorf("VLESS sid requires security=reality")
	}
	value := values[0]
	if value == "" {
		return "", nil
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) > 8 {
		return "", fmt.Errorf("VLESS REALITY short ID must be up to 16 hex characters")
	}
	return value, nil
}

func isSupportedVLESSFingerprint(value string) bool {
	switch value {
	case "chrome", "firefox", "safari", "ios", "android", "edge", "360", "qq", "random",
		"chrome120", "firefox120", "safari16", "chrome_psk", "chrome_psk_shuffle",
		"chrome_padding_psk_shuffle", "chrome_pq", "chrome_pq_psk", "randomized":
		return true
	default:
		return false
	}
}

func parseVLESSBool(value string) (bool, error) {
	switch value {
	case "0", "false":
		return false, nil
	case "1", "true":
		return true, nil
	default:
		return false, fmt.Errorf("VLESS boolean must be true, false, 1, or 0")
	}
}
