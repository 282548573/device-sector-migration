/*
 * @Author: bill
 * @Date: 2021-11-04 15:59:27
 * @LastEditors: bill
 * @LastEditTime: 2021-11-05 17:43:09
 * @Description:
 * @FilePath: /device-sector-migration/main.go
 */
package main

import (
	"os"

	"device-sector-migration/configs/appyaml"
	"device-sector-migration/global"
	"device-sector-migration/initialize"
	"device-sector-migration/sector"

	cli "github.com/urfave/cli/v2"
)

func main() {
	appyaml.InitYamlConfig()
	global.Logger = initialize.Zap() // 初始化zap日志库
	app := &cli.App{
		Name: "storage-sector",
		Commands: []*cli.Command{
			{
				Name:  "sector-migration",
				Usage: "sector-扇区转移",
				Action: func(c *cli.Context) error {
					sector.SectorMigration()
					return nil
				},
			},
			{
				Name:  "sector-check",
				Usage: "sector-校验结果",
				Action: func(c *cli.Context) error {
					sector.SectorCheck()
					return nil
				},
			},
		},
	}
	_ = app.Run(os.Args)
}
