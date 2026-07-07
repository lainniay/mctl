package sub

import "testing"

func Test_Clean_removesBlockedNodes_whenNamesAreAdsOrStatus(t *testing.T) {
	// Given
	proxies := []Proxy{
		{Name: "剩余流量：384.33 GB", Type: "anytls", Server: "ad.example.com", Port: 443, Password: "ad"},
		{Name: "🇨🇳[官方]👇Telegram群组", Type: "anytls", Server: "group.example.com", Port: 443, Password: "group"},
		{Name: "HK 01", Type: "anytls", Server: "hk.example.com", Port: 443, Password: "secret", SNI: "hk.example.com"},
	}

	// When
	got := Clean(proxies)

	// Then
	if len(got) != 1 {
		t.Fatalf("expected 1 proxy, got %d", len(got))
	}
	if got[0].Name != "Hong Kong 01" {
		t.Fatalf("expected Hong Kong 01, got %q", got[0].Name)
	}
}

func Test_Clean_dedupesAndRenames_whenConnectionsRepeat(t *testing.T) {
	// Given
	proxies := []Proxy{
		{Name: "HK 01", Type: "anytls", Server: "hk1.example.com", Port: 443, Password: "same", SNI: "hk1.example.com"},
		{Name: "香港 备用", Type: "anytls", Server: "hk1.example.com", Port: 443, Password: "same", SNI: "hk1.example.com"},
		{Name: "Pre HK 01", Type: "anytls", Server: "hk2.example.com", Port: 443, Password: "pro", SNI: "hk2.example.com"},
		{Name: "JP 01", Type: "anytls", Server: "jp1.example.com", Port: 443, Password: "jp", SNI: "jp1.example.com"},
	}

	// When
	got := Clean(proxies)

	// Then
	wantNames := []string{"Hong Kong 01", "Hong Kong 02 Pro", "Japan 01"}
	if len(got) != len(wantNames) {
		t.Fatalf("expected %d proxies, got %d", len(wantNames), len(got))
	}
	for idx, want := range wantNames {
		if got[idx].Name != want {
			t.Fatalf("expected proxy %d name %q, got %q", idx, want, got[idx].Name)
		}
	}
}
