package stringutils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"sync/atomic"
)

const (
	emptyUUID = "00000000-0000-0000-0000-000000000000"
)

var (
	uuidSeed    [24]byte
	uuidCounter uint64
)

func init() {
	// Setup seed & counter once
	if _, err := rand.Read(uuidSeed[:]); err != nil {
		return
	}
	uuidCounter = binary.LittleEndian.Uint64(uuidSeed[:8])
}

// UUID generates an universally unique identifier (UUID)
func UUID() string {
	if atomic.LoadUint64(&uuidCounter) <= 0 {
		return emptyUUID
	}
	// first 8 bytes differ, taking a slice of the first 16 bytes
	x := atomic.AddUint64(&uuidCounter, 1)
	uuid := uuidSeed
	binary.LittleEndian.PutUint64(uuid[:8], x)
	uuid[6], uuid[9] = uuid[9], uuid[6]

	// RFC4122 v4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = uuid[8]&0x3f | 0x80

	// create UUID representation of the first 128 bits
	b := make([]byte, 36)
	hex.Encode(b[0:8], uuid[0:4])
	b[8] = '-'
	hex.Encode(b[9:13], uuid[4:6])
	b[13] = '-'
	hex.Encode(b[14:18], uuid[6:8])
	b[18] = '-'
	hex.Encode(b[19:23], uuid[8:10])
	b[23] = '-'
	hex.Encode(b[24:], uuid[10:16])

	return UnsafeString(b)
}
