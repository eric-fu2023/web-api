package datastructures

import "log"

func KVtoVK[K comparable, V comparable](m map[K]V) map[V]K {
	newmap := make(map[V]K)

	for k, v := range m {
		if _, exist := newmap[v]; exist {
			log.Printf("KVtoVK key already exist (not one to one mapping. will override) source key:value = %v:%v\n", k, v)
		}
		newmap[v] = k
	}

	return newmap
}
