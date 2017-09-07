package gosql

import (
	"regexp"
	"strconv"
)

//err 数据库错误
type err struct {
	code int64
	msg  string
	e    error
}

//Error 错误接口
type Error interface {
	Code() int64
	Msg() string
	Err() error
	Error() string
}

var rep *regexp.Regexp

func init() {
	rep, _ = regexp.Compile("\\s?Error (\\d+):(.*)")
}

//解析错误
func formatMySQLError(e error) Error {
	if e == nil {
		return nil
	}
	eb := &err{
		code: 1,
		msg:  e.Error(),
		e:    e,
	}
	if rep.MatchString(e.Error()) {
		d := rep.FindAllStringSubmatch(e.Error(), -1)
		eb.msg = d[0][2]
		code, err := strconv.ParseInt(d[0][1], 10, 64)
		if err == nil {
			eb.code = code
		}
	}
	return eb
}

//Error 错误
func (e *err) Error() string {
	if e.msg != "" {
		return e.msg
	} else if e.e != nil {
		return e.e.Error()
	}
	return ""
}

//Code 错误代码
func (e *err) Code() int64 {
	return e.code
}

//Msg 错误消息
func (e *err) Msg() string {
	return e.msg
}

//Err 原始错误
func (e *err) Err() error {
	if e.e == nil {
		return nil
	}
	return e.e
}

//NewError 新建错误
func NewError(code int64, msg string, errs ...error) Error {
	ret := &err{code: code, msg: msg}
	if errs != nil && len(errs) > 0 {
		ret.e = errs[0]
	}
	return ret
}
