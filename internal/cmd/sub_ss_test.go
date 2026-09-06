package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/lainniay/mctl/internal/config"
	"go.yaml.in/yaml/v4"
)

func TestSubAddSSAndUpdateWithRemoteSubscription(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	password := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{255}, 32))
	raw := "ss://" + url.UserPassword("2022-blake3-aes-256-gcm", password).String() + "@ss.example.com:24443#JP-SS2022"
	t.Setenv("MCTL_SUB_URL", raw)
	if err := subAddCmd.RunE(subAddCmd, []string{"personal"}); err != nil {
		t.Fatal(err)
	}
	// Remote URI lists use the same parser and dedupe with direct nodes.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, raw)
		fmt.Fprintln(w, "anytls://test-password@hk.example.com:443#HK")
	}))
	defer server.Close()
	subs, err := config.LoadSubs()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs.Items) != 1 || subs.Items[0].URL != raw || !subs.Items[0].Enabled {
		t.Fatal("direct source not persisted")
	}
	if err := subs.AddSub("remote", server.URL); err != nil {
		t.Fatal(err)
	}
	// Even invalid direct sources are skipped when disabled.
	subs.Items = append(subs.Items, config.Sub{Name: "disabled", URL: "ss://invalid", Enabled: false})
	if err := subs.Save(); err != nil {
		t.Fatal(err)
	}
	count, path, err := runSubUpdate()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count=%d, want 2", count)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Proxies []struct {
			Name, Type, Server, Cipher, Password string
			Port                                 int
			UDP                                  bool
		} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	ss := got.Proxies[0]
	if ss.Type != "ss" || ss.Name != "Japan-01-SS" || ss.Cipher != "2022-blake3-aes-256-gcm" || ss.Password != password || !ss.UDP || ss.Server != "ss.example.com" || ss.Port != 24443 {
		t.Fatal("SS config not preserved")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("provider permissions=%o", info.Mode().Perm())
	}
	var output bytes.Buffer
	subListCmd.SetOut(&output)
	t.Cleanup(func() { subListCmd.SetOut(nil) })
	if err := subListCmd.RunE(subListCmd, nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), password) || strings.Contains(output.String(), "ss://") {
		t.Fatal("list leaks node credentials")
	}

	if err := setSubEnabled("remote", false); err != nil {
		t.Fatal(err)
	}
	if err := setSubEnabled("personal", false); err != nil {
		t.Fatal(err)
	}
	if count, _, err := runSubUpdate(); err != nil || count != 0 {
		t.Fatalf("disable: count=%d, err=%v", count, err)
	}
	if err := setSubEnabled("personal", true); err != nil {
		t.Fatal(err)
	}
	if count, _, err := runSubUpdate(); err != nil || count != 1 {
		t.Fatalf("enable: count=%d, err=%v", count, err)
	}
	if err := subRemoveCmd.RunE(subRemoveCmd, []string{"personal"}); err != nil {
		t.Fatal(err)
	}
	if count, _, err := runSubUpdate(); err != nil || count != 0 {
		t.Fatalf("remove: count=%d, err=%v", count, err)
	}
}

func TestSubAddInvalidSSDoesNotPersistOrLeak(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MCTL_SUB_URL", "ss://2022-blake3-aes-256-gcm:private-password@example.com:443")
	err := subAddCmd.RunE(subAddCmd, []string{"invalid"})
	if err == nil {
		t.Fatal("accepted invalid SS2022 key")
	}
	if strings.Contains(err.Error(), "private-password") {
		t.Fatal("error leaked credentials")
	}
	subs, err := config.LoadSubs()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs.Items) != 0 {
		t.Fatal("invalid source was saved")
	}
}
