package shared

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func buildLumberjackSyncer(maxSize int, maxAge int, filePath string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   fmt.Sprintf("/var/log/%s", filePath),
		MaxSize:    maxSize, // maximum size in megabytes before log file gets rotated
		MaxBackups: 7,       // maximum number of old log files to retain
		MaxAge:     maxAge,  // maximum number of days to retain old log files
		Compress:   false,   // whether to compress the rotated log files using gzip
	}
}

type Logger struct {
	FilePath string
	MaxAge   int
	MaxSize  int
	Level    string
	AppName  string
	logger   *zap.Logger
}

func NewLogger(filePath string, maxAge int, maxSize int, level string, appName string) *Logger {
	return &Logger{
		FilePath: filePath,
		MaxAge:   maxAge,
		MaxSize:  maxSize,
		Level:    level,
		AppName:  appName,
	}
}

// Add new field to the logger fields (to head)
func unshift(fields []zap.Field, field zap.Field) []zap.Field {
	return append([]zap.Field{field}, fields...)
}

func (l *Logger) Init() {

	// file syncer
	lumberJackSyncer := buildLumberjackSyncer(l.MaxSize, l.MaxAge, l.FilePath)
	syncer := zapcore.AddSync(lumberJackSyncer)

	// stdout syncer
	stdoutSyncer := zapcore.AddSync(os.Stdout)
	combine := zapcore.NewMultiWriteSyncer(syncer, stdoutSyncer)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), combine, zap.InfoLevel)
	l.logger = zap.New(core)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, unshift(fields, zap.String("service", l.AppName))...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, unshift(fields, zap.String("service", l.AppName))...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, unshift(fields, zap.String("service", l.AppName))...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, unshift(fields, zap.String("service", l.AppName))...)
}
