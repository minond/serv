package serv

import (
	"fmt"
	LOG "log"
)

func Info(tmpl string, parts ...interface{}) {
	LOG.Printf(labeled("INFO", tmpl), parts...)
}

func Warn(tmpl string, parts ...interface{}) {
	LOG.Printf(labeled("WARN", tmpl), parts...)
}

func Fatal(tmpl string, parts ...interface{}) {
	LOG.Fatalf(labeled("FATAL", tmpl), parts...)
}

func labeled(label, msg string) string {
	return fmt.Sprintf("[%s] %s\n", label, msg)
}
