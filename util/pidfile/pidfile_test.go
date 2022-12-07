package pidfile

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWritePidFile(t *testing.T) {
	f, err := os.CreateTemp("", "pidfiletest")
	fname := f.Name()
	defer os.Remove(fname)
	assert.NoError(t, err, "failed to create file for test")

	err = WritePidFile(fname)
	assert.NoError(t, err)

	data, err := os.ReadFile(fname)
	assert.NoError(t, err)
	pid, err := strconv.Atoi(string(data))
	assert.NoError(t, err)
	assert.Equal(t, os.Getpid(), pid)
}
