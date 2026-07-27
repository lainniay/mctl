package config

import (
	"fmt"
	"regexp"
)

var builtInPolicies = map[string]bool{
	"DIRECT":      true,
	"PASS":        true,
	"REJECT":      true,
	"REJECT-DROP": true,
}

func (c Config) Validate() error {
	if c.Version == 1 {
		return fmt.Errorf("config version 1 stores subscriptions in mctl.json; migrate them to subscriptions.json")
	}
	if c.Version == 0 {
		return fmt.Errorf("mctl.json is missing or has no version")
	}
	if c.Version != 2 {
		return fmt.Errorf("unsupported config version: %d", c.Version)
	}

	groups := make(map[string]Group, len(c.Groups))
	for _, group := range c.Groups {
		if group.Name == "" {
			return fmt.Errorf("group name is required")
		}
		if _, exists := groups[group.Name]; exists {
			return fmt.Errorf("duplicate group: %s", group.Name)
		}
		switch group.Type {
		case "select":
			if group.URL != "" || group.Interval != 0 {
				return fmt.Errorf("select group %s cannot set url or interval", group.Name)
			}
		case "url-test":
			if group.URL == "" || group.Interval <= 0 {
				return fmt.Errorf("url-test group %s requires url and positive interval", group.Name)
			}
		default:
			return fmt.Errorf("unsupported group type %q for %s", group.Type, group.Name)
		}
		if group.Filter != "" {
			if _, err := regexp.Compile(group.Filter); err != nil {
				return fmt.Errorf("invalid filter for group %s: %w", group.Name, err)
			}
		}
		groups[group.Name] = group
	}

	if c.DefaultGroup == "" {
		return fmt.Errorf("default group is required")
	}
	defaultGroup, exists := groups[c.DefaultGroup]
	if !exists {
		return fmt.Errorf("default group not found: %s", c.DefaultGroup)
	}
	if !defaultGroup.Selectable {
		return fmt.Errorf("default group must be selectable: %s", c.DefaultGroup)
	}

	if len(c.Rules) == 0 {
		return fmt.Errorf("at least one rule is required")
	}
	for i, rule := range c.Rules {
		if rule.Match == "" || rule.Policy == "" {
			return fmt.Errorf("rule %d requires match and policy", i+1)
		}
		if _, exists := groups[rule.Policy]; !exists && !builtInPolicies[rule.Policy] {
			return fmt.Errorf("rule %d references unknown policy: %s", i+1, rule.Policy)
		}
		if rule.Match == "MATCH" && i != len(c.Rules)-1 {
			return fmt.Errorf("MATCH must be the last rule")
		}
		if rule.Match == "MATCH" && len(rule.Options) > 0 {
			return fmt.Errorf("MATCH cannot have options")
		}
	}
	if c.Rules[len(c.Rules)-1].Match != "MATCH" {
		return fmt.Errorf("last rule must be MATCH")
	}
	return nil
}
