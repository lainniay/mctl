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

func (p Proxy) Equal(other Proxy) bool {
	return p.Type == other.Type &&
		p.Server == other.Server &&
		p.Port == other.Port &&
		p.UUID == other.UUID &&
		p.Password == other.Password &&
		p.SNI == other.SNI &&
		p.Network == other.Network &&
		p.Security == other.Security &&
		p.SkipCertVerify == other.SkipCertVerify
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
