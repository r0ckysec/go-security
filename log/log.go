package log

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"io"
	"os"
	"path"
	"runtime"
)

var logger *logrus.Logger

func init() {
	logger = &logrus.Logger{
		Out:   os.Stdout,
		Level: logrus.ErrorLevel,
		Formatter: &prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}
}

func GetLogger() *logrus.Logger {
	return logger
}

func SetOutput(logPath string, isStdout bool) {
	//set logfile Stdout
	if _, err := os.Stat(path.Dir(logPath)); os.IsNotExist(err) {
		_ = os.MkdirAll(path.Dir(logPath), os.ModePerm)
	}
	logFile, logErr := os.OpenFile(logPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		FatalF("Fail to find %s Server start Failed", logFile)
	}
	writers := []io.Writer{
		logFile,
	}
	if isStdout {
		writers = append(writers, os.Stdout)
	}
	writer := io.MultiWriter(writers...)
	logger.SetOutput(writer)
}

func SetDebug() {
	logger.SetLevel(logrus.DebugLevel)
}

func SetVerbose() {
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
}

func SetLevel(level logrus.Level) {
	logger.SetLevel(level)
}

func Log(format string, args ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(format, args...))
}

// Green print good message
func Green(format string, args ...interface{}) {
	good := color.GreenString("[+]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

func Yellow(format string, args ...interface{}) {
	good := color.YellowString("[!]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

func Blue(format string, args ...interface{}) {
	good := color.BlueString("[*]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

func Attack(format string, args ...interface{}) {
	good := color.RedString("[ATTACK!]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

// InforF print info message
func InforF(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	logger.Infoln(args)
}

// ErrorF print good message
func ErrorF(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLine := fmt.Sprintf("%s:%d", file, line)
	logger.Errorln(fileLine, fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLine := fmt.Sprintf("%s:%d", file, line)
	logger.Errorln(fileLine, args)
}

func FatalF(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLine := fmt.Sprintf("%s:%d", file, line)
	logger.Fatalln(fileLine, fmt.Sprintf(format, args...))
}

func Fatal(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fileLine := fmt.Sprintf("%s:%d", file, line)
	logger.Fatalln(fileLine, args)
}

func WarningF(format string, args ...interface{}) {
	logger.Warningf(fmt.Sprintf(format, args...))
}

func Warning(args ...interface{}) {
	logger.Warningln(args)
}

// DebugF print debug message
func DebugF(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}

func Debug(args ...interface{}) {
	logger.Debugln(args)
}
