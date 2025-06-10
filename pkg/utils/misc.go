package utils

import "time"

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

func FormatTime(t int64) string {
	return time.Unix(t, 0).Format(time.RFC3339)
}

func Timestamp() int64 {
	return time.Now().Unix()
}
