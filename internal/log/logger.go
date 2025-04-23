package log

import "go.uber.org/zap"

type Logger struct {
	*zap.Logger
}

type State int

const (
	Development State = iota + 1
	Production
)

func NewLogger(state State, options ...zap.Option) *Logger {
	switch state {
	case Development:
		return &Logger{
			Logger: zap.Must(zap.NewDevelopment(options...)),
		}
	case Production:
		return &Logger{
			Logger: zap.Must(zap.NewProduction(options...)),
		}
	default:
		return &Logger{
			Logger: zap.Must(zap.NewProduction(options...)),
		}
	}
}

func (lg *Logger) Close() error {
	return lg.Logger.Sync()
}
