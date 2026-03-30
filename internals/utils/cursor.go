package utils

import (
    "encoding/base64"
    "fmt"
    "strconv"
    "strings"
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