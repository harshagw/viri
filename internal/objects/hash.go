package objects

import (
	"fmt"
	"sort"
	"strings"
)

// Hash represents a string-keyed hash map.
type Hash struct {
	Pairs map[string]Object
}

func NewHash() *Hash {
	return &Hash{
		Pairs: make(map[string]Object),
	}
}

func (h *Hash) Type() Type {
	return TypeHash
}

func (h *Hash) Inspect() string {
	keys := make([]string, 0, len(h.Pairs))
	for k := range h.Pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, h.Pairs[k].Inspect()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (h *Hash) Get(key string) (Object, bool) {
	val, ok := h.Pairs[key]
	return val, ok
}

func (h *Hash) Set(key string, value Object) {
	h.Pairs[key] = value
}
