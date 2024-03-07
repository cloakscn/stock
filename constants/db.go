package constants

import (
	"github.com/cloakscn/fyne-stock/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func init() {
	err := initDatabase()
	if err != nil {
		panic(err)
	}
}

func DB() *gorm.DB {
	return db
}

func initDatabase() (err error) {
	db, err = gorm.Open(sqlite.Open(SQLDBPath), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&model.Code{})
	if err != nil {
		return err
	}

	return nil
}
