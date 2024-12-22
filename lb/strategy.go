package lb

import (
	"hash/fnv"
)

type Strategy int

const (
	RoundRobin Strategy = iota
	WeightedRoundRobin
	IPHash
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
