package mozilla

import (
	"fmt"
	"gomark/config"
	"gomark/database"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

const (
	ConfigName = "firefox"
)

var (
	// user mutable config
	Config = FirefoxDefaultConfig

	// Bookmark directory (including profile path)
	bookmarkDir string
)

// Config modifiable by user
type FirefoxConfig struct {
	// Default data source name query options for `places.sqlite` db
	PlacesDSN        database.DsnOptions
	WatchAllProfiles bool
	DefaultProfile   string
}

func (fc *FirefoxConfig) Set(opt string, v interface{}) error {
	//log.Debugf("setting option %s = %v", opt, v)
	s := structs.New(fc)
	f, ok := s.FieldOk(opt)
	if !ok {
		return fmt.Errorf("%s option not defined", opt)
	}

	return f.Set(v)
}

func (fc *FirefoxConfig) Get(opt string) (interface{}, error) {
	s := structs.New(fc)
	f, ok := s.FieldOk(opt)
	if !ok {
		return nil, fmt.Errorf("%s option not defined", opt)
	}

	return f.Value(), nil
}

func (fc *FirefoxConfig) Dump() map[string]interface{} {
	s := structs.New(fc)
	return s.Map()
}

func (fc *FirefoxConfig) String() string {
	s := structs.New(fc)
	return fmt.Sprintf("%v", s.Map())
}

func (fc *FirefoxConfig) MapFrom(src interface{}) error {
	return mapstructure.Decode(src, fc)
}

func SetBookmarkDir(dir string) {
	log.Debugf("setting bookmark dir to %s", dir)
	bookmarkDir = dir
}

func GetBookmarkDir() string {
	return bookmarkDir
}

func SetConfig(c *FirefoxConfig) {
	Config = c
}

func init() {
	config.RegisterConfigurator(ConfigName, Config)
}
