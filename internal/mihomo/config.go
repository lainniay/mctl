package mihomo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	appconfig "github.com/lainniay/mctl/internal/config"
	"go.yaml.in/yaml/v4"
)

const providerName = "MctlNodes"

type renderedConfig struct {
	Base               map[string]any           `yaml:",inline"`
	ExternalController string                   `yaml:"external-controller"`
	Secret             string                   `yaml:"secret"`
	ProxyProviders     map[string]proxyProvider `yaml:"proxy-providers"`
	ProxyGroups        []proxyGroup             `yaml:"proxy-groups"`
	Rules              []string                 `yaml:"rules"`
}

type proxyProvider struct {
	Type        string      `yaml:"type"`
	Path        string      `yaml:"path"`
	HealthCheck healthCheck `yaml:"health-check"`
}

type healthCheck struct {
	Enable   bool   `yaml:"enable"`
	Interval int    `yaml:"interval"`
	URL      string `yaml:"url"`
}

type proxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Use      []string `yaml:"use"`
	Filter   string   `yaml:"filter,omitempty"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

// RenderConfig  render config.yaml, base is base.yaml, cfg is mctl.Config
func RenderConfig(base []byte, cfg appconfig.Config, controller appconfig.Controller) ([]byte, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	baseConfig := map[string]any{}
	if err := yaml.Unmarshal(base, &baseConfig); err != nil {
		return nil, fmt.Errorf("parse base config: %w", err)
	}
	for _, key := range []string{"proxies", "proxy-providers", "proxy-groups", "rules", "secret", "external-controller"} {
		if _, exists := baseConfig[key]; exists {
			return nil, fmt.Errorf("base config contains mctl-managed key: %s", key)
		}
	}

	groups := make([]proxyGroup, 0, len(cfg.Groups))
	for _, group := range cfg.Groups {
		groups = append(groups, proxyGroup{
			Name:     group.Name,
			Type:     group.Type,
			Use:      []string{providerName},
			Filter:   group.Filter,
			URL:      group.URL,
			Interval: group.Interval,
		})
	}

	rules := make([]string, 0, len(cfg.Rules))
	for _, rule := range cfg.Rules {
		parts := append([]string{rule.Match, rule.Policy}, rule.Options...)
		rules = append(rules, strings.Join(parts, ","))
	}

	return yaml.Marshal(renderedConfig{
		Base:               baseConfig,
		ExternalController: controller.Address,
		Secret:             controller.Secret,
		ProxyProviders: map[string]proxyProvider{
			providerName: {
				Type: "file",
				Path: "./providers/nodes.yaml",
				HealthCheck: healthCheck{
					Enable:   true,
					Interval: 300,
					URL:      "https://www.gstatic.com/generate_204",
				},
			},
		},
		ProxyGroups: groups,
		Rules:       rules,
	})
}

func ValidateConfig(ctx context.Context, data []byte, home string) error {
	file, err := os.CreateTemp("", "mctl-config-*.yaml")
	if err != nil {
		return err
	}
	name := file.Name()

	defer func() {
		_ = os.Remove(name)
	}()

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	output, err := exec.CommandContext(ctx, "mihomo", "-t", "-d", home, "-f", name).CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return fmt.Errorf("validate config: %w", err)
		}
		return fmt.Errorf("validate config: %s: %w", message, err)
	}
	return nil
}
