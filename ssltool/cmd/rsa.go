package main

import (
	"flag"
	"log"

	"os"

	"io/ioutil"

	"github.com/beego/bee/cmd/commands"
	"github.com/kinwyb/golang/crypto/rsautil"
)

var rsaCmd = &commands.Command{
	CustomFlags: true,
	UsageLine:   "rsa -fileName=\"\" -size=2048 [-pkcs=pkcs8|pkcs1]",
	Short:       "创建RSA密钥",
	Long:        `rsa 用来创建一RSA密钥`,
	Run:         createRSA,
}

var (
	size int
	pkcs string
)

func init() {
	fs := flag.NewFlagSet("rsa", flag.ContinueOnError)
	fs.StringVar(&caFileName, "fileName", "", "保存文件名")
	fs.IntVar(&size, "size", 2048, "密钥长度默认2048")
	fs.StringVar(&pkcs, "pkcs", "pkcs8", "密钥编码类型[pkcs8|pkcs1]")
	rsaCmd.Flag = *fs
	commands.AvailableCommands = append(commands.AvailableCommands, rsaCmd)
}

func createRSA(cmd *commands.Command, args []string) int {
	if err := cmd.Flag.Parse(args); err != nil {
		log.Fatalf("Error while parsing flags: %v", err.Error())
		return 1
	}
	var priv []byte
	var pub []byte
	var err error
	if "pkcs8" == pkcs {
		priv, pub, err = rsautil.GenRsaKey(size, rsautil.PKCS8)
	} else {
		priv, pub, err = rsautil.GenRsaKey(size, rsautil.PKCS1)
	}
	err = ioutil.WriteFile(caFileName+".private.key", priv, os.ModePerm)
	if err != nil {
		log.Fatalf("私钥文件保存失败:%s", err.Error())
		return 1
	}
	err = ioutil.WriteFile(caFileName+".public.key", pub, os.ModePerm)
	if err != nil {
		log.Fatalf("公钥文件保存失败:%s", err.Error())
		return 1
	}
	return 0
}
