// Very simple utility to syntax check JSON files before deployment.
// It is only a syntax checker (does not check eg duplicate field names).
//
// Usage: jsoncheck filename

package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s filename\n", os.Args[0])
	}
	var x any
	bytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Can't read %s: %v", os.Args[1], err)
	}
	err = json.Unmarshal(bytes, &x)
	if err != nil {
		log.Fatalf("Unmarshaling data failed: %v", err)
	}
}
