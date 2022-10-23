package gh

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/gogf/gf/container/garray"
	"github.com/moqsien/ghosts/pkgs/conf"
	"github.com/moqsien/ghosts/pkgs/utils"
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
	once        *sync.Once
	sourceUrls  []string
	hostList    *garray.StrArray
	ipReg       *regexp.Regexp
	urlReg      *regexp.Regexp
	hostsReg    *regexp.Regexp
	timeout     time.Duration // Second
	maxAvgRtt   time.Duration // Millisecond
	pingCount   int           // how many ping packets will be sent
	workerNum   int           // Number of workers
	hostFilters []string      // filters that allows only specific hosts to be added
}

func New(urls ...string) *Ghosts {
	ipReg := `((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}`
	urlReg := `[a-zA-Z0-9\.]{5,}`
	hostsReg := fmt.Sprintf(`%s[\s\S]*%s`, HEAD, TAIL)
	cnf := conf.GhConfig{}
	cnf.Load()
	return &Ghosts{
		once:        &sync.Once{},
		sourceUrls:  urls,
		hostList:    garray.NewStrArray(true),
		ipReg:       regexp.MustCompile(ipReg),
		urlReg:      regexp.MustCompile(urlReg),
		hostsReg:    regexp.MustCompile(hostsReg),
		timeout:     time.Duration(cnf.Conf.ReqTimeout) * time.Second,
		maxAvgRtt:   time.Duration(cnf.Conf.MaxAvgRtt) * time.Millisecond,
		pingCount:   cnf.Conf.PingCount,
		workerNum:   cnf.Conf.WorkerNum,
		hostFilters: cnf.Conf.HostFilters,
	}
}

func (that *Ghosts) HostsFilePath() string {
	return utils.GetHostsFilePath()
}

func (that *Ghosts) BackupFilePath() string {
	return fmt.Sprintf("%s_backups", that.HostsFilePath())
}

func (that *Ghosts) GetHosts(url string) {
	if url != "" {
		r, err := (&http.Client{Timeout: that.timeout}).Get(url)
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

func (that *Ghosts) extractHostUrl(text, ip string) string {
	raw := strings.Replace(text, ip, "", -1)
	return strings.TrimSpace(raw)
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
	defer pool.Release()
	for sc.Scan() {
		text := sc.Text()
		ipList := that.ipReg.FindAllString(text, -1)
		if len(ipList) == 1 {
			url := that.extractHostUrl(text, ipList[0])
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
}

func (that *Ghosts) toSave(url string) bool {
	if len(that.hostFilters) == 0 {
		return true
	}
	for _, filter := range that.hostFilters {
		if strings.Contains(url, filter) {
			return true
		}
	}
	return false
}

func (that *Ghosts) PingHosts(ip, url string) {
	if !that.toSave(url) {
		return
	}
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		fmt.Println("Ping hosts errored: ", err)
		return
	}
	pinger.Count = 10
	if utils.IsWindows() {
		pinger.SetPrivileged(true)
	}
	pinger.Timeout = 400 * time.Millisecond
	err = pinger.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
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

func (that *Ghosts) replace(oldContent, newHostStr string) string {
	if strings.Contains(oldContent, HEAD) {
		return that.hostsReg.ReplaceAllString(oldContent, newHostStr)
	} else if strings.Contains(oldContent, FLAG) {
		return REG.ReplaceAllString(oldContent, newHostStr)
	} else {
		if newHostStr != "" {
			return fmt.Sprintf("%s\n%s", oldContent, newHostStr)
		}
		return oldContent
	}
}

func (that *Ghosts) Gernerate(toclear bool) (newContent string) {
	if toclear {
		oldContent := string(that.ReadAndBackup())
		if oldContent == "" {
			return
		}
		return that.replace(oldContent, "")
	}
	for _, url := range that.sourceUrls {
		that.GetHosts(url)
	}
	if !that.hostList.IsEmpty() {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		oldContent := string(that.ReadAndBackup())
		if oldContent == "" {
			return
		}
		newHostStr := fmt.Sprintf("%s\n%s\n%s\n%s",
			HEAD,
			fmt.Sprintf(TIME, time.Now().In(loc).Format("2006-01-02 15:04:05")),
			that.hostList.Join("\n"),
			TAIL)
		return that.replace(oldContent, newHostStr)
	}
	return
}

func (that *Ghosts) Run(toclear ...bool) {
	var clear bool
	if len(toclear) > 0 && toclear[0] {
		clear = true
	}
	newStr := that.Gernerate(clear)
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
