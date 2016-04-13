package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	logger   *log.Logger
	logOut   *os.File
	logLevel int
}

var (
	logger   Logger
	levelMap map[string]int = map[string]int{
		"DEBUG": 5,
		"INFO":  4,
		"WARN":  3,
		"ERROR": 2,
		"FATAL": 1,
	}
)

func init() {
	logger.logOut = os.Stdout
	logger.logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	logger.logLevel = 5
}

func SetLevel(level string) error {
	level = strings.ToUpper(level)
	if v, ok := levelMap[level]; ok {
		logger.logLevel = v
	} else {
		return errors.New(fmt.Sprintf("The log level %s is invalid", level))
	}
	return nil
}

func SetLogFile(file string) error {
	logfile, err := os.OpenFile(file, os.O_APPEND, 0644)
	if os.IsNotExist(err) {
		logfile, err = os.Create(file)
		if err != nil {
			return errors.New(fmt.Sprintf("Create log file %s error: %s\n", file, err.Error()))
		}
	} else if err != nil {
		return errors.New(fmt.Sprintf("Open log file %s error: %s\n", file, err.Error()))
	}
	logger.logOut = logfile
	logger.logger = log.New(logfile, "", log.LstdFlags)
	return nil
}

func Flush() {
	logger.logOut.Sync()
	logger.logOut.Close()
}

func Debug(v ...interface{}) {
	if logger.logLevel >= levelMap["DEBUG"] {
		logger.logger.Output(2, "DEBUG "+fmt.Sprintln(v...))
	}
}

func Debugf(format string, v ...interface{}) {
	if logger.logLevel >= levelMap["DEBUG"] {
		logger.logger.Output(2, "DEBUG "+fmt.Sprintf(format, v...))
	}
}

func Info(v ...interface{}) {
	if logger.logLevel >= levelMap["INFO"] {
		logger.logger.Output(2, "INFO "+fmt.Sprintln(v...))
	}
}

func Infof(format string, v ...interface{}) {
	if logger.logLevel >= levelMap["INFO"] {
		logger.logger.Output(2, "INFO "+fmt.Sprintf(format, v...))
	}
}

func Warn(v ...interface{}) {
	if logger.logLevel >= levelMap["WARN"] {
		logger.logger.Output(2, "WARN "+fmt.Sprintln(v...))
	}
}

func Warnf(format string, v ...interface{}) {
	if logger.logLevel >= levelMap["WARN"] {
		logger.logger.Output(2, "WARN "+fmt.Sprintf(format, v...))
	}
}

func Error(v ...interface{}) {
	if logger.logLevel >= levelMap["ERROR"] {
		logger.logger.Output(2, "ERROR "+fmt.Sprintln(v...))
	}
}

func Errorf(format string, v ...interface{}) {
	if logger.logLevel >= levelMap["ERROR"] {
		logger.logger.Output(2, "ERROR "+fmt.Sprintf(format, v...))
	}
}

func Fatal(v ...interface{}) {
	defer func() {
		Flush()
		os.Exit(1)
	}()
	logger.logger.Output(2, "FATAL "+fmt.Sprintln(v...))
}

func Fatalf(format string, v ...interface{}) {
	defer func() {
		Flush()
		os.Exit(1)
	}()
	logger.logger.Output(2, "FATAL "+fmt.Sprintf(format, v...))
}
