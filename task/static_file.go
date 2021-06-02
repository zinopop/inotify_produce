package task

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"github.com/pkg/sftp"
	"inotify_produce/lib"
	"io"
	"log"
	"net"
	"os"
	"path"
	"time"

	"golang.org/x/crypto/ssh"
)

var StaticFile = new(static)

type static struct{}

var (
	errGlob    error
	sftpClient *sftp.Client
)

func init() {
	ip := g.Cfg().GetString("static.ssh.ip")
	port := g.Cfg().GetString("static.ssh.port")
	user := g.Cfg().GetString("static.ssh.user")
	pass := g.Cfg().GetString("static.ssh.pass")
	sftpClient, errGlob = connect(user, pass, ip, gconv.Int(port))
	if errGlob != nil {
		log.Fatal(errGlob)
	}
}

func connect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 30 * time.Second,
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

func (s *static) CreateFile(ch chan string) {
	var tryTime int = 0
	for {
		fileInfos, err := sftpClient.ReadDir(g.Cfg().GetString("static.dir.remoteDir"))
		//defer sftpClient.Close()
		if err != nil {
			fmt.Println("读取文件夹错误:", err)
			ch <- "创建static失败"
			break
		}

		if len(fileInfos) <= 0 {
			fmt.Println("static empty next loop after 2 Minute")
			time.Sleep(2 * time.Minute)
			continue
		}

		fileTmpArray := make([]string, 0)
		var sizeTmp int64 = 0
		for _, val := range fileInfos {
			// 1gb = 1048576
			// 文件大小计数 1g 一个
			//val.Size()
			fileTmpArray = append(fileTmpArray, g.Cfg().GetString("static.dir.remoteDir")+"/"+val.Name())
			sizeTmp += val.Size()
			tryTime++
			if sizeTmp <= 1073741824 && tryTime != len(fileInfos) {
				continue
			}
			// 处理文件数组
			go handleFile(fileTmpArray)
			// 重新初始化处理数组
			sizeTmp = 0
			fileTmpArray = nil
		}
	}
}

func handleFile(fileNames []string) {
	files := make([]*os.File, 0)
	localFileNames := make([]string, 0)
	for _, val := range fileNames {
		srcFile := LoopHandelRemoteFile(val)
		dstFile, err := os.Create(g.Cfg().GetString("static.dir.localBakDir") + "\\" + path.Base(val))
		if err != nil {
			fmt.Println("创建临时文件失败", err)
			continue
		}
		if _, err = srcFile.WriteTo(dstFile); err != nil {
			fmt.Println("写入临时文件失败", err)
			continue
		}
		srcFile.Close()
		localFileNames = append(localFileNames, g.Cfg().GetString("static.dir.localBakDir")+"\\"+path.Base(val))

		dstFile.Close()
		if err := sftpClient.Remove(val); err != nil {
			fmt.Println("远程删除失败", err)
		}
	}

	for _, val := range localFileNames {
		file, err := os.Open(val)
		if err == nil {
			files = append(files, file)
		}
	}

	date := gconv.String(gtime.Timestamp())

	compreFile, err := os.Create(g.Cfg().GetString("static.dir.localBakDir") + "\\static_" + date + ".zip")

	if err == nil {
		fmt.Println("开始压缩", g.Cfg().GetString("static.dir.localBakDir")+"\\static_"+date+".zip")
		if err := lib.Zip.Compress(files, compreFile); err != nil {
			fmt.Println("压缩失败,", err)
		}
		fmt.Println("开始拷贝", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+".zip.dat")
		nBytes, err := copyFile(g.Cfg().GetString("static.dir.localBakDir")+"\\static_"+date+".zip", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+".zip.dat")
		if err != nil {
			fmt.Printf("Copied %d bytes!\n", nBytes)
		}
		fmt.Println("开始重命名", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+".zip")
		if err := os.Rename(g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+".zip.dat", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+".zip"); err != nil {
			fmt.Println("rename", err)
		}
	}

	fmt.Println("删除临时文件")
	for _, val := range fileNames {
		os.Remove(g.Cfg().GetString("static.dir.localBakDir") + "\\" + path.Base(val))
	}
}

func LoopHandelRemoteFile(file string) *sftp.File {
	var filesize int64
	var loop = func(f string) *sftp.File {
		file, err := sftpClient.Open(f)
		if err != nil {
			return nil
		}
		for {
			fileinfo, err := file.Stat()
			if err != nil {
				return nil
			}
			//if !fileinfo.Mode().IsRegular() {
			//	continue
			//}else{
			//	break
			//}
			//fmt.Println("文件大小:",fileinfo.Size())
			//fmt.Println("文件map大小:",filesize)
			if fileinfo.Size() != filesize {
				filesize = fileinfo.Size()
				time.Sleep(time.Millisecond * 1001)
				continue
			} else {
				break
			}
		}
		return file
	}(file)
	return loop
}

func copyFile(src, dst string) (int64, error) {
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
