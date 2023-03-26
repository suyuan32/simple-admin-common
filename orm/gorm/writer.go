package gorm

import (
	"gorm.io/gorm/logger"

	"github.com/zeromicro/go-zero/core/logx"
)

type writer struct {
	logger.Writer
}

func (l writer) Printf(message string, data ...any) {
	logx.Errorf(message, data...)
}
