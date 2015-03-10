package main

import (
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"os"
)

type Logger struct {
	errors []interface{}
	infos  []interface{}
	warns  []interface{}
}

func (logger *Logger) Error(msgs ...interface{}) {
	logger.Errors(1, msgs...)
}

func (logger *Logger) Errors(status int, msgs ...interface{}) {
	if isTestEnv() {
		logger.errors = append(logger.errors, msgs...)
		return
	}

	msgs = append([]interface{}{"@r"}, msgs...)
	color.Print(msgs...)
	os.Exit(status)
}

func (logger *Logger) Warn(msgs ...interface{}) {
	if isTestEnv() {
		logger.warns = append(logger.warns, msgs...)
		return
	}

	msgs = append([]interface{}{"@y"}, msgs...)
	color.Print(msgs...)
}

func (logger *Logger) Info(msgs ...interface{}) {
	if isTestEnv() {
		logger.infos = append(logger.infos, msgs...)
		return
	}

	color.Print(msgs...)
}

func (logger *Logger) Errorln(msgs ...interface{}) {
	logger.Errorsln(1, msgs...)
}

func (logger *Logger) Errorsln(status int, msgs ...interface{}) {
	logger.Errors(status, append(msgs, "\n")...)
}

func (logger *Logger) Warnln(msgs ...interface{}) {
	logger.Warn(append(msgs, "\n")...)
}

func (logger *Logger) Infoln(msgs ...interface{}) {
	logger.Info(append(msgs, "\n")...)
}

func (logger *Logger) clear() {
	logger.infos = logger.infos[:0]
	logger.warns = logger.warns[:0]
	logger.errors = logger.errors[:0]
}

func (logger *Logger) flush() {
	fmt.Println(logger.infos)
	fmt.Println(logger.warns)
	fmt.Println(logger.errors)
	logger.clear()
}

var logger = new(Logger)
