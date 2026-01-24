package utils

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

// SanitizePagination ensures limit and offset are safe for SQL.
func SanitizePagination(limit, offset int32) (int, int) {
	l := int(limit)
	o := int(offset)

	if l <= 0 {
		l = DefaultLimit
	}
	if l > MaxLimit {
		l = MaxLimit
	}
	if o < 0 {
		o = 0
	}
	
	return l, o
}