package main

import (
	"fmt"
	"os"

	"gitlab.bit9.local/octarine/cbctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
