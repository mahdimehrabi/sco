package logger

type Logger interface {
	Error(err error)
	Warning(message string)
	Info(message string)
}
