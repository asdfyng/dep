package crypto

import (
	"fmt"
	"math/rand"
	"time"
)

func MakeToken(key interface{}) string {
	token1 := MD5Encrypt(fmt.Sprintf("%v%d%d", key, time.Now().UnixNano(), rand.Int()))
	token2 := MD5Encrypt(fmt.Sprintf("%d%d%d", rand.Int(), time.Now().UnixNano(), rand.Int()))
	return token1 + token2
}
