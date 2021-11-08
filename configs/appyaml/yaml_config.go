/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-11-08 10:49:48
 * @Description: 服务config.yaml中的配置
 * @FilePath: /device-sector-migration/configs/appyaml/yaml_config.go
 */

package appyaml

import (
	"device-sector-migration/global"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var YamlCfg YamlConfig

type YamlConfig struct {
	Zap    Zap    `mapstructure:"zap" json:"zap" yaml:"zap"`          // zap日志配置
	System System `mapstructure:"system" json:"system" yaml:"system"` // 服务基础配置
}

func InitYamlConfig() {
	config := "./configs/config.yaml"

	v := viper.New()
	v.SetConfigFile(config)
	err := v.ReadInConfig()
	if err != nil {
		global.Logger.Error(err.Error())
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.Unmarshal(&YamlCfg); err != nil {
			global.Logger.Error(err.Error())
		}
	})

	if err := v.Unmarshal(&YamlCfg); err != nil {
		global.Logger.Error(err.Error())
	}
}
