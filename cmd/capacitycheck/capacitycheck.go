package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/nabowler/capacitycheck"
	flag "github.com/spf13/pflag"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var dir string
	flag.StringVarP(&dir, "dir", "d", "", "test directory")

	var size string
	flag.StringVarP(&size, "size", "s", "", "test file size")

	var bufferSize string
	flag.StringVarP(&bufferSize, "buffersize", "b", "", "buffer size")

	// TODO: verbose and progress bars don't play well together
	// var verbose bool
	// flag.BoolVarP(&verbose, "verbose", "v", false, "enable verbose (debug) messages")

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

	var bufferSizeBytes uint64
	if bufferSize != "" {
		bufferSizeBytes, err = humanize.ParseBytes(bufferSize)
		if err != nil {
			os.Stderr.WriteString("unable to parse buffer size: " + err.Error() + "\n")
			os.Exit(1)
		}

	}

	fmt.Printf("Verifying %s can hold %s (%d bytes)\n", dir, size, bytes)

	progress := mpb.New(mpb.WithWidth(64))
	writeBar := progress.AddBar(0,
		mpb.PrependDecorators(decor.Name("Write")),
		mpb.AppendDecorators(decor.Percentage()),
	)
	readBar := progress.AddBar(0,
		mpb.PrependDecorators(decor.Name("Read ")),
		mpb.AppendDecorators(decor.Percentage()),
	)

	options := capacitycheck.CheckOptions{
		OnWrite:    on(writeBar),
		OnRead:     on(readBar),
		BufferSize: bufferSizeBytes,
	}
	// if verbose {
	// 	options.DebugF = func(format string, args ...interface{}) {
	// 		msg := "DEBUG: " + fmt.Sprintf(format, args...) + "\n"
	// 		os.Stderr.WriteString(msg)
	// 	}
	// }

	ctx, cancel := context.WithCancel(context.Background())
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cancel()
	}()

	err = capacitycheck.CheckWithOptions(ctx, bytes, dir, options)
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

func on(bar *mpb.Bar) func(uint64, uint64) {
	return func(bytes, total uint64) {
		bar.SetTotal(int64(total), false)
		bar.SetCurrent(int64(bytes))
		bar.SetTotal(int64(total), bytes == total)
	}
}
