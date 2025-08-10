package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var levelStrings = [4]string{"DEBUG", "INFO", "WARN", "ERROR"}

// 预分配字符串缓冲池
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 256)
	},
}

// callerInfo 缓存调用信息以减少runtime.Caller调用
type callerInfo struct {
	funcName string
	fileName string
}

var callerCache sync.Map

// Logger 高性能自定义日志器
type Logger struct {
	*log.Logger
	level        LogLevel
	enableCaller bool
}

var defaultLogger *Logger

func init() {
	defaultLogger = New()
}

// New 创建新的日志器
func New() *Logger {
	return &Logger{
		Logger:       log.New(os.Stdout, "", 0),
		level:        DEBUG,
		enableCaller: true,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetCallerEnabled 设置是否启用调用者信息（可以禁用以提高性能）
func (l *Logger) SetCallerEnabled(enabled bool) {
	l.enableCaller = enabled
}

// SetLevel 设置默认日志器级别
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetCallerEnabled 设置默认日志器是否启用调用者信息
func SetCallerEnabled(enabled bool) {
	defaultLogger.SetCallerEnabled(enabled)
}

// getCallerInfoOptimized 优化的调用者信息获取
func getCallerInfoOptimized(skip int) (string, string) {
	if !defaultLogger.enableCaller {
		return "unknown", "unknown"
	}

	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", "unknown"
	}

	// 尝试从缓存获取
	if cached, exists := callerCache.Load(pc); exists {
		info := cached.(callerInfo)
		return info.funcName, info.fileName
	}

	// 缓存未命中，计算并缓存
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	fileName := "unknown"

	if fn != nil {
		fullName := fn.Name()
		if lastDot := strings.LastIndex(fullName, "."); lastDot >= 0 {
			funcName = fullName[lastDot+1:]
		} else {
			funcName = fullName
		}

		file, _ := fn.FileLine(pc)
		fileName = filepath.Base(file)
	}

	// 缓存结果
	callerCache.Store(pc, callerInfo{
		funcName: funcName,
		fileName: fileName,
	})

	return funcName, fileName
}

// func (l *Logger) formatLog(level LogLevel, message string) {
// 	if level < l.level {
// 		return
// 	}

// 	// 从池中获取缓冲区 - 这里有一定优化
// 	buf := bufferPool.Get().([]byte)
// 	defer bufferPool.Put(buf[:0])

// 	// 时间戳 - 实际上手动格式化并没有显著优势
// 	now := time.Now()
// 	timeStr := now.Format("2006-01-02 15:04:05.000")
// 	buf = append(buf, timeStr...)

// 	// 日志级别
// 	buf = append(buf, " |["...)
// 	buf = append(buf, levelStrings[level]...)
// 	buf = append(buf, "]|"...)

// 	// 调用者信息 - runtime.Caller
// 	if l.enableCaller {
// 		funcName, fileName := getCallerInfoOptimized(4)
// 		buf = append(buf, funcName...)
// 		buf = append(buf, '|')
// 		buf = append(buf, fileName...)

// 		// 获取行号
// 		_, _, line, ok := runtime.Caller(3)
// 		if ok {
// 			buf = append(buf, ':')
// 			buf = strconv.AppendInt(buf, int64(line), 10)
// 		}
// 	} else {
// 		buf = append(buf, "unknown|unknown:0"...)
// 	}

// 	buf = append(buf, '|')
// 	buf = append(buf, ' ')
// 	buf = append(buf, message...)

// 	l.Logger.Println(string(buf))
// }

func (l *Logger) formatLog(level LogLevel, message string) {
	if level < l.level {
		return
	}

	// 从池中获取缓冲区
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf[:0])

	// 时间戳
	now := time.Now()
	timeStr := now.Format("2006-01-02 15:04:05.000")
	buf = append(buf, timeStr...)

	// 日志级别
	buf = append(buf, " |["...)
	buf = append(buf, levelStrings[level]...)
	buf = append(buf, "]|"...)

	// 协程ID
	buf = append(buf, []byte("goroutine-")...)
	var stack [32]byte
	runtime.Stack(stack[:], false)
	// 提取 goroutine ID（格式如 "goroutine 123 [running]:")
	parts := strings.SplitN(string(stack[:]), " ", 3)
	if len(parts) > 1 {
		goroutineID := parts[1]
		buf = append(buf, goroutineID...)
	} else {
		buf = append(buf, "unknown"...)
	}
	buf = append(buf, "|"...)

	// 调用者信息
	if l.enableCaller {
		funcName, fileName := getCallerInfoOptimized(4)
		buf = append(buf, funcName...)
		buf = append(buf, '|')
		buf = append(buf, fileName...)

		// 获取行号
		_, _, line, ok := runtime.Caller(3)
		if ok {
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, int64(line), 10)
		}
	} else {
		buf = append(buf, "unknown|unknown:0"...)
	}

	buf = append(buf, '|')
	buf = append(buf, ' ')
	buf = append(buf, message...)

	l.Logger.Println(string(buf))
}

func (l *Logger) Debug(args ...any) {
	if DEBUG < l.level {
		return
	}
	message := fmt.Sprint(args...)
	l.formatLog(DEBUG, message)
}

func (l *Logger) Debugf(format string, args ...any) {
	if DEBUG < l.level {
		return
	}
	message := fmt.Sprintf(format, args...)
	l.formatLog(DEBUG, message)
}

func (l *Logger) Info(args ...any) {
	if INFO < l.level {
		return
	}
	message := fmt.Sprint(args...)
	l.formatLog(INFO, message)
}

func (l *Logger) Infof(format string, args ...any) {
	if INFO < l.level {
		return
	}
	message := fmt.Sprintf(format, args...)
	l.formatLog(INFO, message)
}

func (l *Logger) Warn(args ...any) {
	if WARN < l.level {
		return
	}
	message := fmt.Sprint(args...)
	l.formatLog(WARN, message)
}

func (l *Logger) Warnf(format string, args ...any) {
	if WARN < l.level {
		return
	}
	message := fmt.Sprintf(format, args...)
	l.formatLog(WARN, message)
}

func (l *Logger) Error(args ...any) {
	if ERROR < l.level {
		return
	}
	message := fmt.Sprint(args...)
	l.formatLog(ERROR, message)
}

func (l *Logger) Errorf(format string, args ...any) {
	if ERROR < l.level {
		return
	}
	message := fmt.Sprintf(format, args...)
	l.formatLog(ERROR, message)
}

// 全局函数（使用优化版本）
func Debug(args ...any) {
	defaultLogger.Debug(args...)
}

func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

func Info(args ...any) {
	defaultLogger.Info(args...)
}

func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

func Warn(args ...any) {
	defaultLogger.Warn(args...)
}

func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

func Error(args ...any) {
	defaultLogger.Error(args...)
}

func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}
