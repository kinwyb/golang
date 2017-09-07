package utils

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"io/ioutil"

	"fmt"

	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/clientv3"
	"github.com/shirou/gopsutil/process"
)

//PsInfo 进程信息
type PsInfo struct {
	CPUPercent    float64 //cpu使用率
	MemoryPercent float32 //内存使用率
	Memory        uint64  //实际使用内存
}

//NewPsInfo 获取进程系统使用信息
//@param {time.Duration} t cpu采集时间，默认1s
func NewPsInfo(t ...time.Duration) *PsInfo {
	pid := os.Getpid()
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil
	} else if t == nil || len(t) < 1 {
		t = []time.Duration{1 * time.Second}
	}
	v, _ := p.Percent(t[0])
	ret := &PsInfo{
		CPUPercent: v,
	}
	m, _ := p.MemoryInfo()
	if m != nil {
		ret.Memory = m.RSS
	}
	ret.MemoryPercent, _ = p.MemoryPercent()
	return ret
}

//EtcdPsInfo etcd中进程结构
type EtcdPsInfo struct {
	PsInfo
	Name       string //进程名称
	ExternalIP string //外部IP
	InternalIP string //内网IP
}

func (e *EtcdPsInfo) String() string {
	return fmt.Sprintf(`{"CPUPercent":%.2f,"MemoryPercent":%.2f,"Memory":%d,"Name":"%s","ExternalIP":"%s","InternalIP":"%s"}`, e.CPUPercent, e.MemoryPercent, e.Memory, e.Name, e.ExternalIP, e.InternalIP)
}

//获取外网IP
func externalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	ret, _ := ioutil.ReadAll(resp.Body)
	return strings.TrimSpace(string(ret))
}

//获取内网IP
func internalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

//NotifyEtcdPsInfo 更新数据应用信息到etcd
//@param {string} name 名称
//@parma {[]string} endpoints etcd节点ip地址
//@param {Logger} lg 日志
//@param {Logger} psinfolg 性能采集结果日志记录
func NotifyEtcdPsInfo(name string, endpoints []string, lg Logger, psinfolg ...Logger) {
	if lg == nil {
		lg = logs.NewLogger()
	}
	defer func() {
		if err := recover(); err != nil {
			lg.Error("NotifyEtcdPsInfo异常崩溃:%s", err)
			go NotifyEtcdPsInfo(name, endpoints, lg, psinfolg...)
		}
	}()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		lg.Error("etcd连接失败:%s", err.Error())
		return
	}
	lg.Trace("连接etcd......[成功]")
	defer cli.Close()
	eIP := externalIP()
	nIP := internalIP()
	key := fmt.Sprintf("/application/psinfo/%d", time.Now().UnixNano())
	resp, err := cli.Grant(context.TODO(), 10)
	if err != nil {
		lg.Error("创建etcd租约......[失败]%s", err.Error())
		return
	}
	lg.Trace("创建etcd租约......[成功]")
	for {
		info := &EtcdPsInfo{
			PsInfo:     *NewPsInfo(5 * time.Second),
			ExternalIP: eIP,
			InternalIP: nIP,
			Name:       name,
		}
		if psinfolg != nil && len(psinfolg) > 0 {
			psinfolg[0].Info("CPU:%.2f Memory:%.2f MemoryUse:%d", info.CPUPercent, info.MemoryPercent, info.Memory)
		}
		_, err = cli.Put(context.TODO(), key, info.String(), clientv3.WithLease(resp.ID))
		cli.KeepAliveOnce(context.TODO(), resp.ID)
		if err != nil {
			lg.Warning("[%s]数据推送......[失败]%s", key, err.Error())
		} else {
			lg.Trace("[%s]数据推送......[成功]", key)
		}
	}
}
