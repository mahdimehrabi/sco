package zerolog

import "github.com/rs/zerolog/log"

type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l Logger) Error(err error) {
	log.Err(err).Send()
}

func (l Logger) Warning(msg string) {
	warn := log.Warn()
	warn.Msg(msg)
}

func (l Logger) Info(msg string) {
	warn := log.Info()
	warn.Msg(msg)
}
