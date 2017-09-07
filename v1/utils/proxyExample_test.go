package utils

import (
	"net/http"
	"net/url"
)

func ExampleProxy() {
	var req *http.Request
	var res http.ResponseWriter
	var params url.Values
	proxy := NewProxyForString("http://www.baidu.com")
	proxy.SetRequestAndResposne(req, res)
	proxy.ReqHandler = func(*http.Request) {
		//执行请求之前调用该方法.可以进行一些请求的自定义
	}
	proxy.RspHandler = func(*http.Response) bool {
		//在请求返回时会调用该函数,用于处理请求结果如果返回false整个过程结束,如果返回true,进行后续将内容复制给Response对象
		return true
	}
	//直接输出内容到res对象
	proxy.Execute(params, "POST")
	//返回请求结果
	result := proxy.ExecuteResponse(params, "POST")
	if proxy.Error() != nil {
		//error.....
	}
	result.Body()   //返回内容体
	result.Header() //返回的头信息
	result.Status() //返回结果代码
}
