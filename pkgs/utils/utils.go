package utils

import (
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	IsWin   bool
	winOnce *sync.Once = &sync.Once{}
)

const (
	HostsPahtWin = `C:\Windows\System32\drivers\etc\hosts`
	HostsPath    = "/etc/hosts"
)

func GetHostsFilePath() string {
	if IsWindows() {
		return HostsPahtWin
	}
	return HostsPath
}

func IsWindows() bool {
	winOnce.Do(func() {
		if strings.Contains(runtime.GOOS, "windows") {
			IsWin = true
		}
	})
	return IsWin
}

func PahtIsExist(path string) (bool, error) {
	_, _err := os.Stat(path)
	if _err == nil {
		return true, nil
	}
	if os.IsNotExist(_err) {
		return false, nil
	}
	return false, _err
}

func VerifyUrl(rawUrl string) (r bool) {
	r = true
	_, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		r = false
		return
	}
	url, err := url.Parse(rawUrl)
	if err != nil || url.Scheme == "" || url.Host == "" {
		r = false
		return
	}
	return
}

func InTest(list []string, str string) (r bool) {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return
}
