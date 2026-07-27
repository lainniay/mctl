package sub

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func Fetch(address string) (string, error) {
	resp, err := httpClient.Get(address)
	if err != nil {
		var urlErr *neturl.Error
		if errors.As(err, &urlErr) {
			return "", fmt.Errorf("subscription request failed: %v", urlErr.Err)
		}
		return "", errors.New("subscription request failed")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch subscription: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
