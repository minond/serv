package main

import (
	"fmt"
	"log"
)

func info(tmpl string, parts ...interface{}) {
	log.Printf(labeled("INFO", tmpl), parts...)
}

func warn(tmpl string, parts ...interface{}) {
	log.Printf(labeled("WARN", tmpl), parts...)
}

func fatal(tmpl string, parts ...interface{}) {
	log.Fatalf(labeled("FATAL", tmpl), parts...)
}

func labeled(label, msg string) string {
	return fmt.Sprintf("[%s] %s\n", label, msg)
}
