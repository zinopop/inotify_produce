package task

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/util/gconv"
	"github.com/pkg/sftp"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

var StaticFile = new(static)

type static struct{}


var (
	errGlob        error
	sftpClient *sftp.Client
)


func init(){
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
		User:    user,
		Auth:    auth,
		HostKeyCallback:func(hostname string, remote net.Addr, key ssh.PublicKey) error {
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


func(s *static) CreateFile(ch chan string){
	for {
		fileInfos,err := sftpClient.ReadDir(g.Cfg().GetString("static.dir.remoteDir"))
		//defer sftpClient.Close()
		if err != nil {
			fmt.Println("读取文件夹错误:",err)
			ch <- "创建static失败"
			break
		}

		if len(fileInfos) <= 0 {
			fmt.Println("loop文件为空:",err)
			time.Sleep(5*time.Second)
			continue
		}

		for _,val := range fileInfos {
			fmt.Println(val.Name())
			//handleFile(fileDirPath+val.Name())
		}
	}

}

