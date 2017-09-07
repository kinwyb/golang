package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/beego/bee/cmd/commands"
)

var pkcs12Cmd = &commands.Command{
	CustomFlags: true,
	UsageLine:   "pkcs12 -filename=\"\" -crt=\"\" [-key=\"\"] [-ca=\"\"]",
	Short:       "创建一个pkcs12证书",
	Long:        `pkcs12 用来创建一个pkcs12证书,依赖于openssl命令`,
	Run:         createPKCS12,
}

var (
	pkcs12Crt    string
	pkcs12CrtKey string
	pkcs12CA     string
)

func init() {
	fs := flag.NewFlagSet("pkcs12", flag.ContinueOnError)
	fs.StringVar(&caFileName, "filename", "", "导出文件名")
	fs.StringVar(&pkcs12Crt, "crt", "", "证书")
	fs.StringVar(&pkcs12CrtKey, "key", "", "证书密钥[没有被指定私钥必须包含在证书当中]")
	fs.StringVar(&pkcs12CA, "ca", "", "CA证书")
	pkcs12Cmd.Flag = *fs
	commands.AvailableCommands = append(commands.AvailableCommands, pkcs12Cmd)
}

func createPKCS12(cmd *commands.Command, args []string) int {
	if err := cmd.Flag.Parse(args); err != nil {
		log.Fatalf("Error while parsing flags: %v", err.Error())
		return 1
	} else if pkcs12Crt == "" {
		log.Fatalf("crt证书不能为空")
		return 1
	}
	if caFileName == "" {
		log.Printf("没有设置文件名,将使用默认值:pkcs12")
		caFileName = "pkcs12"
	}
	params := []string{"pkcs12", "-export", "-in", pkcs12Crt}
	if pkcs12CrtKey != "" {
		params = append(params, "-inkey", pkcs12CrtKey)
	}
	if pkcs12CA != "" {
		params = append(params, "-CAfile", pkcs12CA)
	}
	params = append(params, "-chain", "-out", caFileName+".pfx")
	osCmd := exec.Command("openssl", params...)
	osCmd.Stderr = os.Stderr
	// 运行命令
	if err := osCmd.Start(); err != nil {
		log.Fatal(err.Error())
		return 1
	}
	err := osCmd.Wait()
	if err != nil {
		return 1
	}
	return 0
}
