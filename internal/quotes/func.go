package quotes

import (
	"encoding/hex"
	"hash"
	"strconv"

	"github.com/robotomize/powwy/pkg/hashcash"
)

type (
	GenerateTokenFunc func(f func() hash.Hash, header hashcash.Header) string
	HashFunc          func() hash.Hash
)

func GenerateToken(hashFn func() hash.Hash, header hashcash.Header) string {
	h := hashFn()
	h.Write([]byte(header.Nonce))
	mid := h.Sum(nil)
	h.Reset()

	h.Write(mid)
	h.Write([]byte(header.Subject))
	h.Write([]byte(header.Alg))
	h.Write([]byte(strconv.Itoa(int(header.ExpiredAt))))

	return hex.EncodeToString(h.Sum(nil))
}
