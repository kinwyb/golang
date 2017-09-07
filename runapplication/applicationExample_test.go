package runapplication_test

import (
	"git.oschina.net/kinwyb/golang/runapplication"
)

func Example() {
	app := &runapplication.Application{
		Start: func() {
			//程序内容
		},
	}
	app.Run() //启动
}

// func main() {
// 	lf, err := os.OpenFile("angel.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
// 	if err != nil {
// 		os.Exit(1)
// 	}
// 	defer lf.Close()

// 	// 日志
// 	l := log.New(lf, "", os.O_APPEND)

// 	for {
// 		cmd := exec.Command("/usr/local/bin/node", "/*****.js")
// 		err := cmd.Start()
// 		if err != nil {
// 			l.Printf("%s 启动命令失败", time.Now().Format("2006-01-02 15:04:05"), err)

// 			time.Sleep(time.Second * 5)
// 			continue
// 		}
// 		l.Printf("%s 进程启动", time.Now().Format("2006-01-02 15:04:05"), err)
// 		err = cmd.Wait()
// 		l.Printf("%s 进程退出", time.Now().Format("2006-01-02 15:04:05"), err)

// 		time.Sleep(time.Second * 1)
// 	}
// 	filePath, _ := filepath.Abs(os.Args[0]) //将命令行参数中执行文件路径转换成可用路径
// 	cmd := exec.Command(filePath, os.Args[1:]...)
// 	//将其他命令传入生成出的进程
// 	cmd.Stdin = os.Stdin //给新进程设置文件描述符，可以重定向到文件中
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	cmd.Start() //开始执行新进程，不等待新进程退出
// 	return
// }
