package capacitycheck_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/nabowler/capacitycheck"
)

func TestCheck(t *testing.T) {

	testSizes := []uint64{
		0, 1, 314, 1025, 1024 * 1024,
	}

	for i := range testSizes {
		size := testSizes[i]
		t.Run("Check "+strconv.FormatUint(size, 10), func(t *testing.T) {
			err := capacitycheck.Check(size, os.TempDir())
			if err != nil {
				t.Errorf("Failed to check size %d do to %v", size, err)
			}
		})
	}

}
