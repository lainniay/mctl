// Package sub
package sub

import "go.yaml.in/yaml/v4"

type providerFile struct {
	Proxies []providerProxy `yaml:"proxies"`
}

type providerProxy struct {
	Name           string            `yaml:"name"`
	Type           string            `yaml:"type"`
	Server         string            `yaml:"server"`
	Port           int               `yaml:"port"`
	Password       string            `yaml:"password,omitempty"`
	UUID           string            `yaml:"uuid,omitempty"`
	SNI            string            `yaml:"sni,omitempty"`
	ServerName     string            `yaml:"servername,omitempty"`
	Network        string            `yaml:"network,omitempty"`
	TLS            bool              `yaml:"tls,omitempty"`
	WSOpts         *websocketOptions `yaml:"ws-opts,omitempty"`
	SkipCertVerify bool              `yaml:"skip-cert-verify,omitempty"`
	UDP            bool              `yaml:"udp"`
}

type websocketOptions struct {
	Path    string            `yaml:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

func RenderProvider(proxies []Proxy) ([]byte, error) {
	items := make([]providerProxy, 0, len(proxies))
	for _, proxy := range proxies {
		items = append(items, toProviderProxy(proxy))
	}
	return yaml.Marshal(providerFile{Proxies: items})
}

func toProviderProxy(proxy Proxy) providerProxy {
	out := providerProxy{
		Name:           proxy.Name,
		Type:           proxy.Type,
		Server:         proxy.Server,
		Port:           proxy.Port,
		Password:       proxy.Password,
		UUID:           proxy.UUID,
		Network:        proxy.Network,
		SkipCertVerify: proxy.SkipCertVerify,
		UDP:            true,
	}
	if proxy.Type == "anytls" {
		out.SNI = proxy.SNI
	}
	if proxy.Type == "vless" {
		out.ServerName = proxy.SNI
		out.TLS = proxy.Security == "tls"
		if proxy.Network == "ws" && (proxy.WSPath != "" || proxy.WSHost != "") {
			out.WSOpts = &websocketOptions{Path: proxy.WSPath}
			if proxy.WSHost != "" {
				out.WSOpts.Headers = map[string]string{"Host": proxy.WSHost}
			}
		}
	}
	return out
}
