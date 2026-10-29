// Package sub
package sub

import "go.yaml.in/yaml/v4"

type providerFile struct {
	Proxies []providerProxy `yaml:"proxies"`
}

type providerProxy struct {
	Name              string            `yaml:"name"`
	Type              string            `yaml:"type"`
	Server            string            `yaml:"server"`
	Port              int               `yaml:"port"`
	Cipher            string            `yaml:"cipher,omitempty"`
	Plugin            string            `yaml:"plugin,omitempty"`
	PluginOpts        map[string]any    `yaml:"plugin-opts,omitempty"`
	UDPOverTCP        bool              `yaml:"udp-over-tcp,omitempty"`
	UDPOverTCPVersion int               `yaml:"udp-over-tcp-version,omitempty"`
	Password          string            `yaml:"password,omitempty"`
	UUID              string            `yaml:"uuid,omitempty"`
	SNI               string            `yaml:"sni,omitempty"`
	ServerName        string            `yaml:"servername,omitempty"`
	Network           string            `yaml:"network,omitempty"`
	TLS               bool              `yaml:"tls,omitempty"`
	ClientFingerprint string            `yaml:"client-fingerprint,omitempty"`
	WSOpts            *websocketOptions `yaml:"ws-opts,omitempty"`
	Flow              string            `yaml:"flow,omitempty"`
	RealityOpts       *realityOptions   `yaml:"reality-opts,omitempty"`
	SkipCertVerify    bool              `yaml:"skip-cert-verify,omitempty"`
	UDP               bool              `yaml:"udp"`
}

type realityOptions struct {
	PublicKey string `yaml:"public-key"`
	ShortID   string `yaml:"short-id,omitempty"`
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
	if proxy.Type == "ss" {
		out.Cipher = proxy.SS.Cipher
		out.UDP = proxy.SS.UDP
		out.UDPOverTCP = proxy.SS.UDPOverTCP
		out.UDPOverTCPVersion = proxy.SS.UDPOverTCPVersion
		out.Plugin = proxy.SS.Plugin
		if out.Plugin != "" {
			opts := proxy.SS.PluginOptions
			out.PluginOpts = map[string]any{"mode": opts.Mode}
			if opts.Host != "" {
				out.PluginOpts["host"] = opts.Host
			}
			if out.Plugin == "v2ray-plugin" {
				out.PluginOpts["tls"] = opts.TLS
				out.PluginOpts["mux"] = opts.Mux
				out.PluginOpts["skip-cert-verify"] = opts.SkipCertVerify
				if opts.Path != "" {
					out.PluginOpts["path"] = opts.Path
				}
			}
		}
	}
	if proxy.Type == "anytls" {
		out.SNI = proxy.SNI
	}
	if proxy.Type == "vless" {
		out.ServerName = proxy.SNI
		out.TLS = proxy.Security == "tls" || proxy.Security == "reality"
		out.ClientFingerprint = proxy.ClientFingerprint
		out.Flow = proxy.Flow
		if proxy.Security == "reality" {
			out.RealityOpts = &realityOptions{PublicKey: proxy.RealityPublicKey, ShortID: proxy.RealityShortID}
		}
		if proxy.Network == "ws" && (proxy.WSPath != "" || proxy.WSHost != "") {
			out.WSOpts = &websocketOptions{Path: proxy.WSPath}
			if proxy.WSHost != "" {
				out.WSOpts.Headers = map[string]string{"Host": proxy.WSHost}
			}
		}
	}
	return out
}
