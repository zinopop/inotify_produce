package task

import (
	"fmt"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"github.com/pkg/sftp"
	"inotify_produce/lib"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var StaticFile = new(static)

type static struct{}

var (
	errGlob    error
	sftpClient *sftp.Client
)
var wgStatic sync.WaitGroup

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

	for {
		var tryTime int = 0
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
		fileSlice := make([][]string, 0)
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
			//handleFile(fileTmpArray)
			fileSlice = append(fileSlice, fileTmpArray)
			// 重新初始化处理数组
			sizeTmp = 0
			fileTmpArray = nil
		}

		wgStatic.Add(len(fileSlice))
		for index, val := range fileSlice {
			ch := make(chan int)
			go handleFile(val, ch)
			ch <- index
		}
		wgStatic.Wait()
	}
}

func sftpDown(srcPath, dstPath string) (*os.File, error) {
	var loop = func(f string) *sftp.File {
		//var filesize int64
		file, err := sftpClient.Open(f)

		if err != nil {
			return nil
		}
		//for  {
		//	fileinfo, err := file.Stat()
		//	if err != nil {
		//		return nil
		//	}
		//	if fileinfo.Size() != filesize{
		//		filesize = fileinfo.Size()
		//		time.Sleep(time.Millisecond*1001)
		//		continue
		//	}else{
		//		break
		//	}
		//}

		return file
	}(srcPath)
	if loop == nil {
		return nil, gerror.New("远程文件打开失败")
	}

	dstFile, err := os.Create(dstPath) //本地
	if err != nil {
		return nil, err
	}

	if _, err := loop.WriteTo(dstFile); err != nil {
		return nil, err
	}
	_ = loop.Close()
	_ = dstFile.Close()
	return dstFile, nil
}

func handleFile(fileNames []string, ch chan int) {
	files := make([]*os.File, 0)
	localFileNames := make([]string, 0)
	for _, val := range fileNames {
		dstPath := g.Cfg().GetString("static.dir.localBakDir") + "\\" + path.Base(val)
		_, err := sftpDown(val, dstPath)
		if err != nil {
			fmt.Println("远程下载失败", err)
			continue
		}
		if err := sftpClient.Remove(val); err != nil {
			fmt.Println("远程删除失败", err)
			continue
		}
		localFileNames = append(localFileNames, g.Cfg().GetString("static.dir.localBakDir")+"\\"+path.Base(val))
	}

	for _, val := range localFileNames {
		file, err := os.Open(val)
		if err == nil {
			files = append(files, file)
		}
	}
	s := lib.Common.ScalerDay("static")
	date := gconv.String(gtime.TimestampNano())

	compreFile, err := os.Create(g.Cfg().GetString("static.dir.localBakDir") + "\\static_" + date + "_" + s + ".zip")

	if err == nil {

		fmt.Println("开始压缩", g.Cfg().GetString("static.dir.localBakDir")+"\\static_"+date+"_"+s+".zip")
		if err := lib.Zip.Compress(files, compreFile); err != nil {
			fmt.Println("压缩失败,", err)
		}
		fmt.Println("开始拷贝", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+"_"+s+".zip.dat")
		nBytes, err := lib.Common.CopyFile(g.Cfg().GetString("static.dir.localBakDir")+"\\static_"+date+"_"+s+".zip", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+"_"+s+".zip.dat")
		if err != nil {
			fmt.Printf("Copied %d bytes!\n", nBytes)
		}
		fmt.Println("开始重命名", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+"_"+s+".zip")
		if err := os.Rename(g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+"_"+s+".zip.dat", g.Cfg().GetString("static.dir.targetDir")+"\\static_"+date+"_"+s+".zip"); err != nil {
			fmt.Println("rename", err)
		}
	}

	fmt.Println("删除临时文件")
	for _, val := range fileNames {
		os.Remove(g.Cfg().GetString("static.dir.localBakDir") + "\\" + path.Base(val))
	}

	fmt.Println(<-ch)
	wgStatic.Done()
}
