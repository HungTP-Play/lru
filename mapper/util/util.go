package util

func Base62Encode(n int64) string {
	const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if n == 0 {
		return string(base62Chars[0])
	}

	encoded := ""
	base := len(base62Chars)
	for n > 0 {
		remainder := int(n) % base
		n /= int64(base)
		encoded = string(base62Chars[remainder]) + encoded
	}

	return encoded
}
