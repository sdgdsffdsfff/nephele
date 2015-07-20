package data

import (
	"github.com/Unknwon/goconfig"
	"strconv"
	"strings"
)

var lock chan int = make(chan int, 1)

var instance *goconfig.ConfigFile

func init() {
	if instance != nil {
		return //nil
	}
	lock <- 1
	if instance == nil {
		instance, _ = goconfig.LoadConfigFile("../conf/conf.ini")
		return //err
	}
	<-lock
	return //nil
}

//key: nfs1,nfs2

func GetFdfsDomain() (string, error) {
	return getValue("", "fdfsdomain")
}

func GetFdfsPort() int {
	return mustInt("", "fdfsport", 22122)
}

func GetDirPath(channel, storagetype string) (string, error) {
	//storagetype: nfs1,nfs2
	return getValue(channel, storagetype)
}

func GetResizeTypes(channel string) (string, error) {
	return getValue(channel, "resizetypes")
}
func GetSizes(channel string) (string, error) {
	return getValue(channel, "sizes")
}

func GetRotates(channel string) (string, error) {
	return getValue(channel, "rotates")
}

func GetQuality(channel string) (string, error) {
	return getValue(channel, "quality")
}
func GetQualities(channel string) (string, error) {
	return getValue(channel, "qualities")
}
func GetLogodir(channel string) (string, error) {
	return getValue(channel, "logodir")
}
func IsEnableNameLogo(channel string) (bool, error) {
	isenable, err := getValue(channel, "isenablenamelogo")
	if err != nil {
		return false, err
	}
	if isenable == "1" {
		return true, nil
	} else {
		return false, nil
	}
}
func GetDefaultLogo(channel string) (string, error) {
	return getValue(channel, "defaultlogo")
}
func GetLogoNames(channel string) (string, error) {
	return getValue(channel, "logonames")
}
func GetImagelesswidthForLogo(channel string) (int64, error) {
	width, err := getValue(channel, "imagelesswidthforlogo")
	if err != nil {
		return 0, err
	}
	if width == "" {
		return 0, nil
	}
	return strconv.ParseInt(width, 10, 64)
}
func GetImagelessheightForLogo(channel string) (int64, error) {
	height, err := getValue(channel, "imagelessheightforlogo")
	if err != nil {
		return 0, err
	}
	if height == "" {
		return 0, nil
	}
	return strconv.ParseInt(height, 10, 64)
}

func GetDissolves(channel string) (string, error) {
	return getValue(channel, "dissolves")
}
func GetDissolve(channel string) int {
	return mustInt(channel, "dissolve", 100)
}

func GetNamelogoDissolve(channel string) int {
	i, err := instance.Int(channel, "namelogodissolve")
	if err != nil {
		return 0
	} else {
		return i
	}
}
func GetSequenceofoperation(channel string) ([]string, error) {
	v, err := getValue(channel, "sequenceofoperation")
	if err != nil {
		return nil, err
	}
	if v == "" {
		v = "s,resize,q,m,rotate"
	}
	return strings.Split(v, ","), nil
}

func Reload() error {
	return instance.Reload()
}

func getValue(channel string, key string) (string, error) {
	v, _ := instance.GetValue(channel, key)
	if v == "" && channel != "" {
		v, _ = instance.GetValue("", key)
		return v, nil
	}
	if v == "nil" {
		v = ""
	}
	return v, nil
}

func mustInt(channel string, key string, defaultvalue int) int {
	//if err := loadConfiguration(); err != nil {
	//	return defaultvalue
	//}
	return instance.MustInt(channel, key, defaultvalue)
}
