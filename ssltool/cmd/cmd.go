package main

//ssl证书生成

import (
	"flag"
	"log"
	"os"

	"github.com/beego/bee/cmd/commands"
	"github.com/beego/bee/config"
	"github.com/beego/bee/utils"
)

var usageTemplate = `ssl证书生成工具.
{{"USAGE" | headline}}
    {{"ssltool command [arguments]" | bold}}

{{"AVAILABLE COMMANDS" | headline}}
{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s" | bold}} {{.Short}}{{end}}{{end}}

Use {{"ssltool help [command]" | bold}} for more information about a command.
`

var helpTemplate = `{{"USAGE" | headline}}
  {{.UsageLine | printf "ssltool %s" | bold}}
{{if .Options}}{{endline}}{{"OPTIONS" | headline}}{{range $k,$v := .Options}}
  {{$k | printf "-%s" | bold}}
      {{$v}}
  {{end}}{{end}}
{{"DESCRIPTION" | headline}}
  {{tmpltostr .Long . | trim}}
`

var errorTemplate = `ssltool: %s.
Use {{"ssltool help" | bold}} for more information.
`

func usage() {
	utils.Tmpl(usageTemplate, commands.AvailableCommands)
}

func help(args []string) {
	if len(args) == 0 {
		usage()
	}
	if len(args) != 1 {
		utils.PrintErrorAndExit("Too many arguments", errorTemplate)
	}

	arg := args[0]

	for _, cmd := range commands.AvailableCommands {
		if cmd.Name() == arg {
			utils.Tmpl(helpTemplate, cmd)
			return
		}
	}
	utils.PrintErrorAndExit("Unknown help topic", errorTemplate)
}

func main() {

	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()

	if len(args) < 1 {
		usage()
		os.Exit(2)
		return
	}

	if args[0] == "help" {
		help(args[1:])
		return
	} else if len(args) < 2 {
		help(args[:])
		return
	}

	for _, c := range commands.AvailableCommands {
		if c.Name() == args[0] && c.Run != nil {
			c.Flag.Usage = func() { c.Usage() }
			if c.CustomFlags {
				args = args[1:]
			} else {
				c.Flag.Parse(args[1:])
				args = c.Flag.Args()
			}

			if c.PreRun != nil {
				c.PreRun(c, args)
			}

			config.LoadConfig()

			os.Exit(c.Run(c, args))
			return
		}
	}
	utils.PrintErrorAndExit("Unknown subcommand", errorTemplate)
}
