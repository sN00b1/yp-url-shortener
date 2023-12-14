package tools

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

func MakeHash(s []byte) string {
	h := fnv.New32a()
	_, err := h.Write(s)
	if err != nil {
		fmt.Println("makeHash error")
		return ""
	}
	return strconv.Itoa(int(h.Sum32()))
}
