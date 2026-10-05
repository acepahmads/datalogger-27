package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}

func (l Level) String() string {
	if int(l) >= 0 && int(l) < len(levelNames) {
		return levelNames[l]
	}
	return "UNKNOWN"
}

type Logger struct {
	mu        sync.Mutex
	level     Level
	file      *os.File
	out       io.Writer
	logChan   chan string
	closeOnce sync.Once
}

var globalLogger *Logger

func Init(logDir string, levelStr string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFile := filepath.Join(logDir, "datalogger.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		file = nil
	}

	lvl := INFO
	switch levelStr {
	case "debug":
		lvl = DEBUG
	case "warn":
		lvl = WARN
	case "error":
		lvl = ERROR
	}

	var writer io.Writer
	if file != nil {
		writer = io.MultiWriter(os.Stdout, file)
	} else {
		writer = os.Stdout
	}

	l := &Logger{
		level:   lvl,
		file:    file,
		out:     writer,
		logChan: make(chan string, 1000), // asynchronous buffer to avoid blocking edge CPU
	}

	// Background worker writes logs asynchronously
	go func() {
		for msg := range l.logChan {
			l.mu.Lock()
			fmt.Fprint(l.out, msg)
			l.mu.Unlock()
		}
	}()

	globalLogger = l
	return l, nil
}

func (l *Logger) log(lvl Level, format string, args ...interface{}) {
	if lvl < l.level {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, lvl.String(), msg)

	select {
	case l.logChan <- entry:
	default:
		// Drop or direct write if buffer full on low-spec device
		l.mu.Lock()
		fmt.Fprint(l.out, entry)
		l.mu.Unlock()
	}
}

func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(DEBUG, format, args...)
	} else {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(INFO, format, args...)
	} else {
		log.Printf("[INFO] "+format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(WARN, format, args...)
	} else {
		log.Printf("[WARN] "+format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(ERROR, format, args...)
	} else {
		log.Printf("[ERROR] "+format, args...)
	}
}

func Close() {
	if globalLogger != nil {
		globalLogger.closeOnce.Do(func() {
			close(globalLogger.logChan)
			if globalLogger.file != nil {
				_ = globalLogger.file.Close()
			}
		})
	}
}
