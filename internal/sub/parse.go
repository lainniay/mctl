package sub

import (
	"fmt"
	"net/url"
	"strings"
)

type Proxy struct {
	SS                ShadowsocksOptions
	Name              string
	Type              string
	Server            string
	Port              int
	UUID              string
	Password          string
	SNI               string
	Network           string
	Security          string
	ClientFingerprint string
	WSPath            string
	WSHost            string
	Flow              string
	RealityPublicKey  string
	RealityShortID    string
	SkipCertVerify    bool
}

type proxyIdentity struct {
	SS                ShadowsocksOptions
	Type              string
	Server            string
	Port              int
	UUID              string
	Password          string
	SNI               string
	Network           string
	Security          string
	ClientFingerprint string
	WSPath            string
	WSHost            string
	Flow              string
	RealityPublicKey  string
	RealityShortID    string
	SkipCertVerify    bool
}

func (p Proxy) identity() proxyIdentity {
	return proxyIdentity{
		SS:                p.SS,
		Type:              p.Type,
		Server:            p.Server,
		Port:              p.Port,
		UUID:              p.UUID,
		Password:          p.Password,
		SNI:               p.SNI,
		Network:           p.Network,
		Security:          p.Security,
		ClientFingerprint: p.ClientFingerprint,
		WSPath:            p.WSPath,
		WSHost:            p.WSHost,
		Flow:              p.Flow,
		RealityPublicKey:  p.RealityPublicKey,
		RealityShortID:    p.RealityShortID,
		SkipCertVerify:    p.SkipCertVerify,
	}
}

func (p Proxy) Equal(other Proxy) bool {
	return p.identity() == other.identity()
}

// IsDirectNodeURI reports whether raw is a node URI that can be stored as a
// subscription source without being fetched over HTTP.
func IsDirectNodeURI(raw string) bool {
	return strings.HasPrefix(raw, "ss://") || strings.HasPrefix(raw, "vless://")
}

func ParseURL(raw string) (Proxy, error) {
	if strings.HasPrefix(raw, "ss://") {
		return parseSS(raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Proxy{}, err
	}
	switch u.Scheme {
	case "anytls":
		return parseAnyTLS(raw)
	case "vless":
		return parseVLESS(raw)
	default:
		return Proxy{}, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}
}

func Parse(body string) ([]Proxy, error) {
	body = DecodeBody(body)
	var proxies []Proxy
	for line := range strings.Lines(body) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		proxy, err := ParseURL(line)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, proxy)
	}
	return proxies, nil
}

func isTruthy(value string) bool {
	return value == "1" || value == "true"
}
