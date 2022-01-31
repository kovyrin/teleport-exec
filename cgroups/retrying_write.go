package cgroups

import (
	"errors"
	"io/ioutil"
	"os"
	"syscall"
)

func retryingWriteFile(path string, data string, mode os.FileMode) error {
	// Retry writes on EINTR; see:
	//    https://github.com/golang/go/issues/38033
	for {
		err := os.WriteFile(path, []byte(data), mode)
		if err == nil {
			return nil
		} else if !errors.Is(err, syscall.EINTR) {
			return err
		}
	}
}
