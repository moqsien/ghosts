package conf

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/moqsien/ghosts/pkgs/utils"
	"gopkg.in/yaml.v3"
)

const (
	fileName = ".ghosts/ghosts.yml"
)

type Conf struct {
	SourceUrls  []string `yaml:"sourceUrls"`
	HostFilters []string `yaml:"hostFilters"`
	ReqTimeout  int      `yaml:"reqTimeout"`
	MaxAvgRtt   int      `yaml:"maxAvgRtt"`
	PingCount   int      `yaml:"pingCount"`
	WorkerNum   int      `yaml:"workerNum"`
}

var defaultConf = &Conf{
	SourceUrls: []string{
		"https://www.foul.trade:3000/Johy/Hosts/raw/branch/main/hosts.txt",
		"https://gitlab.com/ineo6/hosts/-/raw/master/next-hosts",
		"https://raw.hellogithub.com/hosts",
	},
	HostFilters: []string{
		"github",
	},
	ReqTimeout: 30,
	MaxAvgRtt:  400,
	PingCount:  10,
	WorkerNum:  100,
}

type GhConfig struct {
	path      string // config file path
	Conf      *Conf
	fieldList map[string]string
}

func (that *GhConfig) Path() (p string, err error) {
	if that.path != "" {
		p = that.path
		return
	}
	var u *user.User
	u, err = user.Current()
	if err != nil {
		return
	}
	p = filepath.Join(u.HomeDir, fileName)
	dir := filepath.Dir(p)
	if ok, _ := utils.PahtIsExist(dir); !ok {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			p = ""
			return
		}
		that.path = p
	}
	return
}

func (that *GhConfig) Create() {
	that.Conf = defaultConf
	that.save()
}

func (that *GhConfig) Load() {
	if p, err := that.Path(); err == nil {
		if ok, _ := utils.PahtIsExist(p); ok {
			file, err := os.OpenFile(p, os.O_CREATE|os.O_RDONLY, 0666)
			if err != nil {
				fmt.Println("Open config file: ", that.path, " errored: ", err)
				return
			}
			defer file.Close()
			decoder := yaml.NewDecoder(file)
			if that.Conf == nil {
				that.Conf = &Conf{}
			}
			err = decoder.Decode(that.Conf)
			if err != nil {
				fmt.Println("Decode config file errored: ", err, ", New a default one!")
				that.Create()
				return
			}
		} else {
			that.Conf = defaultConf
			that.Create()
		}
	}
}

func (that *GhConfig) parseFiled() {
	if that.fieldList == nil {
		that.fieldList = make(map[string]string)
	}
	val := reflect.ValueOf(defaultConf)
	valType := val.Type().Elem()
	var name string
	for i := 0; i < valType.NumField(); i++ {
		name = valType.Field(i).Name
		that.fieldList[strings.ToLower(name)] = name
	}
}

func (that *GhConfig) Set(key, value string) {
	that.Load()
	if that.fieldList == nil || len(that.fieldList) == 0 {
		that.parseFiled()
	}
	if fName, found := that.fieldList[key]; found {
		val := reflect.ValueOf(that.Conf)
		field := val.Elem().FieldByName(fName)
		var changed bool = false
		if field.Type().Kind() == reflect.String {
			if strings.Contains(key, "url") && !utils.VerifyUrl(value) {
				fmt.Println("Illegal url value: ", value)
				return
			}
			if field.String() != value {
				field.SetString(value)
				changed = true
			}
		} else if field.CanInt() {
			if v, err := strconv.Atoi(value); err == nil {
				if field.Int() != int64(v) {
					field.SetInt(int64(v))
					changed = true
				}
			}
		} else {
			return
		}
		if changed {
			that.save()
		}
	} else {
		fmt.Println("Field: ", key, ", not found!")
	}
}

func (that *GhConfig) AddUrls(urls ...string) {
	that.Load()
	var changed bool
	for _, v := range urls {
		if utils.VerifyUrl(v) && !utils.InTest(that.Conf.SourceUrls, v) {
			that.Conf.SourceUrls = append(that.Conf.SourceUrls, v)
			changed = true
		} else {
			fmt.Println("Invalid url or url already exists: ", v)
		}
	}
	if changed {
		that.save()
	}
}

func (that *GhConfig) RemoveUrls(idx int) {
	that.Load()
	var changed bool
	if idx < len(that.Conf.SourceUrls) && idx >= 0 {
		next := idx + 1
		var tail []string
		if next < len(that.Conf.SourceUrls) {
			tail = that.Conf.SourceUrls[next:]
		}
		that.Conf.SourceUrls = append(that.Conf.SourceUrls[:idx], tail...)
		changed = true
	}
	if changed {
		that.save()
	}
}

func (that *GhConfig) ShowConfig() {
	if p, err := that.Path(); err == nil {
		if ok, _ := utils.PahtIsExist(p); ok {
			content, err := ioutil.ReadFile(p)
			if err != nil {
				fmt.Println("Read config file failed!")
				return
			}
			fmt.Println("========================")
			fmt.Println(string(content))
			fmt.Println("========================")
		} else {
			fmt.Println("Config file not found!")
		}
	}
}

func (that *GhConfig) ConfigPath() string {
	if that.path == "" {
		that.path, _ = that.Path()
	}
	return that.path
}

func (that *GhConfig) save() {
	if p, err := that.Path(); err == nil && that.Conf != nil {
		file, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println("Open config file: ", that.path, " errored: ", err)
			return
		}
		defer file.Close()
		encoder := yaml.NewEncoder(file)
		err = encoder.Encode(that.Conf)
		if err != nil {
			fmt.Println("Write config file failed: ", err)
			return
		}
		defer encoder.Close()
	} else {
		fmt.Println("Save config file errored!")
	}
}
