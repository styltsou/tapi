package storage

import (
	"regexp"
	"strings"
)

var varRegex = regexp.MustCompile(`\{\{(.*?)\}\}`)

// Substitute replaces variables in the input string with values from the environment.
// It supports recursive substitution up to a certain depth.
func Substitute(input string, env map[string]string) string {
	if len(env) == 0 || !strings.Contains(input, "{{") {
		return input
	}

	result := input
	// Limit recursion to prevent infinite loops
	for i := 0; i < 5; i++ {
		if !varRegex.MatchString(result) {
			return result
		}

		result = varRegex.ReplaceAllStringFunc(result, func(match string) string {
			// Extract key from {{key}}
			key := match[2 : len(match)-2]
			key = strings.TrimSpace(key)

			if val, ok := env[key]; ok {
				return val
			}
			return match // Keep original tag if not found
		})
	}

	return result
}
