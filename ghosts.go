package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	HelloGithubUrl  = "https://raw.hellogithub.com/hosts"
	HostsPath       = "/etc/hosts"
	HostsBackupPath = "/etc/hosts_backups"
	SudoPath        = "/usr/bin/sudo"
	Flag            = "# GitHub520 Host Start"
)

var Reg *regexp.Regexp = regexp.MustCompile(`# GitHub520 Host Start[\s\S]*# GitHub520 Host End`)

func GetHelloGithub() string {
	r, err := (&http.Client{}).Get(HelloGithubUrl)
	if err != nil {
		return ""
	}
	byt, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	defer r.Body.Close()
	return string(byt)
}

func ReadAndBackup() string {
	file, err := os.Open(HostsPath)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	err = ioutil.WriteFile(HostsBackupPath, content, 0644)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(content)
}

func GenerateNew(old, newhello string) (result string) {
	if strings.Contains(old, Flag) {
		result = Reg.ReplaceAllString(old, newhello)
	} else {
		result = fmt.Sprintf("%s\n%s", old, newhello)
	}
	return
}

func main() {
	if len(os.Args) == 1 {
		exePath, _ := os.Executable()
		cmd := exec.Command(SudoPath, exePath, "1")
		var stdOut, stdErr bytes.Buffer
		cmd.Stderr = &stdErr
		cmd.Stdout = &stdOut
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Got error:%s, msg:%s", err, stdErr.String())
		}
		fmt.Println("success:", stdOut.String())
	} else {
		newhello := GetHelloGithub()
		if !strings.Contains(newhello, Flag) {
			return
		}
		old := ReadAndBackup()
		if len(old) == 0 {
			return
		}
		result := GenerateNew(old, newhello)
		if result != "" && result != old {
			err := ioutil.WriteFile(HostsPath, []byte(result), 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		fmt.Println("successed!")
	}
}
