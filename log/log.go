package log

import (
	"log"
	"os"
)

type Logger struct {
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	filename      string
}

var Log *Logger

func NewLog() {
	Log = createLogger("./log.txt")
}

func GetInstance() *Logger {
	return Log
}

func createLogger(fname string) *Logger {
	file, _ := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)

	return &Logger{
		filename:      fname,
		InfoLogger:    log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		ErrorLogger:   log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		WarningLogger: log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}
