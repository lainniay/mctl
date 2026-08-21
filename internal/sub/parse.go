package sub

import (
	"fmt"
	"net/url"
	"strings"
)

type Proxy struct {
	Name           string
	Type           string
	Server         string
	Port           int
	UUID           string
	Password       string
	SNI            string
	Network        string
	Security       string
	SkipCertVerify bool
}

type proxyIdentity struct {
	Type           string
	Server         string
	Port           int
	UUID           string
	Password       string
	SNI            string
	Network        string
	Security       string
	SkipCertVerify bool
}

func (p Proxy) identity() proxyIdentity {
	return proxyIdentity{
		Type:           p.Type,
		Server:         p.Server,
		Port:           p.Port,
		UUID:           p.UUID,
		Password:       p.Password,
		SNI:            p.SNI,
		Network:        p.Network,
		Security:       p.Security,
		SkipCertVerify: p.SkipCertVerify,
	}
}

func (p Proxy) Equal(other Proxy) bool {
	return p.identity() == other.identity()
}

func ParseURL(raw string) (Proxy, error) {
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
