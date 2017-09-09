package runapplication_test

import "github.com/kinwyb/golang/runapplication"

func Example() {
	app := &runapplication.Application{
		Start: func() {
			//程序内容
		},
	}
	app.Run() //启动
}
