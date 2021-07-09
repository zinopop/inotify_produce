package main

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"inotify_produce/task"
	"time"
)

func main() {

	//for  {
	//	//time.Sleep(time.Second * 1)
	//	fmt.Println("binglog",lib.Common.ScalerDay("binlog"))
	//	fmt.Println("static",lib.Common.ScalerDay("static"))
	//}

	g.Cfg().SetFileName("config.toml")

	// 注入任务
	taskChan := taskInit(
		task.StaticFile.CreateFile,
		task.MysqlBinlog.CreateFile,
	)

	// 用select模型阻塞住主线程
	for {
		select {
		case taskName := <-taskChan:
			fmt.Println(taskName)
		default:
			// fmt.Println("没有可执行的任务")
			time.Sleep(time.Second * 1)

		}
	}

}

// 异步任务初始化
func taskInit(callback ...func(ic chan string)) chan string {
	ch := make(chan string)
	for _, val := range callback {
		go val(ch)
	}
	return ch
}
