package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gogf/gf/os/genv"
	"github.com/moqsien/ghosts/pkgs/conf"
	"github.com/moqsien/ghosts/pkgs/gh"
	"github.com/moqsien/ghosts/pkgs/utils"
	"github.com/urfave/cli/v2"
)

// RunCli parses cmd and runs it.
func RunCli() {
	app := cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "urlnum",
				Aliases: []string{"u", "url", "unum"},
				Usage:   "choose source url.",
			},
		},
		Action: run,
		Commands: []*cli.Command{
			{
				Name:    "shell",
				Aliases: []string{"s", "sh"},
				Usage:   "open that shell.",
				Action: func(ctx *cli.Context) error {
					fmt.Println("open shell")
					return nil
				},
			},
			{
				Name:    "open",
				Aliases: []string{"o"},
				Usage:   "open hosts file.",
				Action: func(ctx *cli.Context) error {
					hostsPath := utils.GetHostsFilePath()
					utils.OpenFileWithEditor(hostsPath)
					return nil
				},
			},
			{
				Name:    "config",
				Aliases: []string{"c", "cnf", "conf"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Usage:   "show config file path.",
					},
					&cli.BoolFlag{
						Name:    "open",
						Aliases: []string{"o"},
						Usage:   "open config file.",
					},
				},
				Usage: "show config file content.",
				Action: func(ctx *cli.Context) error {
					if ctx.Bool("path") {
						cnf := &conf.GhConfig{}
						cnf.Load()
						fmt.Println(cnf.ConfigPath())
						return nil
					}
					if ctx.Bool("open") {
						cnf := &conf.GhConfig{}
						cnf.Load()
						utils.OpenFileWithEditor(cnf.ConfigPath())
						return nil
					}
					cnf := &conf.GhConfig{}
					cnf.ShowConfig()
					return nil
				},
			},
			{
				Name:    "erase",
				Aliases: []string{"e"},
				Usage:   "erase all customed hosts.",
				Action: func(ctx *cli.Context) error {
					ghs := gh.New()
					ghs.Run(true)
					return nil
				},
			},
			{
				Name:    "initconf",
				Aliases: []string{"i", "init"},
				Usage:   "init the config file.",
				Action: func(ctx *cli.Context) error {
					cnf := &conf.GhConfig{}
					cnf.Create()
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx *cli.Context) error {
	fmt.Println("Fetching hosts, please wait...")
	urlNum := ctx.Int("urlnum")
	// actions by default: update hosts file from your os.
	cnf := &conf.GhConfig{}
	cnf.Load()
	urlsToFetch := []string{}
	if urlNum > 0 && urlNum <= len(cnf.Conf.SourceUrls) {
		urlsToFetch = append(urlsToFetch, cnf.Conf.SourceUrls[urlNum-1])
	} else {
		urlsToFetch = append(urlsToFetch, cnf.Conf.SourceUrls...)
	}
	ghs := gh.New(urlsToFetch...)
	ghs.Run()
	// flush dns for windows
	if utils.IsWindows() {
		cmd := exec.Command("ipconfig")
		cmd.Args = []string{"ipconfig", "/flushdns"}
		cmd.Env = genv.All()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			fmt.Printf("Flush dns errored: %s", err.Error())
		}
	}
	return nil
}
