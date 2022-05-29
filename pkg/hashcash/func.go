package hashcash

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

func uint64ToBase64(v uint64) string {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	base64Counter := base64.StdEncoding.EncodeToString(buf)
	return base64Counter
}

func randomString(bytesNum int) (string, error) {
	buf := make([]byte, bytesNum)

	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	randValue := base64.StdEncoding.EncodeToString(buf)

	return randValue, nil
}
