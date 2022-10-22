package gh

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/gogf/gf/container/garray"
	"github.com/panjf2000/ants/v2"
)

const (
	HEAD        = "# FromGhosts Start"
	TAIL        = "# FromGhosts End"
	TIME        = "# UpdateTime: %s"
	LinePattern = "%s\t\t\t%s # %s"
)

var (
	REG  *regexp.Regexp = regexp.MustCompile(`# GitHub520 Host Start[\s\S]*# GitHub520 Host End`)
	FLAG string         = "# GitHub520 Host Start"
)

type FuncArg struct {
	IP  string
	Url string
}

type Ghosts struct {
	url       string
	hostList  *garray.StrArray
	ipReg     *regexp.Regexp
	urlReg    *regexp.Regexp
	hostsReg  *regexp.Regexp
	timeout   time.Duration // Second
	maxAvgRtt time.Duration // Millisecond
	pingCount int           // how many ping packets will be sent
	workerNum int           // Number of workers
	isWin     bool          // windows flag
}

func New(url string) *Ghosts {
	ipReg := `((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}`
	urlReg := `[a-zA-Z\.]{5,}`
	hostsReg := fmt.Sprintf(`%s[\s\S]*%s`, HEAD, TAIL)
	var isWin bool = false
	if strings.Contains(runtime.GOOS, "windows") {
		isWin = true
	}
	return &Ghosts{
		url:       url,
		hostList:  garray.NewStrArray(true),
		ipReg:     regexp.MustCompile(ipReg),
		urlReg:    regexp.MustCompile(urlReg),
		hostsReg:  regexp.MustCompile(hostsReg),
		timeout:   20 * time.Second,
		maxAvgRtt: 400 * time.Millisecond,
		pingCount: 10,
		workerNum: 100,
		isWin:     isWin,
	}
}

func (that *Ghosts) HostsFilePath() string {
	if that.isWin {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

func (that *Ghosts) BackupFilePath() string {
	return fmt.Sprintf("%s_backups", that.HostsFilePath())
}

func (that *Ghosts) GetHosts() {
	if that.url != "" {
		r, err := (&http.Client{Timeout: that.timeout}).Get(that.url)
		if err != nil {
			fmt.Println("Get hosts errored: ", err)
			return
		}
		byt, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}
		that.Parse(byt)
		defer r.Body.Close()
	}
}

func (that *Ghosts) Parse(resp []byte) {
	sc := bufio.NewScanner(strings.NewReader(string(resp)))
	var (
		wg   sync.WaitGroup
		err  error
		pool *ants.PoolWithFunc
	)
	pool, err = ants.NewPoolWithFunc(that.workerNum, func(arg interface{}) {
		defer wg.Done()
		a, ok := arg.(*FuncArg)
		if ok {
			that.PingHosts(a.IP, a.Url)
		}
	})
	if err != nil {
		panic(err)
	}
	for sc.Scan() {
		text := sc.Text()
		ipList := that.ipReg.FindAllString(text, -1)
		if len(ipList) == 1 {
			url := that.urlReg.FindString(text)
			if url == "" {
				continue
			}
			wg.Add(1)
			err = pool.Invoke(&FuncArg{
				IP:  ipList[0],
				Url: url,
			})
		}
	}
	wg.Wait()
	fmt.Println(that.hostList.String())
}

func (that *Ghosts) PingHosts(ip, url string) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		fmt.Println("Ping hosts errored: ", err)
		return
	}
	pinger.Count = that.pingCount
	if that.isWin {
		pinger.SetPrivileged(true)
	}
	pinger.Timeout = that.maxAvgRtt
	pinger.Run()
	statics := pinger.Statistics()
	if len(statics.Rtts) > 0 {
		line := fmt.Sprintf(LinePattern, ip, url, statics.AvgRtt)
		that.hostList.Append(line)
	}
	return
}

func (that *Ghosts) ReadAndBackup() (content []byte) {
	var (
		err  error
		file *os.File
	)
	file, err = os.Open(that.HostsFilePath())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	content, err = ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(that.BackupFilePath(), content, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (that *Ghosts) Gernerate() (newContent string) {
	that.GetHosts()
	if !that.hostList.IsEmpty() {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		oldContent := string(that.ReadAndBackup())
		if oldContent == "" {
			return
		}
		newHostStr := fmt.Sprintf("%s\n%s\n%s%s",
			HEAD,
			fmt.Sprintf(TIME, time.Now().In(loc).Format("2006-01-02 15:04:05")),
			that.hostList.Join("\n"),
			TAIL)
		if strings.Contains(oldContent, HEAD) {
			return that.hostsReg.ReplaceAllString(oldContent, newHostStr)
		} else if strings.Contains(oldContent, FLAG) {
			return REG.ReplaceAllString(oldContent, newHostStr)
		} else {
			return fmt.Sprintf("%s\n%s", oldContent, newHostStr)
		}
	}
	return
}

func (that *Ghosts) Run() {
	newStr := that.Gernerate()
	if newStr == "" {
		return
	}
	err := ioutil.WriteFile(that.HostsFilePath(), []byte(newStr), 0644)
	if err != nil {
		fmt.Println("Write file errored: ", err)
		return
	}
	fmt.Println("Successed!")
}

func Test(url string) {
	// g := New(url)
	// g.GetHosts()
	// loc, _ := time.LoadLocation("Asia/Shanghai")
	// fmt.Println(time.Now().In(loc).Format("2006-01-02 15:04:05"))
	// p, _ := os.Executable()
	// fmt.Println(filepath.Dir(p))
	u, _ := user.Current()
	fmt.Println(u.Username)
	fmt.Println(u.HomeDir)
}
