package zaplog

import (
	"github.com/chenxijunlove/go-log/conf"
	"github.com/chenxijunlove/go-log/fileout"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

type Log struct {
	logger *zap.Logger
}

//var Log *zap.Logger //全局日志

func parseLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "panic", "dpanic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.DebugLevel
	}
}

//创建日志
func New(opts ...conf.Option) *Log {
	o := &conf.Options{
		LogPath:     conf.LogPath,
		LogName:     conf.LogName,
		LogLevel:    conf.LogLevel,
		MaxSize:     conf.MaxSize,
		MaxAge:      conf.MaxAge,
		Stacktrace:  conf.Stacktrace,
		IsStdOut:    conf.IsStdOut,
		ProjectName: conf.ProjectName,
		FileAsync:   conf.FileAsync,
	}
	for _, opt := range opts {
		opt(o)
	}
	writers := make([]zapcore.WriteSyncer, 0)
	if o.FileAsync == "yes" {
		writers = append(writers, fileout.NewRollingFile(o.LogPath, o.LogName, o.MaxSize, o.MaxAge))
	} else {
		writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
			Filename:  filepath.Join(o.LogPath, o.LogName+".log"),
			MaxSize:   o.MaxSize, //megabytes
			MaxAge:    o.MaxAge,  //days
			LocalTime: true,
			Compress:  false,
		}))
	}
	if o.IsStdOut == "yes" {
		writers = append(writers, os.Stdout)
	}
	logger := newZapLogger(parseLevel(o.LogLevel), parseLevel(o.Stacktrace), zapcore.NewMultiWriteSyncer(writers...))
	zap.RedirectStdLog(logger)
	logger = logger.With(zap.String("project", o.ProjectName)) //加上项目名称
	return &Log{logger: logger}
}

func newZapLogger(level, stacktrace zapcore.Level, output zapcore.WriteSyncer) *zap.Logger {
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "app",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		//EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		//	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		//},
		EncodeTime: zapcore.ISO8601TimeEncoder,
	}

	var encoder zapcore.Encoder
	dyn := zap.NewAtomicLevel()
	//encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	//encoder = zapcore.NewJSONEncoder(encCfg) // zapcore.NewConsoleEncoder(encCfg)
	dyn.SetLevel(level)
	encCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoder = zapcore.NewJSONEncoder(encCfg)

	return zap.New(zapcore.NewCore(encoder, output, dyn), zap.AddCaller(), zap.AddStacktrace(stacktrace), zap.AddCallerSkip(2))
}
