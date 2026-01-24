package utils

func StringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func PtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func IntToPtr(i int) *int {
	return &i
}

func PtrToInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func BoolToPtr(b bool) *bool {
	return &b
}

func PtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}