package task

import (
	"fmt"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"inotify_produce/lib"
	"os"
	"sync"
	"time"
)

var MysqlBinlog = new(mysqlBinlog)

type mysqlBinlog struct{}

var wg sync.WaitGroup

func init() {
	config := gredis.Config{
		Host: g.Cfg().GetString("mysql.redis.ip"),
		Port: g.Cfg().GetInt("mysql.redis.port"),
		Pass: g.Cfg().GetString("mysql.redis.pass"),
		Db:   g.Cfg().GetInt("mysql.redis.db"),
	}
	gredis.SetConfig(&config, "default")
}

func (m *mysqlBinlog) CreateFile(ch chan string) {
	for {
		result, err := g.Redis().DoVar("KEYS", "messages*")
		if err != nil {
			fmt.Println("redis:", err)
			ch <- "创建binlog失败"
		}
		resultArray := result.Array()
		otherKey := []string{"contacts", "medias", "targets", "users", "grouplist", "message_collect"}

		for _, vv := range otherKey {
			if key, _ := g.Redis().DoVar("KEYS", vv); len(key.Array()) > 0 {
				resultArray = append(resultArray, vv)
			}
		}
		if len(resultArray) <= 0 {
			fmt.Println("binlog empty next loop after 2 Minute")
			time.Sleep(time.Minute * 2)
			continue
		}
		wg.Add(len(resultArray))
		for index, val := range resultArray {
			ch := make(chan int)
			go createFile(gconv.String(val), ch)
			ch <- index
		}
		wg.Wait()
		fmt.Println("loop done next")
	}
}

func createFile(key string, ic chan int) {
	var writeString string
	var insertNum = 0
	fileTmpNameArray := make([]string, 0)

	bakDir := g.Cfg().GetString("mysql.dir.localBakDir")

	// 创建文件
	createTmpMethod := func() {
		fileTmpName := bakDir + "\\" + gconv.String(gtime.TimestampNano()) + "-" + key + ".binlog"
		file, err := os.Create(fileTmpName)
		if err != nil {
			fmt.Println("create err", err)
		}
		fileTmpNameArray = append(fileTmpNameArray, fileTmpName)
		file.WriteString(writeString)
		file.Close()
	}
	// 初始化阈值
	initLoopMethod := func() {
		insertNum = 0
		writeString = ""
	}

	for {
		result, _ := g.Redis().DoVar("LPOP", key)
		if insertNum == g.Cfg().GetInt("mysql.dir.rowSize") {
			createTmpMethod()
			initLoopMethod()
			continue
		}
		if result.String() == "" && insertNum != 0 {
			createTmpMethod()
			break
		}
		writeString += result.String() + "\n"
		insertNum++
	}

	/*******************************************新版带密码开始*********************************************/
	if len(fileTmpNameArray) > 0 {
		fileName := "mysql_binlog_" + lib.Common.ScalerDay("binlog") + ".zip"
		zipName := g.Cfg().GetString("mysql.dir.localBakDir") + "\\" + fileName
		fmt.Println("开始压缩", zipName)
		err := lib.ZipPlus.Zip(zipName, g.Cfg().GetString("zip.password"), fileTmpNameArray)
		if err != nil {
			fmt.Println("压缩失败,", err)
		}
		fmt.Println("压缩完成", zipName)
		fmt.Println("开始拷贝", g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName)
		if nBytes, err := lib.Common.CopyFile(zipName, g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName+".dat"); err != nil {
			fmt.Printf("Copied %d bytes!\n", nBytes)
			//失败后拷贝
			lib.Common.CopyFile(zipName, g.Cfg().GetString("mysql.dir.failDir")+"\\"+fileName)
		}
		if err := os.Rename(g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName+".dat", g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName); err != nil {
			fmt.Println("rename", err)
		}

		fmt.Println("删除临时文件")
		for _, val := range fileTmpNameArray {
			os.Remove(val)
		}
	} else {
		fmt.Println("没有压缩的文件")
	}
	/*******************************************新版带密码结束*********************************************/

	/*******************************************不带密码开始*********************************************/
	//zipFileObj := make([]*os.File, 0)
	//for _, val := range fileTmpNameArray {
	//	file, err := os.Open(val)
	//	if err == nil {
	//		zipFileObj = append(zipFileObj, file)
	//	}
	//}
	//
	//if len(zipFileObj) > 0 {
	//	//fileName := "mysql_binlog_" + gconv.String(gtime.TimestampNano()) + "_" + lib.Common.ScalerDay("binlog") + ".zip"
	//	fileName := "mysql_binlog_" + lib.Common.ScalerDay("binlog") + ".zip"
	//	zipName := g.Cfg().GetString("mysql.dir.localBakDir") + "\\" + fileName
	//
	//	// 压缩部分开始
	//	compreFile, err := os.Create(zipName)
	//	if err == nil {
	//		fmt.Println("开始压缩", zipName)
	//		if err := lib.Zip.Compress(zipFileObj, compreFile); err != nil {
	//			fmt.Println("压缩失败,", err)
	//		}
	//		fmt.Println("压缩完成", zipName)
	//	}
	//	// 压缩部分开始
	//
	//	fmt.Println("开始拷贝", g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName)
	//	if nBytes, err := lib.Common.CopyFile(zipName, g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName+".dat"); err != nil {
	//		fmt.Printf("Copied %d bytes!\n", nBytes)
	//		//失败后拷贝
	//		lib.Common.CopyFile(zipName, g.Cfg().GetString("mysql.dir.failDir")+"\\"+fileName)
	//	}
	//	if err := os.Rename(g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName+".dat", g.Cfg().GetString("mysql.dir.targetDir")+"\\"+fileName); err != nil {
	//		fmt.Println("rename", err)
	//	}
	//
	//	fmt.Println("删除临时文件")
	//	for _, val := range fileTmpNameArray {
	//		os.Remove(val)
	//	}
	//} else {
	//	fmt.Println("没有压缩的文件")
	//}
	/*******************************************不带密码结束*********************************************/
	time.Sleep(time.Second * 10)
	fmt.Println("睡")
	fmt.Println(<-ic)
	wg.Done()
}
