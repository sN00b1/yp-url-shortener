package tools

import (
	"hash/fnv"
	"strconv"
)

type HashGenerator struct{}

func (g *HashGenerator) MakeHash(s string) (string, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(h.Sum32())), nil
}
