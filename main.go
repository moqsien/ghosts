package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-ping/ping"
	"github.com/gogf/gf/os/genv"

	command "github.com/moqsien/ghosts/pkgs/cmd"
	"github.com/moqsien/ghosts/pkgs/utils"
)

func testPing() bool {
	pinger, _ := ping.NewPinger("www.baidu.com")
	pinger.Timeout = 400 * time.Millisecond
	err := pinger.Run()
	if err != nil {
		return false
	}
	return true
}

func main() {
	// sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
	if utils.IsWindows() {
		command.RunCli()
	} else {
		if utils.IsLinux() && !testPing() {
			cmd := exec.Command("sudo", "sysctl", "-w", `net.ipv4.ping_group_range=0 2147483647`)
			cmd.Env = genv.All()
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				fmt.Printf("Call sysctl errored: %s\n", err.Error())
				return
			}
		}
		command.RunCli()
	}
}
