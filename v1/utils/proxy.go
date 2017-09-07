package utils

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//RequestHandler 请求处理
//
//@params - *http.Request 请求对象体
type RequestHandler func(*http.Request)

//ResponseHandler 返回处理函数
//
//@params - *http.Response 返回内容体
//
//@return bool 如果返回false表示处理过程截至,否则会将respose对象复制给Proxy.Resposne
type ResponseHandler func(*http.Response) bool

//Proxy 代理对象
type Proxy struct {
	URL        *url.URL            //请求的URL
	AppendPath bool                //拼接请求的地址,如果Request不为空且存在URL.Path.则最终的请求Path = URL.Path + Request.URL.Path
	Request    *http.Request       //请求对象如果不设置默认生成一个空的请求对象
	Response   http.ResponseWriter //请求返回对象如果不设置请调用ExecuteResponse方法返回自动生成的Response对象
	ReqHandler RequestHandler      //请求处理函数,在正式请求之前会调用该函数方便设置需要的请求信息
	RspHandler ResponseHandler     //返回处理函数,在请求返回时会调用该函数,用于处理请求结果如果返回false这过程结束,如果返回true这会将内容复制给Response对象
	err        error               //错误对象
}

//Response 返回数据接口
type Response interface {
	//获取返回的请求头
	Header() http.Header
	//获取返回的内容
	Body() []byte
	//返回状态码
	Status() int
}

//返回头信息设置
type resp struct {
	header     http.Header
	buffer     bytes.Buffer
	statuscode int
}

func (r *resp) Header() http.Header {
	return r.header
}

func (r *resp) Write(data []byte) (int, error) {
	return r.buffer.Write(data)
}

func (r *resp) WriteHeader(statuscode int) {
	r.statuscode = statuscode
}

func (r *resp) Body() []byte {
	return r.buffer.Bytes()
}

func (r *resp) Status() int {
	return r.statuscode
}

//NewProxyForString 根据字符串创建代理对象
func NewProxyForString(urlpath string) *Proxy {
	proxy := &Proxy{AppendPath: false}
	remote, err := url.Parse(urlpath)
	if err != nil {
		proxy.err = err
	} else {
		proxy.URL = remote
	}
	return proxy
}

//NewProxyForURL 根据URL创建代理对象
func NewProxyForURL(url *url.URL) *Proxy {
	return &Proxy{URL: url}
}

//SetRequestAndResposne 设置请求和返回
//
//如果不设置默认生成空的http.request和proxy.Response对象
func (p *Proxy) SetRequestAndResposne(req *http.Request, resp http.ResponseWriter) {
	p.Request = req
	p.Response = resp
}

//do 代理请求
func (p *Proxy) do(params url.Values, method ...string) (*http.Response, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	req := new(http.Request)
	if p.Request != nil {
		*req = *p.Request
		req.URL.Scheme = p.URL.Scheme
		req.URL.Host = p.URL.Host
		if p.AppendPath {
			req.URL.Path = singleJoiningSlash(p.URL.Path, req.URL.Path)
		} else {
			req.URL.Path = p.URL.Path
		}
	} else {
		req.URL = p.URL
		req.Header = http.Header{}
		req.Method = "GET"
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
	if method != nil && len(method) > 0 {
		req.Method = method[0]
	}
	if params != nil {
		targetQuery := params.Encode()
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}
	if strings.ToLower(req.Method) == "post" && req.PostForm != nil {
		buf := bytes.NewBufferString(req.PostForm.Encode())
		req.Body = ioutil.NopCloser(buf)
		req.Header.Set("ContentLength", strconv.FormatInt(int64(buf.Len()), 10))
		req.ContentLength = int64(buf.Len())
	} else {
		req.Header.Set("ContentLength", "0")
		req.ContentLength = 0
	}
	if p.ReqHandler != nil {
		p.ReqHandler(req)
	}
	req.Host = ""
	req.RequestURI = ""
	req.RemoteAddr = ""
	return client.Do(req)
}

//Execute 执行代理,返回请求错误内容..该方法必须设置Response对象才可调用,否则请调用ExecuteResposne方法
func (p *Proxy) Execute(params url.Values, method ...string) {
	if p.err != nil {
		return
	} else if p.Response == nil {
		p.err = errors.New("请设置返回对象,如无返回对象请调用ExecuteNewResponse方法")
	}
	resp, err := p.do(params, method...)
	if err != nil {
		p.Response.WriteHeader(http.StatusInternalServerError)
		p.err = err
		return
	}
	if p.RspHandler != nil && !p.RspHandler(resp) { //执行 ResponseHandler 函数方法返回false,结束后续流程
		return
	}
	copyHeader(p.Response.Header(), resp.Header)
	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	if len(resp.Trailer) > 0 {
		var trailerKeys []string
		for k := range resp.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		p.Response.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}
	p.Response.WriteHeader(resp.StatusCode)
	if len(resp.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := p.Response.(http.Flusher); ok {
			fl.Flush()
		}
	}
	p.copyResponse(p.Response, resp.Body)
	resp.Body.Close() // close now, instead of defer, to populate res.Trailer
	copyHeader(p.Response.Header(), resp.Trailer)
}

//Error 返回请求错误内容
func (p *Proxy) Error() error {
	return p.err
}

//ExecuteResponse 执行请求返回请求结果,生成Response并返回
func (p *Proxy) ExecuteResponse(params url.Values, method ...string) Response {
	if p.err != nil {
		return nil
	}
	response := &resp{
		header: http.Header{},
		buffer: bytes.Buffer{},
	}
	p.Response = response
	p.Execute(params, method...)
	if p.err != nil {
		return nil
	}
	switch response.Header().Get("Content-Encoding") {
	case "gzip":
		reader, _ := gzip.NewReader(&response.buffer)
		defer reader.Close()
		data, _ := ioutil.ReadAll(reader)
		response.buffer.Truncate(0)
		response.buffer.Write(data)
		response.header.Del("Content-Encoding")
	}
	return response
}

func (p *Proxy) copyResponse(dst io.Writer, src io.Reader) {
	var buf []byte
	io.CopyBuffer(dst, src, buf)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
