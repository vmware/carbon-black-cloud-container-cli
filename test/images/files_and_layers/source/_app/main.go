package main

import (
	"math/rand"
	"os"
	"strconv"
)

// This folder uses _ as prefix intentionally as it should be ignored by the go tooling - see https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns
// We only want to use it to build helper executable binaries for testing

func main() {

	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "delete":
			path := os.Args[2]
			os.Remove(path)
		case "change":
			path := os.Args[2]
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			_, err = f.WriteString("random string " + strconv.Itoa(rand.Int()))
			if err != nil {
				panic(err)
			}
		default:
			return
		}
	}
}
