package main

import (
	"bytes"
	"encoding/hex"
	"os"
)

func readFileContents(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		panicWithMessage("failed to open godeps file. Error: %v", err)
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
		panicWithMessage("failed to read from godeps file. Error: %v", err)
	}
	return string(buf.Bytes())
}

func isHexString(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

func dirExists(dirpath string) bool {
	_, err := os.Stat(dirpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
