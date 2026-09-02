package main

import (
	"os"

	"github.com/nayan-bagale/skydock-agent/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
