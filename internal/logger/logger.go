package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logDir = "./logs"
const maxLogFiles = 20

type Logger struct {
	s *zap.SugaredLogger
	l *zap.Logger
}

func Init() (*Logger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return &Logger{}, err
	}

	logFileName := fmt.Sprintf("proxy_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFilePath := filepath.Join(logDir, logFileName)
	file, err := os.Create(logFilePath)
	if err != nil {
		log.Printf("Failed to create log file: %v", err)
		return &Logger{}, err
	}

	config := zap.NewProductionConfig()

	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.TimeOnly)
	consoleEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	jsonEncoderConfig := config.EncoderConfig
	jsonEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.UnixDate)
	jsonEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	jsonEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	jsonEncoder := zapcore.NewJSONEncoder(jsonEncoderConfig)

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(zapcore.Lock(os.Stdout)), zapcore.InfoLevel)
	fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(zapcore.Lock(file)), zapcore.InfoLevel)

	core := zapcore.NewTee(consoleCore, fileCore)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := logger.Sugar()

	l := &Logger{
		s: sugar,
		l: logger,
	}

	l.Info("loaded logger")
	err = l.manageLogFiles()

	return l, err
}

func (l *Logger) GetSugaredLogger() *zap.SugaredLogger {
	return l.s
}

func (l *Logger) GetLogger() *zap.Logger {
	return l.l
}

func (l *Logger) Close() {
	l.l.Sync()
	l.s.Sync()
}

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.s.Infow(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.s.Errorw(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.s.Warnw(msg, keysAndValues...)
}

// Checks folder, counts files and checks if it reaches file limit. Removes oldest files until under limit.
func (l *Logger) manageLogFiles() error {
	files, err := os.ReadDir(logDir)
	if err != nil {
		l.Error("logger read directory error", "files", files, "error", err)
		return err
	}

	if len(files) > maxLogFiles {
		fileInfos := make([]os.FileInfo, len(files))
		for i, file := range files {
			info, err := file.Info()
			if err != nil {
				l.Error("logger get file error", "fileInfo", info, "error", err)
				return err
			}
			fileInfos[i] = info
		}

		sort.Slice(fileInfos, func(i, j int) bool {
			return fileInfos[i].ModTime().Before(fileInfos[j].ModTime())
		})

		for i := 0; i < len(fileInfos)-maxLogFiles; i++ {
			fileDirectory := filepath.Join(logDir, fileInfos[i].Name())
			err := os.Remove(fileDirectory)
			if err != nil {
				l.Error("logger remove file error", "fileDirectory", fileDirectory, "error", err)
				return err
			}
		}
	}

	l.Info("logger files updated")
	return nil
}
