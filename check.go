package capacitycheck

import (
	"context"
	"hash/crc32"
	"io"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/dustin/go-humanize"
)

type (
	checkError string

	CheckOptions struct {
		//BufferSize will override the default read/write buffer size, in bytes
		BufferSize uint64
		//OnWrite will be called after each write to the file with the total number of bytes
		//written and the expected maximum number of bytes that will be written in total
		OnWrite func(written uint64, max uint64)
		//OnRead will be called after each read from the file with the total number of bytes
		//read and the expected maximum number of bytes that will be read in total
		OnRead func(read uint64, max uint64)
		//DebugF will write debug information without newline termination
		DebugF func(format string, args ...interface{})
	}
)

const (
	ChecksumFailed checkError = "Checksum Failed: Capacity may not be truthful"
)

// Check will write at least `bytes` bytes to a file within `dir`, read the file, and verify
// that the written checksum matches the read checksum.
func Check(ctx context.Context, bytes uint64, dir string) error {
	return CheckWithOptions(ctx, bytes, dir, CheckOptions{})
}

// CheckWithOptions does the same as Check, but provides options that can
// be set to get hooks into the in-progress check.
func CheckWithOptions(ctx context.Context, bytes uint64, dir string, options CheckOptions) error {
	debugF := debugFNoop
	onWrite := onNoop
	onRead := onNoop
	bufferSize := uint64(humanize.MByte)

	numLoops := bytes / bufferSize
	if numLoops*bufferSize < bytes {
		numLoops += 1
	}
	actualBytes := numLoops * bufferSize

	if options.BufferSize > 0 {
		bufferSize = options.BufferSize
	}
	if options.OnWrite != nil {
		onWrite = options.OnWrite
	}
	if options.OnRead != nil {
		onRead = options.OnRead
	}
	if options.DebugF != nil {
		debugF = options.DebugF
	}

	tmpFile, err := ioutil.TempFile(dir, "capacitycheck-")
	if err != nil {
		debugF("temp file creation failed: %v", err)
		return err
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	debugF("Beginning writing")
	buff := make([]byte, bufferSize)
	outHash := crc32.NewIEEE()
	totalWritten := uint64(0)
	for n := uint64(0); n <= bytes; n += uint64(len(buff)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err = rand.Read(buff)
		if err != nil {
			return err
		}

		_, err = outHash.Write(buff)
		if err != nil {
			return err
		}

		_, err = tmpFile.Write(buff)
		if err != nil {
			return err
		}

		totalWritten += bufferSize
		onWrite(totalWritten, actualBytes)
	}
	err = tmpFile.Sync()
	if err != nil {
		return err
	}
	debugF("Writing complete. Expected checksum: %d", outHash.Sum32())

	// not sure if seeking is necessary
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil
	}

	debugF("Beginnging reading")
	inHash := crc32.NewIEEE()
	totalRead := uint64(0)
	for n := uint64(0); n <= bytes; n += uint64(len(buff)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err = io.ReadFull(tmpFile, buff)
		if err != nil {
			return err
		}

		_, err = inHash.Write(buff)
		if err != nil {
			return err
		}

		totalRead += bufferSize
		onRead(totalRead, actualBytes)
	}

	if inHash.Sum32() != outHash.Sum32() {
		debugF("checksum validation failed. Expected [%d] Got [%d]", outHash.Sum32(), inHash.Sum32())
		return ChecksumFailed
	}

	return nil
}

func (err checkError) Error() string {
	return string(err)
}

func onNoop(a, b uint64)                    {}
func debugFNoop(s string, a ...interface{}) {}
