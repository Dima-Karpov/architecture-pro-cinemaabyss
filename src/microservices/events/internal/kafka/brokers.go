package kafka

import "strings"

func parseBrokers(brokers string) []string {
	parts := strings.Split(brokers, ",")

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if broker := strings.TrimSpace(part); broker != "" {
			result = append(result, broker)
		}
	}

	return result
}
