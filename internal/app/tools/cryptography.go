package tools

import (
	"errors"
	"hash/fnv"
	"strconv"
)

type Generator interface {
	MakeHash(s string) (string, error)
}

type HashGenerator struct{}

func (g HashGenerator) MakeHash(s string) (string, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", errors.New("making hash error")
	}
	return strconv.Itoa(int(h.Sum32())), nil
}
