package log

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"io"
	"os"
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
	//if debug == true {
	//	logger.SetLevel(logrus.DebugLevel)
	//} else if verbose == true {
	//	logger.SetOutput(os.Stdout)
	//	logger.SetLevel(logrus.InfoLevel)
	//}
}

func SetOutput(output io.Writer) {
	logger.SetOutput(output)
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

// GoodF print good message
func Green(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	good := color.GreenString("[+]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

func Yellow(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	good := color.YellowString("[!]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

func Blue(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	good := color.BlueString("[*]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

func IsVul(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	good := color.RedString("[HACKED]")
	fmt.Printf("%s %s\n", good, fmt.Sprintf(format, args...))
}

// InforF print info message
func InforF(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Info(fmt.Sprintf(format, args...))
}

func Info(args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Infoln(args)
}

// ErrorF print good message
func ErrorF(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Error(fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Errorln(args)
}

func WarningF(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Warningf(fmt.Sprintf(format, args...))
}

func Warning(args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Warningln(args)
}

// DebugF print debug message
func DebugF(format string, args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Debug(fmt.Sprintf(format, args...))
}

func Debug(args ...interface{}) {
	//if progress.Bar != nil {
	//	_ = progress.Bar.Clear()
	//	defer progress.Bar.RenderBlank()
	//}
	logger.Debugln(args)
}
