package crypto

import (
	"fmt"
)

func GenerateSign(timestamp int64) string {
	return HmacSha1(fmt.Sprintf("%d", timestamp), "s7sf#43*&6937()")
}
