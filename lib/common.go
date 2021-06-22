package lib

import (
	"fmt"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"io"
	"os"
)

var Common = new(common)

type common struct{}

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

// 计数器按天和顺序计数
var binlogNum = 1
var staticNum = 1

func (c *common) ScalerDay(taskName string) string {
	var s string
	if taskName == "binlog" {
		s = gtime.Now().Format("Y_m_d") + "_" + gconv.String(binlogNum)
		binlogNum++
	} else if taskName == "static" {
		s = gtime.Now().Format("Y_m_d") + "_" + gconv.String(staticNum)
		staticNum++
	}
	if binlogNum == 90000 {
		binlogNum = 1
	} else if staticNum == 90000 {
		staticNum = 1
	}
	return s
}
