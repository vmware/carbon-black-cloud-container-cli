package main

import (
	"fmt"
	"os"

	"github.com/vmware/carbon-black-cloud-container-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
