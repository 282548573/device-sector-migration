/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-11-05 17:40:59
 * @Description:
 * @FilePath: /device-sector-migration/initialize/zap.go
 */
package initialize

import (
	"fmt"
	"os"
	"path"
	"time"

	"device-sector-migration/configs/appyaml"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	zaprotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

var level zapcore.Level

// 初始化zap日志库
func Zap() (logger *zap.Logger) {
	if ok, _ := pathExists(appyaml.YamlCfg.Zap.Director); !ok { // 判断是否有Director文件夹
		fmt.Printf("create %v directory\n", appyaml.YamlCfg.Zap.Director)
		_ = os.Mkdir(appyaml.YamlCfg.Zap.Director, os.ModePerm)
	}
	fmt.Printf("create %v directory\n", appyaml.YamlCfg.Zap.Director)
	switch appyaml.YamlCfg.Zap.Level { // 初始化配置文件的Level
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	if level == zap.ErrorLevel {
		logger = zap.New(getEncoderCore(), zap.AddStacktrace(level))
	} else if level == zap.InfoLevel {
		logger = zap.New(getInfoEncoderCore())
	} else {
		logger = zap.New(getEncoderCore())
	}
	if appyaml.YamlCfg.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	return logger
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  appyaml.YamlCfg.Zap.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	switch {
	case appyaml.YamlCfg.Zap.EncodeLevel == "LowercaseLevelEncoder": // 小写编码器(默认)
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	case appyaml.YamlCfg.Zap.EncodeLevel == "LowercaseColorLevelEncoder": // 小写编码器带颜色
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case appyaml.YamlCfg.Zap.EncodeLevel == "CapitalLevelEncoder": // 大写编码器
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	case appyaml.YamlCfg.Zap.EncodeLevel == "CapitalColorLevelEncoder": // 大写编码器带颜色
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

// getInfoEncoderConfig 获取Info级别的zapcore.EncoderConfig
func getInfoEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey: "message",
		LevelKey:   "level",
		TimeKey:    "time",
		NameKey:    "logger",
		CallerKey:  "caller",
		//StacktraceKey:  appyaml.YamlCfg.Zap.StacktraceKey,
		LineEnding: zapcore.DefaultLineEnding,
		//EncodeLevel:    zapcore.LowercaseLevelEncoder,
		//EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		//EncodeCaller:   zapcore.FullCallerEncoder,
	}
	switch {
	case appyaml.YamlCfg.Zap.EncodeLevel == "LowercaseLevelEncoder": // 小写编码器(默认)
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	case appyaml.YamlCfg.Zap.EncodeLevel == "LowercaseColorLevelEncoder": // 小写编码器带颜色
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case appyaml.YamlCfg.Zap.EncodeLevel == "CapitalLevelEncoder": // 大写编码器
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	case appyaml.YamlCfg.Zap.EncodeLevel == "CapitalColorLevelEncoder": // 大写编码器带颜色
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder() zapcore.Encoder {
	if appyaml.YamlCfg.Zap.Format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig())
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getInfoEncoder 获取Info级别的zapcore.Encoder
func getInfoEncoder() zapcore.Encoder {
	if appyaml.YamlCfg.Zap.Format == "json" {
		return zapcore.NewJSONEncoder(getInfoEncoderConfig())
	}
	return zapcore.NewConsoleEncoder(getInfoEncoderConfig())
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore() (core zapcore.Core) {
	writer, err := getWriteSyncer() // 使用file-rotatelogs进行日志分割
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v", err.Error())
		return
	}
	return zapcore.NewCore(getEncoder(), writer, level)
}

// getInfoEncoderCore 获取Info级别的Encoder的zapcore.Core
func getInfoEncoderCore() (core zapcore.Core) {
	writer, err := getWriteSyncer() // 使用file-rotatelogs进行日志分割
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v", err.Error())
		return
	}
	return zapcore.NewCore(getInfoEncoder(), writer, level)
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(appyaml.YamlCfg.Zap.Prefix + "2006/01/02 - 15:04:05.000"))
}

//文件目录是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//zap logger中加入file-rotatelogs
func getWriteSyncer() (zapcore.WriteSyncer, error) {
	fileWriter, err := zaprotatelogs.New(
		path.Join(appyaml.YamlCfg.Zap.Director, "%Y-%m-%d.log"),
		zaprotatelogs.WithLinkName(appyaml.YamlCfg.Zap.LinkName),
		zaprotatelogs.WithMaxAge(7*24*time.Hour),
		zaprotatelogs.WithRotationTime(24*time.Hour),
	)
	if appyaml.YamlCfg.Zap.LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}
