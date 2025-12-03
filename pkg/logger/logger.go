package logger

import (
	"log"
	"os"
)

var isDev bool

func Init(dev bool) {
	isDev = dev
	if dev {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
}

// Info logs with green arrow
func Info(v ...any) {
	log.Println(append([]any{"[INFO] →"}, v...)...)
}

// Error logs with red X
func Error(v ...any) {
	log.Println(append([]any{"[ERROR]"}, v...)...)
}

// Debug only in dev
func Debug(v ...any) {
	if isDev {
		log.Println(append([]any{"[DEBUG]"}, v...)...)
	}
}

// Fatal logs and exits
func Fatal(v ...any) {
	log.Println(append([]any{"[FATAL]"}, v...)...)
	os.Exit(1)
}
