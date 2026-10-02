package executor

import (
	"strconv"
	"strings"
)

func isValidCronSchedule(expr string) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}
	bounds := [5][2]int{
		{0, 59},
		{0, 23},
		{1, 31},
		{1, 12},
		{0, 7},
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
		}
		if base == "*" {
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
	}
	return true
}

func isValidProcessCommand(command string) bool {
	command = strings.TrimSpace(command)
	if command == "" {
		return false
	}
	return !strings.ContainsAny(command, "\n\r")
}
