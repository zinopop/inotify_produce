package lib

import (
	"fmt"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/os/gtime"
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
var binlogIntSafe = gtype.NewInt()
var staticIntSafe = gtype.NewInt()

var binlogDatetime = gtime.Now().Timestamp()
var staticDatetime = gtime.Now().Timestamp()

func (c *common) ScalerDay(taskName string) string {
	var s string
	if taskName == "binlog" {
		s = gtime.Now().Format("Y_m_d") + "_" + fmt.Sprintf("%05d", binlogIntSafe.Val())
		binlogIntSafe.Add(+1)
	} else if taskName == "static" {
		s = gtime.Now().Format("Y_m_d") + "_" + fmt.Sprintf("%05d", staticIntSafe.Val())
		staticIntSafe.Add(+1)
	}

	if binlogIntSafe.Val() == 90000 || gtime.Now().Format("d") != gtime.New(binlogDatetime).Format("d") {
		binlogIntSafe = gtype.NewInt()
		binlogDatetime = gtime.Now().Timestamp()
	} else if staticIntSafe.Val() == 90000 || gtime.Now().Format("d") != gtime.New(staticDatetime).Format("d") {
		staticIntSafe = gtype.NewInt()
		staticDatetime = gtime.Now().Timestamp()
	}
	return s
}
