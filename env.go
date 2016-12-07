package main

import (
	"os"
	"strings"
)

func envToMap() map[string]string {
	out := make(map[string]string)
	for _, v := range os.Environ() {
		vp := strings.Split(v, "=")
		out[vp[0]] = vp[1]
	}

	return out
}
