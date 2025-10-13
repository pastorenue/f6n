package logger

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	file, err := os.OpenFile("f6n-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file")
	}
	Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}
