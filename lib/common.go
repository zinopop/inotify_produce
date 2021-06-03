package lib

import (
	"fmt"
	"io"
	"os"
)

var Common = new(common)

type common struct {
}

func (c *common) CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)

	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.CopyN(destination, source, sourceFileStat.Size())
	return nBytes, err
}
