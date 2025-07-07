package main

import (
	"fmt"
	"os"

	"port-knocker/cmd"
)

// Version и BuildTime устанавливаются при сборке через ldflags
var (
	Version   = "v1.0.7"
	BuildTime = "unknown"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}
