package lib

import (
	"fmt"
	"github.com/gogf/gf/container/gtype"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/guid"
	"io"
	"os"
)

// 计数器按天和顺序计数
var Common = new(common)
var binlogIntSafe = gtype.NewInt()
var staticIntSafe = gtype.NewInt()
var binlogDatetime = gtime.Now().Timestamp()
var staticDatetime = gtime.Now().Timestamp()

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

func (c *common) ScalerDay(taskName string) string {
	switch taskName {
	case "binlog":
		defer func() {
			if binlogIntSafe.Val() >= 90000 || gtime.Now().Format("d") != gtime.New(binlogDatetime).Format("d") {
				binlogIntSafe = gtype.NewInt()
				binlogDatetime = gtime.Now().Timestamp()
			} else {
				binlogIntSafe.Add(+1)
			}
		}()
		return guid.S() + "_" + gtime.Now().Format("Y_m_d") + "_" + fmt.Sprintf("%05d", binlogIntSafe.Val())
	case "static":
		defer func() {
			if staticIntSafe.Val() >= 90000 || gtime.Now().Format("d") != gtime.New(staticDatetime).Format("d") {
				staticIntSafe = gtype.NewInt()
				staticDatetime = gtime.Now().Timestamp()
			} else {
				staticIntSafe.Add(+1)
			}
		}()
		return guid.S() + "_" + gtime.Now().Format("Y_m_d") + "_" + fmt.Sprintf("%05d", staticIntSafe.Val())
	default:
		return guid.S()
	}
}
