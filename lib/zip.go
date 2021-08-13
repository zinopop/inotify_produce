package lib

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
)

var Zip = new(zipLib)

type zipLib struct{}

func (z *zipLib) Compress(files []*os.File, compreFile *os.File) (err error) {
	zw := zip.NewWriter(compreFile)
	defer zw.Close()
	for _, file := range files {
		err := compressZip(file, zw)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

func compressZip(file *os.File, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {

		fmt.Println("压缩文件失败：", err.Error())
		return err
	}
	// 获取压缩头信息
	head, err := zip.FileInfoHeader(info)
	if err != nil {
		fmt.Println("压缩文件失败：", err.Error())
		return err
	}
	// 指定文件压缩方式 默认为 Store 方式 该方式不压缩文件 只是转换为zip保存
	head.Method = zip.Deflate
	fw, err := zw.CreateHeader(head)
	if err != nil {
		fmt.Println("压缩文件失败：", err.Error())
		return err
	}
	// 写入文件到压缩包中
	_, err = io.Copy(fw, file)
	file.Close()
	if err != nil {
		fmt.Println("压缩文件失败：", err.Error())
		return err
	}
	return nil
}

func IsZip(zipPath string) bool {
	f, err := os.Open(zipPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if n, err := f.Read(buf); err != nil || n < 4 {
		return false
	}

	return bytes.Equal(buf, []byte("PK\x03\x04"))
}
