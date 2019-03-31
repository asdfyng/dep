package util

func ErrJSON(code int, msg string) map[string]interface{} {
	return map[string]interface{}{
		"Code":   code,
		"Messag": msg,
	}
}
