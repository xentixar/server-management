package executor

import (
	"fmt"
	"strconv"
	"strings"
)

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func payloadBool(payload map[string]any, key string, def bool) bool {
	if payload == nil {
		return def
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return def
	}
	b, ok := raw.(bool)
	if !ok {
		return def
	}
	return b
}

func payloadInt(payload map[string]any, key string, def int) int {
	if payload == nil {
		return def
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return def
	}
	switch v := raw.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return def
		}
		return n
	default:
		return def
	}
}
