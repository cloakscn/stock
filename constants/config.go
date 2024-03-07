package constants

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"os/user"
	"path"
)

const (
	Name = "fyne-stock"
)

var (
	HomeDir    string
	ConfigPath string
	SQLDBPath  string
)

func init() {
	currentUser, err := user.Current()
	if err != nil {
		dir := os.Getenv("HOME")
		if dir == "" {
			fmt.Println("Can't get current user: %s", err.Error())
			return
		}
		HomeDir = dir
	} else {
		HomeDir = currentUser.HomeDir
	}

	dirPath := path.Join(HomeDir, ".config", Name)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			fmt.Println("Can't create config directory %s: %s", dirPath, err.Error())
			return
		}
	}

	ConfigPath = path.Join(dirPath, "config.ini")
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		fmt.Println("Can't find config, create a empty file")
		os.OpenFile(ConfigPath, os.O_CREATE|os.O_WRONLY, 0644)
	}

	SQLDBPath = path.Join(dirPath, "fyne-stock.db")
}

func GetConfig() (*ini.File, error) {
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		return nil, err
	}
	return ini.LoadSources(
		ini.LoadOptions{AllowBooleanKeys: true},
		ConfigPath,
	)
}
