package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/nabowler/capacitycheck"
	flag "github.com/spf13/pflag"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var dir string
	flag.StringVarP(&dir, "dir", "d", "", "test directory")

	var size string
	flag.StringVarP(&size, "size", "s", "", "test file size")

	flag.Parse()

	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			os.Stderr.WriteString("unable to determine current working directory: " + err.Error() + "\n")
			os.Exit(1)
		}
		dir = cwd
	}

	if size == "" {
		os.Stderr.WriteString("size is required\n")
		os.Exit(1)
	}

	bytes, err := humanize.ParseBytes(size)
	if err != nil {
		os.Stderr.WriteString("unable to parse size: " + err.Error() + "\n")
		os.Exit(1)
	}

	err = capacitycheck.Check(bytes, dir)
	if err != nil {
		if errors.Is(capacitycheck.ChecksumFailed, err) {
			fmt.Printf("Capacity check of %s failed\n", dir)
			os.Exit(2)
		}
		os.Stderr.WriteString("capacity check did not complete due to " + err.Error() + "\n")
		os.Exit(1)
	}

	fmt.Printf("Capacity check of %s passed\n", dir)
}
