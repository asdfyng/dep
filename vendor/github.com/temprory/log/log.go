package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	LEVEL_DEBUG = iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_PANIC
	LEVEL_FATAL
	LEVEL_NONE
)

var (
	logsep   = ""
	inittime = time.Now()

	DefaultLogLevel      = LEVEL_DEBUG
	DefaultLogDepth      = 2
	DefaultLogWriter     = os.Stdout
	DefaultLogTimeLayout = "2006-01-02 15:04:05.000"

	DefaultLogger = NewLogger()
)

func init() {
	DefaultLogger.depth = DefaultLogDepth + 1
}

type Logger struct {
	Writer   io.Writer
	depth    int
	Level    int
	Layout   string
	Formater func(lvl int, format string, v ...interface{}) string
}

func (logger *Logger) Printf(fmtstr string, v ...interface{}) {
	fmt.Fprintf(logger.Writer, fmtstr, v...)
}

func (logger *Logger) Println(v ...interface{}) {
	fmt.Fprintln(logger.Writer, v...)
}

func (logger *Logger) Debug(format string, v ...interface{}) {
	if LEVEL_DEBUG >= logger.Level {
		fmt.Fprintln(logger.Writer, logger.Formater(LEVEL_DEBUG, format, v...))
	}
}

func (logger *Logger) Info(format string, v ...interface{}) {
	if LEVEL_INFO >= logger.Level {
		fmt.Fprintln(logger.Writer, logger.Formater(LEVEL_INFO, format, v...))
	}
}

func (logger *Logger) Warn(format string, v ...interface{}) {
	if LEVEL_WARN >= logger.Level {
		fmt.Fprintln(logger.Writer, logger.Formater(LEVEL_WARN, format, v...))
	}
}

func (logger *Logger) Error(format string, v ...interface{}) {
	if LEVEL_ERROR >= logger.Level {
		fmt.Fprintln(logger.Writer, logger.Formater(LEVEL_ERROR, format, v...))
	}
}

func (logger *Logger) Panic(format string, v ...interface{}) {
	if LEVEL_PANIC >= logger.Level {
		s := logger.Formater(LEVEL_PANIC, format, v...)
		fmt.Fprintln(logger.Writer, s)
		panic(errors.New(s))
	}
}

func (logger *Logger) Fatal(format string, v ...interface{}) {
	if LEVEL_FATAL >= logger.Level {
		fmt.Fprintln(logger.Writer, logger.Formater(LEVEL_FATAL, format, v...))
		os.Exit(-1)
	}
}

func (logger *Logger) SetLevel(level int) {
	if level >= 0 && level <= LEVEL_NONE {
		logger.Level = level
	} else {
		log.Fatal(fmt.Errorf("log SetLogLevel Error: Invalid Level - %d\n", level))
	}
}

func (logger *Logger) SetOutput(out io.Writer) {
	logger.Writer = out
}

func (logger *Logger) SetFormater(f func(lvl int, format string, v ...interface{}) string) {
	logger.Formater = f
}

func (logger *Logger) defaultLogFormater(lvl int, format string, v ...interface{}) string {
	now := time.Now()
	_, file, line, ok := runtime.Caller(logger.depth)
	if !ok {
		file = "???"
		line = -1
	} else {
		pos := strings.LastIndex(file, "/")
		if pos >= 0 {
			file = file[pos+1:]
		}
	}

	switch lvl {
	case LEVEL_DEBUG:
		return strings.Join([]string{now.Format(logger.Layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_INFO:
		return strings.Join([]string{now.Format(logger.Layout), fmt.Sprintf(" [ Info] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_WARN:
		return strings.Join([]string{now.Format(logger.Layout), fmt.Sprintf(" [ Warn] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_ERROR:
		return strings.Join([]string{now.Format(logger.Layout), fmt.Sprintf(" [Error] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_PANIC:
		return strings.Join([]string{now.Format(logger.Layout), fmt.Sprintf(" [Panic] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_FATAL:
		return strings.Join([]string{now.Format(logger.Layout), fmt.Sprintf(" [Fatal] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	default:
	}
	return ""
}

func (logger *Logger) SetLogTimeFormat(layout string) {
	logger.Layout = layout
}

/********* default logger *********/
func Printf(fmtstr string, v ...interface{}) {
	DefaultLogger.Printf(fmtstr, v...)
}

func Println(v ...interface{}) {
	DefaultLogger.Println(v...)
}

func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	DefaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}

func Panic(format string, v ...interface{}) {
	DefaultLogger.Panic(format, v...)
}

func Fatal(format string, v ...interface{}) {
	DefaultLogger.Fatal(format, v...)
}

func SetLevel(level int) {
	DefaultLogger.SetLevel(level)
}

func SetOutput(out io.Writer) {
	DefaultLogger.SetOutput(out)
}

func SetFormater(f func(lvl int, format string, v ...interface{}) string) {
	DefaultLogger.SetFormater(f)
}

func SetLogTimeFormat(layout string) {
	DefaultLogger.SetLogTimeFormat(layout)
}

func LogWithFormater(lvl int, depth int, layout string, format string, v ...interface{}) string {
	now := time.Now()
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = -1
	} else {
		pos := strings.LastIndex(file, "/")
		if pos >= 0 {
			file = file[pos+1:]
		}
	}

	switch lvl {
	case LEVEL_DEBUG:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Debug] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_INFO:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [ Info] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_WARN:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [ Warn] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_ERROR:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Error] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_PANIC:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Panic] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	case LEVEL_FATAL:
		return strings.Join([]string{now.Format(layout), fmt.Sprintf(" [Fatal] [%s:%d] ", file, line), fmt.Sprintf(format, v...)}, "")
	default:
	}
	return ""
}

func NewLogger() *Logger {
	logger := &Logger{
		Level:  DefaultLogLevel,
		depth:  DefaultLogDepth,
		Writer: DefaultLogWriter,
		Layout: DefaultLogTimeLayout,
	}
	logger.Formater = logger.defaultLogFormater
	return logger
}
