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
		result2, _ := g.Redis().DoVar("KEYS", "messages*")
		result2_array := result2.Array()
		other_key := []string{"contacts", "medias", "targets", "users"}

		for _, vv := range other_key {
			if key, _ := g.Redis().DoVar("KEYS", vv); len(key.Array()) > 0 {
				result2_array = append(result2_array, vv)
			}
		}
		if len(result2_array) <= 0 {
			fmt.Println("done empty")
			time.Sleep(time.Second * 2)
			continue
		}
		wg.Add(len(result2_array))
		for index, val := range result2_array {
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
	zipFileObj := make([]*os.File, 0)
	bakDir := g.Cfg().GetString("mysql.dir.localBakDir")
	for {
		result, _ := g.Redis().DoVar("LPOP", key)

		if insertNum == g.Cfg().GetInt("mysql.dir.rowSize") {
			fileTmpName := bakDir + "\\" + gconv.String(gtime.Timestamp()) + "-" + key + ".binlog"
			file, err := os.Create(fileTmpName)
			if err != nil {
				fmt.Println("create err", err)
			}
			fileTmpNameArray = append(fileTmpNameArray, fileTmpName)
			file.WriteString(writeString)
			file.Close()
			insertNum = 0
			writeString = ""
			continue
		}
		if result.String() == "" && insertNum != 0 {
			fileTmpName := bakDir + "\\" + gconv.String(gtime.Timestamp()) + "-" + key + ".binlog"
			file, err := os.Create(fileTmpName)
			if err != nil {
				fmt.Println("create2 done err", err)
			}
			fileTmpNameArray = append(fileTmpNameArray, fileTmpName)
			file.WriteString(writeString)
			file.Close()
			break
		}
		writeString += result.String() + "\n"
		insertNum++
	}

	for _, val := range fileTmpNameArray {
		file, err := os.Open(val)
		if err == nil {
			zipFileObj = append(zipFileObj, file)
		}
	}

	if len(zipFileObj) > 0 {
		zipName := g.Cfg().GetString("mysql.dir.localBakDir") + "\\mysql_binlog_" + gconv.String(gtime.Timestamp()) + ".zip"
		compreFile, err := os.Create(zipName)
		if err == nil {
			fmt.Println("开始压缩", zipName)
			if err := lib.Zip.Compress(zipFileObj, compreFile); err != nil {
				fmt.Println("压缩失败,", err)
			}
			fmt.Println("压缩完成", zipName)
		}
		fmt.Println("删除临时文件")
		for _, val := range fileTmpNameArray {
			os.Remove(val)
		}
	} else {
		fmt.Println("没有压缩的文件")
	}

	fmt.Println(<-ic)
	wg.Done()
}
