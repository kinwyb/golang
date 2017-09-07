package main

import (
	"flag"
	"log"

	"git.oschina.net/kinwyb/golang/ssltool"

	"github.com/beego/bee/cmd/commands"
)

var caCmd = &commands.Command{
	CustomFlags: true,
	UsageLine:   "ca -filename=\"\" [-days=10] [-months=0] [-years=0] [-country=\"\"] [-commonName=\"\"] [-organization=\"\"] [-organizationalUnit=\"\"] [-emailAddress=\"\"] [-province=\"\"] [-locality=\"\"]",
	Short:       "创建一个根证书",
	Long:        `ca 用来创建一个ssl根证书`,
	Run:         createCert,
}

var (
	caCommonName         string
	caOrganization       string
	caOrganizationalUnit string
	caEmailAddress       string
	caProvince           string
	caLocality           string
	caFileName           string
	caCountry            string
	caDays               int
	caYears              int
	caMonths             int
)

func init() {
	fs := flag.NewFlagSet("ca", flag.ContinueOnError)
	fs.StringVar(&caFileName, "fileName", "", "保存文件名")
	fs.StringVar(&caCommonName, "commonName", "", "常用名称")
	fs.StringVar(&caCountry, "country", "", "国家/地区")
	fs.StringVar(&caOrganization, "organization", "", "组织")
	fs.StringVar(&caOrganizationalUnit, "organizationalUnit", "", "机构机构")
	fs.StringVar(&caEmailAddress, "emailAddress", "", "邮件地址")
	fs.StringVar(&caProvince, "province", "", "省/市/自治区")
	fs.StringVar(&caLocality, "locality", "", "所在地")
	fs.IntVar(&caDays, "days", 0, "有效多少天")
	fs.IntVar(&caMonths, "months", 0, "有效多少月")
	fs.IntVar(&caYears, "years", 0, "有效多少年")
	caCmd.Flag = *fs
	commands.AvailableCommands = append(commands.AvailableCommands, caCmd)
}

func createCert(cmd *commands.Command, args []string) int {
	if err := cmd.Flag.Parse(args); err != nil {
		log.Fatalf("Error while parsing flags: %v", err.Error())
		return 1
	}
	if caFileName == "" {
		log.Printf("没有设置文件名,将使用默认值:ca")
		caFileName = "ca"
	}
	baseinfo := ssltool.CertInformation{
		CommonName: caCommonName,
		CrtName:    caFileName + ".crt",
		KeyName:    caFileName + ".key",
		IsCA:       true,
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
	err := ssltool.CreateCRT(nil, nil, baseinfo)
	if err != nil {
		log.Fatalf("证书创建失败Error info:%s", err.Error())
		return 1
	}
	return 0
}
