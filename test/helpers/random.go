package helpers

import (
	"crypto/rand"
	"fmt"
)

func RandomString(prefix string) string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s-%x", prefix, b)
}
