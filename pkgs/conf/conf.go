package conf

import (
	"os/user"
	"path"
)

const (
	FileName = "/.ghosts/ghosts.yaml"
)

type GhConfig struct {
	filepath string
}

func (that *GhConfig) Path() string {
	if that.filepath != "" {
		return that.filepath
	}
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(u.HomeDir, FileName)
}

func (that *GhConfig) Create() {}

func (that *GhConfig) Load() {}

func (that *GhConfig) Set(key, value string) {}

func (that *GhConfig) save() {}
