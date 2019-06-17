package utils

import (
	"bytes"
	"os"
	"strings"
)

func ReadFileContents(filePath string, logger *Logger) string {
	f, err := os.Open(filePath)
	if err != nil {
		logger.PanicWithMessage("failed to open godeps file. Error: %v", err)
	}
	defer f.Close()
	var n int64 = bytes.MinRead
	if fi, err := f.Stat(); err == nil {
		if size := fi.Size() + bytes.MinRead; size > n {
			n = size
		}
	}
	var buf bytes.Buffer
	if int64(int(n)) == n {
		buf.Grow(int(n))
	}
	_, err = buf.ReadFrom(f)
	if err != nil {
		logger.PanicWithMessage("failed to read from godeps file. Error: %v", err)
	}
	return string(buf.Bytes())
}

func DirExists(dirpath string) bool {
	_, err := os.Stat(dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func ClearQuotes(s string) string {
	return strings.Replace(strings.Replace(s, "\"", "", -1), "'", "", -1)
}
