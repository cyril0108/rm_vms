package utils

import "sync"

// CopyMapValues safely extracts all values from a map into a slice.
// [K comparable, V any] defines the generic types for the map's Key and Value.
func CopyMapValues[K comparable, V any](m map[K]V, mu sync.Locker) []V {
	mu.Lock()
	defer mu.Unlock()

	list := make([]V, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}

	return list
}

func CopyMapValuesNL[K comparable, V any](m map[K]V) []V {

	list := make([]V, 0, len(m))
	for _, v := range m {
		list = append(list, v)
	}

	return list
}

// CopyMapKeys safely extracts all keys from a map into a slice.
// It returns a slice of type K (the key type).
func CopyMapKeys[K comparable, V any](m map[K]V, mu sync.Locker) []K {
	mu.Lock()
	defer mu.Unlock()

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}