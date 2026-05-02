package util

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeToBase62 converts a uint64 to a variable-length Base62 string.
// Uses the exact same charset as Java ShortLinkComponent: 0-9a-zA-Z.
func EncodeToBase62(num uint64) string {
	if num == 0 {
		return "0"
	}
	buf := make([]byte, 0, 12)
	for num > 0 {
		buf = append(buf, base62Chars[num%62])
		num /= 62
	}
	// Reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
