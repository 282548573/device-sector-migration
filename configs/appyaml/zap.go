/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-10-29 14:22:29
 * @Description: 服务config.yaml中的zap日志打印配置
 * @FilePath: /ars-device-server/configs/appyaml/zap.go
 */
package appyaml

type Zap struct {
	// 打印日志级别debug、info、warn、error、dpanic、panic、fatal
	Level string `mapstructure:"level" json:"level" yaml:"level"`

	// 日志输出的格式：默认json
	Format string `mapstructure:"format" json:"format" yaml:"format"`

	// 日志输出的格式: 添加前缀
	Prefix string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`

	// 存放日志的目录
	Director string `mapstructure:"director" json:"director"  yaml:"director"`

	LinkName string `mapstructure:"link-name" json:"linkName" yaml:"link-name"`

	// 使用zap时显示文件名、行号和函数名注释等消息
	ShowLine bool `mapstructure:"show-line" json:"showLine" yaml:"showLine"`

	// 编码器级别："LowercaseLevelEncoder": // 小写编码器(默认)、"LowercaseColorLevelEncoder": // 小写编码器带颜色、
	// "CapitalLevelEncoder": // 大写编码器、"CapitalColorLevelEncoder": // 大写编码器带颜色
	EncodeLevel string `mapstructure:"encode-level" json:"encodeLevel" yaml:"encode-level"`

	// zap EncoderConfig配置的StacktraceKey属性值
	StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktraceKey" yaml:"stacktrace-key"`

	// 同时记录在控制台
	LogInConsole bool `mapstructure:"log-in-console" json:"logInConsole" yaml:"log-in-console"`
}
