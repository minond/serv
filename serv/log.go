package serv

import (
	"fmt"
	"log"
)

// Info prints a message with an INFO prefix. See log.Printf.
func Info(tmpl string, parts ...interface{}) {
	log.Printf(labeled("INFO", tmpl), parts...)
}

// Warn prints a message with an WARN prefix. See log.Printf.
func Warn(tmpl string, parts ...interface{}) {
	log.Printf(labeled("WARN", tmpl), parts...)
}

// Fatal prints a message with an FATAL prefix. See log.Fatalf.
func Fatal(tmpl string, parts ...interface{}) {
	log.Fatalf(labeled("FATAL", tmpl), parts...)
}

func labeled(label, msg string) string {
	return fmt.Sprintf("[%s] %s\n", label, msg)
}
