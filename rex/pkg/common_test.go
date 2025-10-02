package pkg

import (
	"log"
	"os"
)

func init() {
	// Show the log output.
	log.SetOutput(os.Stderr)
}
