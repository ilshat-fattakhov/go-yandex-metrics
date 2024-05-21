package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func InitLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Can't initialize zap logger: %v", err)
	}
	//defer func() {
	//	if err := logger.Sync(); err != nil {
	//		log.Printf("failed to sync logger: %v", err)
	//		return
	//	}
	//}()
	return logger
}
