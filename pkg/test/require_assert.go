package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// nolint:gochecknoglobals
	isParallel = computeIsParallel()
)

// Returns true if args is: `GOMAXPROCS=4 go test -parallel=4 ./...`
// Returns false is args is: `GOMAXPROCS=4 go test ./...`
func computeIsParallel() bool {
	// call flag.Parse() here if TestMain uses flags
<<<<<<< HEAD
	//flag.Parse()
	//for _, arg := range os.Args {
	//		if strings.HasPrefix(arg, "-test.parallel") {
	//return true
	//		}
	//}
=======
	// flag.Parse()
	// for _, arg := range os.Args {
	// 	if strings.HasPrefix(arg, "-test.parallel") {
	// 		return true
	// 	}
	// }
>>>>>>> develop
	return false
}

// NewRequire - calls `t.Parallel()` if tests were run with `GOMAXPROCS=4 go test -parallel=4 ./...`.
func NewRequire(t *testing.T) *require.Assertions {
	t.Helper()

	if isParallel {
		t.Parallel()
	}

	require := require.New(t)
	return require
}

// NewAssert - calls `t.Parallel()` if tests were run with `GOMAXPROCS=4 go test -parallel=4 ./...`.
func NewAssert(t *testing.T) *assert.Assertions {
	t.Helper()

	if isParallel {
		t.Parallel()
	}

	assert := assert.New(t)
	return assert
}
