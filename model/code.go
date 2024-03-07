package model

import "gorm.io/gorm"

type Code struct {
	gorm.Model
	Code   string
	Stage  string
	Remark string
}
