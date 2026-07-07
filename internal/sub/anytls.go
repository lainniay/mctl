package sub

import (
	"net/url"
	"strconv"
)

func parseAnyTLS(raw string) (Proxy, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Proxy{}, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return Proxy{}, err
	}
	return Proxy{
		Name:           u.Fragment,
		Type:           "anytls",
		Server:         u.Hostname(),
		Port:           port,
		Password:       u.User.Username(),
		SNI:            u.Query().Get("sni"),
		SkipCertVerify: isTruthy(u.Query().Get("insecure")),
	}, nil
}
