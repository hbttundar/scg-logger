package utils

// SanitizeKV normalizes key-value pairs for structured logging.
// - If the length is odd, it drops the last dangling element and appends kv_error="odd_length".
// - Keys must be strings; non-string keys degrade to an empty string to avoid panics.
func SanitizeKV(kv []any) []any {
	if len(kv) == 0 {
		return kv
	}

	const two = 2

	out := make([]any, 0, len(kv)+two)
	pairs := len(kv) / two

	for i := 0; i < pairs*two; i += two {
		key := kv[i]
		val := kv[i+1]

		var ks string
		switch k := key.(type) {
		case string:
			ks = k
		default:
			ks = ""
		}

		out = append(out, ks, val)
	}

	if len(kv)%two != 0 {
		out = append(out, "kv_error", "odd_length")
	}

	return out
}
