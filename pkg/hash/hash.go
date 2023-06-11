package hash

import (
	"crypto/sha1"
	"fmt"
)

func GenerateHash(value, salt string) string {
	hash := sha1.New()
	hash.Write([]byte(value))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
