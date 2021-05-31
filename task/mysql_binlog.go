package task

import (
	"fmt"
	"time"
)

var MysqlBinlog = new(mysqlBinlog)

type mysqlBinlog struct{}

func (m *mysqlBinlog)CreateFile(ch chan string){
	for i:= 0; i< 10; i++ {
		time.Sleep(time.Second * 3)
		fmt.Println("mysql",i)
	}
	ch <- "mysql"
}
