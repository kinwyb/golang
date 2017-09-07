package main

import (
	"crypto/rsa"
	"crypto/x509"
	"flag"
	"log"
	"net"
	"strings"

	"git.oschina.net/kinwyb/golang/ssltool"

	"github.com/beego/bee/cmd/commands"
)

var crtCmd = &commands.Command{
	CustomFlags: true,
	UsageLine:   "crt -filename=\"\" [-ca=\"\"] [-cakey=\"\"] [-days=10] [-months=0] [-years=0] [-country=\"\"] [-commonName=\"\"] [-organization=\"\"] [-organizationalUnit=\"\"] [-emailAddress=\"\"] [-province=\"\"] [-locality=\"\"] [-dns=\"\"] [-ip=\"\"]",
	Short:       "签发一个证书",
	Long:        `crt 用来签发一个证书`,
	Run:         createCrt,
}

var (
	crtCA    string
	crtCAKey string
	ips      string
	dns      string
)

func init() {
	fs := flag.NewFlagSet("crt", flag.ContinueOnError)
	fs.StringVar(&caFileName, "fileName", "", "保存文件名")
	fs.StringVar(&caCommonName, "commonName", "", "常用名称")
	fs.StringVar(&caCountry, "country", "", "国家/地区")
	fs.StringVar(&caOrganization, "organization", "", "组织")
	fs.StringVar(&caOrganizationalUnit, "organizationalUnit", "", "机构机构")
	fs.StringVar(&caEmailAddress, "emailAddress", "", "邮件地址")
	fs.StringVar(&caProvince, "province", "", "省/市/自治区")
	fs.StringVar(&caLocality, "locality", "", "所在地")
	fs.StringVar(&crtCA, "ca", "", "根证书")
	fs.StringVar(&crtCAKey, "cakey", "", "根证书密钥")
	fs.IntVar(&caDays, "days", 0, "有效多少天")
	fs.IntVar(&caMonths, "months", 0, "有效多少月")
	fs.IntVar(&caYears, "years", 0, "有效多少年")
	fs.StringVar(&ips, "ip", "", "IP地址多个地址按逗号分隔")
	fs.StringVar(&dns, "dns", "", "DNS地址多个地址按逗号分隔")
	crtCmd.Flag = *fs
	commands.AvailableCommands = append(commands.AvailableCommands, crtCmd)
}

func createCrt(cmd *commands.Command, args []string) int {
	if err := cmd.Flag.Parse(args); err != nil {
		log.Fatalf("Error while parsing flags: %v", err.Error())
		return 1
	}
	if caFileName == "" {
		log.Printf("没有设置文件名,将使用默认值:server")
		caFileName = "server"
	}
	baseinfo := ssltool.CertInformation{
		CommonName: caCommonName,
		CrtName:    caFileName + ".crt",
		KeyName:    caFileName + ".key",
		IsCA:       false,
		Year:       caYears,
		Months:     caMonths,
		Days:       caDays,
	}
	if caCountry != "" {
		baseinfo.Country = []string{caCountry}
	}
	if caOrganization != "" {
		baseinfo.Organization = []string{caOrganization}
	}
	if caOrganizationalUnit != "" {
		baseinfo.OrganizationalUnit = []string{caOrganizationalUnit}
	}
	if caEmailAddress != "" {
		baseinfo.EmailAddress = []string{caEmailAddress}
	}
	if caProvince != "" {
		baseinfo.Province = []string{caProvince}
	}
	if caLocality != "" {
		baseinfo.Locality = []string{caLocality}
	}
	if ips != "" {
		ipps := strings.Split(ips, ",")
		baseinfo.IPAddresses = []net.IP{}
		for _, i := range ipps {
			baseinfo.IPAddresses = append(baseinfo.IPAddresses, net.ParseIP(i))
		}
	}
	if dns != "" {
		baseinfo.DNSNames = strings.Split(dns, ",")
	}
	var crt *x509.Certificate
	var pri *rsa.PrivateKey
	var err error
	if crtCA != "" && crtCAKey != "" {
		crt, pri, err = ssltool.Parse(crtCA, crtCAKey)
		if err != nil {
			log.Fatalf("解析根证书失败Error info:%s", err.Error())
			return 1
		}
	} else {
		log.Printf("无CA签发")
		crt = nil
		pri = nil
	}
	err = ssltool.CreateCRT(crt, pri, baseinfo)
	if err != nil {
		log.Fatalf("证书创建失败Error info:%s", err.Error())
		return 1
	}
	return 0
}
