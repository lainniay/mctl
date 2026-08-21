package sub

import (
	"reflect"
	"testing"
)

func TestDedupeNodePreservesFirstOccurrenceAndIdentityFields(t *testing.T) {
	base := Proxy{
		Name:           "first",
		Type:           "vless",
		Server:         "example.com",
		Port:           443,
		UUID:           "uuid",
		Password:       "password",
		SNI:            "sni.example.com",
		Network:        "tcp",
		Security:       "tls",
		SkipCertVerify: true,
	}
	duplicate := base
	duplicate.Name = "duplicate"

	variants := []Proxy{base}
	addVariant := func(name string, change func(*Proxy)) {
		variant := base
		variant.Name = name
		change(&variant)
		variants = append(variants, variant)
	}
	addVariant("type", func(p *Proxy) { p.Type = "anytls" })
	variants = append(variants, duplicate)
	addVariant("server", func(p *Proxy) { p.Server = "other.example.com" })
	addVariant("port", func(p *Proxy) { p.Port = 8443 })
	addVariant("uuid", func(p *Proxy) { p.UUID = "other-uuid" })
	addVariant("password", func(p *Proxy) { p.Password = "other-password" })
	addVariant("sni", func(p *Proxy) { p.SNI = "other-sni.example.com" })
	addVariant("network", func(p *Proxy) { p.Network = "ws" })
	addVariant("security", func(p *Proxy) { p.Security = "reality" })
	addVariant("skip-cert-verify", func(p *Proxy) { p.SkipCertVerify = false })

	want := append([]Proxy{base}, variants[1])
	want = append(want, variants[3:]...)
	if got := dedupeNode(variants); !reflect.DeepEqual(got, want) {
		t.Fatalf("dedupeNode() = %+v, want %+v", got, want)
	}
}
