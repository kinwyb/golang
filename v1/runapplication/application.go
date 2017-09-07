/*
package runapplication 引用程序扩展,可以在程序崩溃后可以自动重启
*/
package runapplication

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/astaxie/beego/logs"
)

//HandlerFun 回调函数
type HandlerFun func()

//Application 应用程序对象
//
//创建对象后请指定Start和Close方法..调用Run运行
//
//在出现panic后,会自动重新调用Start方法再次运行
type Application struct {
	sg       chan os.Signal
	Close    HandlerFun      //关闭回调函数
	Start    HandlerFun      //开始进程
	Recovery bool            //出现异常后是否自动恢复
	Logger   *logs.BeeLogger //日志纪录对象beego/logs
}

//Run 运行程序
func (a *Application) Run() {
	if a.Logger == nil {
		a.Logger = logs.NewLogger(512)
		a.Logger.SetLogger("console", "")
	}
	a.Logger.Trace("程序运行开始[Application.Run]")
	sg := make(chan os.Signal, 1)
	signal.Notify(sg, os.Kill, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	if a.Start != nil {
		go a.runStart()
	}
	<-sg
	if a.Close != nil {
		a.Close()
	}
	a.Logger.Trace("程序关闭[Application.Run]")
	fmt.Scanln() //等待输入
}

func (a *Application) runStart() {
	defer func() {
		if err := recover(); err != nil {
			a.Logger.Error("程序运行崩溃：％s,将重新启动运行", err) //出现异常，重新开始运行
			if a.Recovery {                          //重启程序
				go a.runStart()
			}
		}
	}()
	if a.Start != nil {
		a.Start()
	} else {
		a.sg <- syscall.SIGQUIT //退出运行
	}
}
