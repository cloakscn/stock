package constants

import (
	"github.com/cloakscn/fyne-stock/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

var (
	executablePath string
	db             *gorm.DB
)

func init() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	executablePath = filepath.Dir(executable)

	err = initDatabase()
	if err != nil {
		panic(err)
	}
}

func DB() *gorm.DB {
	return db
}

func initDatabase() (err error) {
	db, err = gorm.Open(sqlite.Open(filepath.Join(executablePath, "cache.db")), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&model.Code{})
	if err != nil {
		return err
	}

	return nil
}
