package capacitycheck

import (
	"hash/crc32"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
)

type (
	checkError string
)

const (
	ChecksumFailed checkError = "Checksum Failed: Capacity may not be truthful"
)

// Check will write at least `bytes` bytes to a file within `dir`, read the file, and verify
// that the written checksum matches the read checksum.
func Check(bytes uint64, dir string) error {
	tmpFile, err := ioutil.TempFile(dir, "capacitycheck-")
	if err != nil {
		return err
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buff := make([]byte, 512)
	outHash := crc32.NewIEEE()
	for n := uint64(0); n <= bytes; n += uint64(len(buff)) {
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
	}
	err = tmpFile.Sync()
	if err != nil {
		return err
	}

	// not sure if seeking is necessary
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return nil
	}

	inHash := crc32.NewIEEE()
	for n := uint64(0); n <= bytes; n += uint64(len(buff)) {
		_, err = io.ReadFull(tmpFile, buff)
		if err != nil {
			return err
		}

		_, err = inHash.Write(buff)
		if err != nil {
			return err
		}
	}

	if inHash.Sum32() != outHash.Sum32() {
		return ChecksumFailed
	}

	return nil
}

func (err checkError) Error() string {
	return string(err)
}
