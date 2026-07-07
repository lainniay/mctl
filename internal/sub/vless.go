package sub

import (
	"net/url"
	"strconv"
)

func parseVLESS(raw string) (Proxy, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Proxy{}, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return Proxy{}, err
	}
	query := u.Query()
	return Proxy{
		Name:           u.Fragment,
		Type:           "vless",
		Server:         u.Hostname(),
		Port:           port,
		UUID:           u.User.Username(),
		SNI:            query.Get("sni"),
		Network:        query.Get("type"),
		Security:       query.Get("security"),
		SkipCertVerify: isTruthy(query.Get("insecure")),
	}, nil
}
