package utils

import (
	"fmt"
	"time"
)

func ConvertInt64(val any) int64 {
	switch v := val.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func ParseInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return 0, err
	}
	return i, nil
}

func FormatTime(t int64) string {
	return time.Unix(t, 0).Format(time.RFC3339)
}

func Timestamp() int64 {
	return time.Now().Unix()
}
