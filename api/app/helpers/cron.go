package helpers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var serviceNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,62}$`)

func IsValidServiceName(name string) bool {
	return serviceNameRe.MatchString(name)
}

func IsValidProcessCommand(command string) bool {
	command = strings.TrimSpace(command)
	if command == "" {
		return false
	}
	return !strings.ContainsAny(command, "\n\r")
}

// IsValidCronSchedule accepts a standard 5-field cron expression.
func IsValidCronSchedule(expr string) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}
	bounds := [5][2]int{
		{0, 59}, // minute
		{0, 23}, // hour
		{1, 31}, // day of month
		{1, 12}, // month
		{0, 7},  // day of week (0 and 7 = Sunday)
	}
	for i, field := range fields {
		if !validCronField(field, bounds[i][0], bounds[i][1]) {
			return false
		}
	}
	return true
}

func validCronField(field string, min, max int) bool {
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		step := 1
		base := part
		if strings.Contains(part, "/") {
			bits := strings.Split(part, "/")
			if len(bits) != 2 {
				return false
			}
			base = bits[0]
			n, err := strconv.Atoi(bits[1])
			if err != nil || n < 1 {
				return false
			}
			step = n
		}
		if base == "*" {
			if step < 1 {
				return false
			}
			continue
		}
		if strings.Contains(base, "-") {
			ends := strings.Split(base, "-")
			if len(ends) != 2 {
				return false
			}
			lo, err1 := strconv.Atoi(ends[0])
			hi, err2 := strconv.Atoi(ends[1])
			if err1 != nil || err2 != nil || lo > hi || lo < min || hi > max {
				return false
			}
			continue
		}
		n, err := strconv.Atoi(base)
		if err != nil || n < min || n > max {
			return false
		}
		_ = step
	}
	return true
}

func TruncateTail(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return "…\n" + s[len(s)-max:]
}

func StringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return ""
	}
}
