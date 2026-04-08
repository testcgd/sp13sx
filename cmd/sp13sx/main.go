package main

import (
	"fmt"
	"os"

	"sp13sx/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "sp13sx: %v\n", err)
		os.Exit(1)
	}
}
