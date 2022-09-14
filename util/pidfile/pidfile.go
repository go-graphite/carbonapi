package pidfile

import (
	"os"
	"strconv"
	"strings"

	"github.com/natefinch/atomic"
)

func WritePidFile(path string) error {
	pid := strconv.Itoa(os.Getpid())
	myReader := strings.NewReader(pid)
	return atomic.WriteFile(path, myReader)
}
