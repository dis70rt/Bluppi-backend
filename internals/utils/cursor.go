package utils

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func DecodeCursor(cursor string) (int, int, string) {
	if cursor == "" {
		return 0, 0, ""
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, 0, ""
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 3 {
		return 0, 0, ""
	}

	tier, _ := strconv.Atoi(parts[0])
	weight, _ := strconv.Atoi(parts[1])
	return tier, weight, parts[2]
}

func EncodeCursor(tier, weight int, id string) string {
	raw := fmt.Sprintf("%d|%d|%s", tier, weight, id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func EncodeTimeCursor(t time.Time, id string) string {
	raw := fmt.Sprintf("%d|%s", t.UnixNano(), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func DecodeTimeCursor(cursor string) (time.Time, string) {
	if cursor == "" {
		return time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), ""
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), ""
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC), ""
	}

	var nano int64
	fmt.Sscanf(parts[0], "%d", &nano)
	return time.Unix(0, nano), parts[1]
}
