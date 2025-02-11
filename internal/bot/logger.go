package bot

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

// Logger 日志记录器
type Logger struct {
	logFile *os.File
}

// NewLogger 创建新的日志记录器
func NewLogger() (*Logger, error) {
	// 创建logs目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// 创建日志文件
	logFileName := fmt.Sprintf("logs/openledger_%s.log", time.Now().Format("2006-01-02"))
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &Logger{
		logFile: logFile,
	}, nil
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Log 记录日志
func (l *Logger) Log(message string) {
	// 获取当前时间
	now := time.Now().Format("2006-01-02 15:04:05 MST")

	// 写入日志文件
	if l.logFile != nil {
		fmt.Fprintf(l.logFile, "[%s] %s\n", now, message)
	}

	// 打印到控制台
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite, color.Bold)
	
	cyan.Printf("[ %s ]", now)
	white.Print(" | ")
	fmt.Println(message)
}

// CleanOldLogs 清理旧日志文件
func (l *Logger) CleanOldLogs(daysToKeep int) error {
	files, err := filepath.Glob("logs/openledger_*.log")
	if err != nil {
		return fmt.Errorf("failed to list log files: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -daysToKeep)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(file)
		}
	}

	return nil
} 