package config

import (
	"sync/atomic"
)

type APIKeyProvider interface {
	Next() string
}

type StaticKey struct {
	key string
}

func NewStaticKey(key string) *StaticKey {
	return &StaticKey{key: key}
}

func (s *StaticKey) Next() string { return s.key }

// KeyRing rotates keys per request: 0..n-1 then wraps.
type KeyRing struct {
	keys []string
	idx  uint64
}

func NewKeyRing(keys []string) *KeyRing {
	cp := make([]string, 0, len(keys))
	for _, k := range keys {
		if k != "" {
			cp = append(cp, k)
		}
	}
	return &KeyRing{keys: cp}
}

func (r *KeyRing) Next() string {
	if len(r.keys) == 0 {
		return ""
	}
	i := atomic.AddUint64(&r.idx, 1) - 1
	return r.keys[int(i%uint64(len(r.keys)))]
}
