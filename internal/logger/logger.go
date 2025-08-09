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
	now := time.Now()

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

	cf := zap.NewProductionConfig()

	consoleECf := zap.NewDevelopmentEncoderConfig()
	consoleECf.EncodeTime = zapcore.TimeEncoderOfLayout(time.TimeOnly)
	consoleECf.EncodeCaller = zapcore.ShortCallerEncoder
	consoleECf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleE := zapcore.NewConsoleEncoder(consoleECf)

	jsonECf := cf.EncoderConfig
	jsonECf.EncodeTime = zapcore.TimeEncoderOfLayout(time.UnixDate)
	jsonECf.EncodeCaller = zapcore.ShortCallerEncoder
	jsonECf.EncodeLevel = zapcore.CapitalLevelEncoder
	jsonE := zapcore.NewJSONEncoder(jsonECf)

	consoleC := zapcore.NewCore(consoleE, zapcore.AddSync(zapcore.Lock(os.Stdout)), zapcore.InfoLevel)
	fileC := zapcore.NewCore(jsonE, zapcore.AddSync(zapcore.Lock(file)), zapcore.InfoLevel)

	c := zapcore.NewTee(consoleC, fileC)

	lg := zap.New(c, zap.AddCaller(), zap.AddCallerSkip(1))
	s := lg.Sugar()

	l := &Logger{
		s: s,
		l: lg,
	}

	l.Info("initialized logger", "duration", time.Since(now))
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
	now := time.Now()

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

		for i := range len(fileInfos) - maxLogFiles {
			fileDirectory := filepath.Join(logDir, fileInfos[i].Name())
			err := os.Remove(fileDirectory)
			if err != nil {
				l.Error("logger remove file error", "fileDirectory", fileDirectory, "error", err)
				return err
			}
		}
	}

	l.Info("logger files updated", "duration", time.Since(now))
	return nil
}
