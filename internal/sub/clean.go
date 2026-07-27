package sub

import (
	"fmt"
	"slices"
	"strings"
)

var blockName = []string{
	"剩余",
	"流量",
	"重置",
	"Telegram",
	"telegram",
	"加入",
	"群组",
	"客服",
	"官网",
	"频道",
	"到期",
	"DIRECT",
	"欢迎",
}

func isBlocked(name string) bool {
	for _, word := range blockName {
		if strings.Contains(name, word) {
			return true
		}
	}
	return false
}

func filterAD(proxies []Proxy) []Proxy {
	return slices.DeleteFunc(proxies, func(e Proxy) bool {
		return isBlocked(e.Name)
	})
}

func dedupeNode(proxies []Proxy) []Proxy {
	out := proxies[:0]
	for _, proxy := range proxies {
		if slices.ContainsFunc(out, proxy.Equal) {
			continue
		}
		out = append(out, proxy)
	}
	return out
}

type MapRule struct {
	Key    string
	Values []string
}

var regionRules = []MapRule{
	{Key: "HongKong", Values: []string{"🇭🇰", "HK", "hk", "香港", "Hong Kong", "HongKong", "hongkong"}},
	{Key: "Japan", Values: []string{"🇯🇵", "JP", "jp", "日本", "Japan", "japan"}},
	{Key: "Taiwan", Values: []string{"🇹🇼", "TW", "tw", "台湾", "Taiwan", "taiwan"}},
	{Key: "Singapore", Values: []string{"🇸🇬", "SG", "sg", "新加坡", "Singapore", "singapore"}},
	{Key: "UnitedStates", Values: []string{"🇺🇸", "US", "us", "美国", "United States", "America"}},
	{Key: "Cloudfare", Values: []string{"CF", "Cloudfare"}},
	{Key: "SouthAfrica", Values: []string{"🇿🇦", "ZA", "za", "南非", "South Africa"}},
	{Key: "India", Values: []string{"🇮🇳", "IN", "in", "印度", "India"}},
	{Key: "Turkey", Values: []string{"🇹🇷", "TR", "tr", "土耳其", "Turkey"}},
	{Key: "Egypt", Values: []string{"🇪🇬", "EG", "eg", "埃及", "Egypt"}},
	{Key: "Mexico", Values: []string{"🇲🇽", "MX", "mx", "墨西哥", "Mexico"}},
	{Key: "Nigeria", Values: []string{"🇳🇬", "NG", "ng", "尼日利亚", "Nigeria"}},
	{Key: "Brazil", Values: []string{"🇧🇷", "BR", "br", "巴西", "Brazil"}},
	{Key: "Vietnam", Values: []string{"🇻🇳", "VN", "vn", "越南", "Vietnam"}},
	{Key: "Argentina", Values: []string{"🇦🇷", "AR", "ar", "阿根廷", "Argentina"}},
	{Key: "UnitedArabEmirates", Values: []string{"🇦🇪", "AE", "ae", "阿联酋", "United Arab Emirates", "UAE"}},
}

var tierRules = []MapRule{
	{Key: "Pro", Values: []string{"Pro", "pro", "PRO", "Premium", "premium", "Prem", "Pre", "pre", "PRE", "高级"}},
}

func detectKey(name string, mapRule []MapRule) string {
	for _, rule := range mapRule {
		for _, val := range rule.Values {
			if strings.Contains(name, val) {
				return rule.Key
			}
		}
	}
	return ""
}

func formatName(proxies []Proxy) []Proxy {
	out := slices.Clone(proxies)
	counts := map[string]int{}
	for idx := range out {
		region := detectKey(out[idx].Name, regionRules)
		if region == "" {
			continue
		}
		counts[region]++
		name := fmt.Sprintf("%s-%02d", region, counts[region])
		if tier := detectKey(out[idx].Name, tierRules); tier != "" {
			name += "-" + tier
		}
		out[idx].Name = name
	}
	return out
}

func Clean(proxies []Proxy) []Proxy {
	proxies = filterAD(proxies)
	proxies = dedupeNode(proxies)
	proxies = formatName(proxies)
	return proxies
}
