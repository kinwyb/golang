package protocol

import (
	"bytes"
	"encoding/binary"
	"sync/atomic"
)

//Protocol 自定义协议
//解决TCP粘包问题
type Protocol struct {
	data            chan []byte   //解析成功的数据
	byteBuffer      *bytes.Buffer //数据存储中心
	dataLength      int64         //数据长度
	NoReadMsgLength int32         //未处理消息长度
	MsgLength       int32         //已经读取数据总长度
	Close           bool          //是否关闭
}

//NewProtocol 初始化一个Protocol
//
//@param chanLength int 解析成功数据缓冲区
func NewProtocol(chanLength ...int) *Protocol {
	length := 100
	if chanLength != nil && len(chanLength) > 0 {
		length = chanLength[0]
	}
	return &Protocol{
		data:            make(chan []byte, length),
		byteBuffer:      bytes.NewBufferString(""),
		NoReadMsgLength: 0,
		MsgLength:       0,
		Close:           false,
	}
}

//Packet 封包
func (p *Protocol) Packet(message []byte) []byte {
	return append(p.intToByte(len(message)), message...)
}

//Read 读取解析成功的数据
func (p *Protocol) Read() []byte {
	data := <-p.data
	atomic.AddInt32(&p.NoReadMsgLength, -1)
	return data
}

//ContinueRead 发送空数据解除Read()方法中的阻塞
func (p *Protocol) ContinueRead() {
	p.data <- []byte{}
}

//Unpack 解包
//
//解析成功的数据请用Read方法获取
func (p *Protocol) Unpack(buffer []byte) {
	p.byteBuffer.Write(buffer)
	for { //多条数据循环处理
		length := p.byteBuffer.Len()
		if length < 8 { //前面8个字节是长度
			return
		}
		p.dataLength = p.byteToInt(p.byteBuffer.Bytes()[0:8])
		if int64(length) < p.dataLength+8 { //数据长度不够,等待下次读取数据
			return
		}
		data := make([]byte, p.dataLength+8)
		p.byteBuffer.Read(data)
		atomic.AddInt32(&p.NoReadMsgLength, 1)
		atomic.AddInt32(&p.MsgLength, 1)
		p.data <- data[8:]
	}
}

func (p *Protocol) intToByte(len int) []byte {
	x := int64(len)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func (p *Protocol) byteToInt(data []byte) int64 {
	bytesBuffer := bytes.NewBuffer(data)
	var x int64
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}
