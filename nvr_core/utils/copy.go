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

// No lock map copying
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

// ConvertSlice transforms a slice of type T1 into a slice of type T2
// by applying the provided converter function to each element.
func ConvertSlice[T1 any, T2 any](input []T1, converter func(T1) T2) []T2 {
	if input == nil {
		return nil
	}

	// Pre-allocate the result slice with the exact length of the input
	// to prevent expensive memory reallocations during the loop.
	result := make([]T2, len(input))
	
	for i, v := range input {
		result[i] = converter(v)
	}

	return result
}