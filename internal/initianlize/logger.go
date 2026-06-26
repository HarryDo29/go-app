package initianlize

import (
	"go-app/global"
	"go-app/pkg/logger"
)

func InitLogger() {
	global.Logger = logger.NewLogger(global.Config.Logger)
}
