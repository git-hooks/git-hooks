package main

import (
	"github.com/wsxiaoys/terminal/color"
	"os"
)

var logger = struct {
	Error    func(...interface{})
	Errors   func(int, ...interface{})
	Warn     func(...interface{})
	Info     func(...interface{})
	Errorln  func(...interface{})
	Errorsln func(int, ...interface{})
	Warnln   func(...interface{})
	Infoln   func(...interface{})
}{
	Error:    Error,
	Errors:   Errors,
	Warn:     Warn,
	Info:     Info,
	Errorln:  Errorln,
	Errorsln: Errorsln,
	Warnln:   Warnln,
	Infoln:   Infoln,
}

func Error(msgs ...interface{}) {
	Errors(1, msgs...)
}

func Errors(status int, msgs ...interface{}) {
	msgs = append([]interface{}{"@r"}, msgs...)
	color.Print(msgs...)
	os.Exit(status)
}
func Warn(msgs ...interface{}) {
	msgs = append([]interface{}{"@y"}, msgs...)
	color.Print(msgs...)
}
func Info(msgs ...interface{}) {
	color.Print(msgs...)
}
func Errorln(msgs ...interface{}) {
	Errorsln(1, msgs...)
}
func Errorsln(status int, msgs ...interface{}) {
	Errors(status, append(msgs, "\n")...)
}
func Warnln(msgs ...interface{}) {
	Warn(append(msgs, "\n")...)
}
func Infoln(msgs ...interface{}) {
	Info(append(msgs, "\n")...)
}
