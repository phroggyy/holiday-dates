package logger

import "go.uber.org/zap"

var (
	Logger *zap.Logger
	SLog   *zap.SugaredLogger
)

func Initialize() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	Logger = l
	SLog = l.Sugar()

	return nil
}
