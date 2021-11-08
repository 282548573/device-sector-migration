/*
 * @Author: bill
 * @Date: 2021-11-04 16:07:02
 * @LastEditors: bill
 * @LastEditTime: 2021-11-05 18:31:29
 * @Description:
 * @FilePath: /device-sector-migration/sector/sector_migration.go
 */
package sector

import (
	"device-sector-migration/configs"
	"device-sector-migration/global"
	"device-sector-migration/utils"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/daemon/graphdriver/copy"
)

type SectorDisk struct {
	Mountpoint    string `json:"mountpoint"`
	DiskType      string `json:"disk_type"`
	Total         uint64 `json:"total"`
	Free          uint64 `json:"free"`
	FreeSectorNum uint64 `json:"free_sector_num"`
}

func SectorMigration() {
	// 1. 读取扇区文件
	var sectorList configs.SectorMaps
	if err := sectorList.ReadConfig(); err != nil {
		global.Logger.Error(err.Error())
		return
	}

	// 2. 获取非hdd的所有磁盘和hdd所有磁盘
	notHddDisks, hddDisks, err := getNotHddDiskAndHddDisk()
	if err != nil {
		global.Logger.Error(err.Error())
		return
	}

	if len(notHddDisks) == 0 {
		global.Logger.Info("没有扇区需要迁移, 程序终止!")
		return
	}

	if len(hddDisks) == 0 {
		global.Logger.Info("没有hdd磁盘, 程序终止!")
		return
	}

	// 3. 更新map表中目标地址
	updateSectorMaps(notHddDisks, sectorList.Maps)

	// 4. 获取没有完成迁移的扇区个数以及扇区大小
	noHddSectorNum, sectorSize, err := getNoHddSectorNumAndSectorSize(sectorList.Maps)
	if err != nil {
		global.Logger.Error(err.Error())
		return
	}

	if noHddSectorNum == 0 {
		global.Logger.Info("扇区已迁移完成, 程序终止!")
		return
	}

	// 5. 获取所有磁盘可存扇区个数以及总个数
	var hddSectorNum uint64
	for i := 0; i < len(hddDisks); i++ {
		freeSectorNum := hddDisks[i].Free / sectorSize
		hddDisks[i].FreeSectorNum = freeSectorNum
		hddSectorNum = hddSectorNum + freeSectorNum
	}

	msg := fmt.Sprintf("需要迁移%d个扇区,当前所有hdd磁盘可存%d个扇区", noHddSectorNum, hddSectorNum)
	global.Logger.Info(msg)
	if noHddSectorNum > hddSectorNum {
		global.Logger.Info("hdd磁盘剩余空间不够迁移, 程序终止!")
		return
	}

	// 6. 拷贝扇区
	curCount := 0
	lenMaps := len(sectorList.Maps)
	for number, info := range sectorList.Maps {
		// 选择空间最多的那个磁盘拷贝扇区
		if info.TargetPath == "" {
			maxIndex := 0
			var maxCount uint64
			for i, v := range hddDisks {
				if v.FreeSectorNum > maxCount {
					maxIndex = i
					maxCount = v.FreeSectorNum
				}
				msg := fmt.Sprintf("磁盘%s剩余可存%d个", v.Mountpoint, v.FreeSectorNum)
				global.Logger.Info(msg)
			}
			info.TargetPath = hddDisks[maxIndex].Mountpoint
			hddDisks[maxIndex].FreeSectorNum -= 1
			global.Logger.Info("选择磁盘：" + info.TargetPath)
		}

		curCount++
		if !info.IsFinish {
			sectorNumber := fmt.Sprintf("s-%s-%d", info.MinerId, number)
			isFinish := copySector(info.SourcePath, info.TargetPath, sectorNumber)
			if isFinish {
				msg := fmt.Sprintf("总共需要拷贝%d个扇区, 当前第%d个扇区(%s), 拷贝成功", lenMaps, curCount, sectorNumber)
				global.Logger.Info(msg)
				global.Logger.Info("===================================================================\n")
			} else {
				msg := fmt.Sprintf("总共需要拷贝%d个扇区, 当前第%d个扇区(%s), 拷贝失败", lenMaps, curCount, sectorNumber)
				global.Logger.Info(msg)
				global.Logger.Info("===================================================================\n")
			}
			info.IsFinish = isFinish
		}
	}

	global.Logger.Info("===========拷贝结果================")
	for number, info := range sectorList.Maps {
		msg := fmt.Sprintf("扇区:%d 源磁盘:%s ====> 目标磁盘:%s 是否完成:%v", number, info.SourcePath, info.TargetPath, info.IsFinish)
		global.Logger.Info(msg)
		global.Logger.Info("===========================\n")
	}

	if err := sectorList.WriteConfig(); err != nil {
		global.Logger.Error(err.Error())
	}
}

func getNotHddDiskAndHddDisk() ([]SectorDisk, []SectorDisk, error) {
	notHddDisks := make([]SectorDisk, 0)
	hddDisks := make([]SectorDisk, 0)
	disks, err := utils.GetAllDiskInfo()
	if err != nil {
		global.Logger.Error(err.Error())
		return notHddDisks, hddDisks, err
	}

	for _, disk := range disks {
		if disk.Mountpoint == "/" {
			if utils.FileOrDirIsExists("/.lotusworker") {
				notHddDisks = append(notHddDisks, SectorDisk{
					Mountpoint: disk.Mountpoint,
					DiskType:   disk.DiskType,
					Total:      disk.Total,
					Free:       disk.Free,
				})
			}
		} else {
			if utils.FileOrDirIsExists(disk.Mountpoint + "/.lotusworker") {
				if strings.Contains(disk.Mountpoint, "/hdd_") {
					hddDisks = append(hddDisks, SectorDisk{
						Mountpoint: disk.Mountpoint,
						DiskType:   disk.DiskType,
						Total:      disk.Total,
						Free:       disk.Free,
					})
				} else {
					notHddDisks = append(notHddDisks, SectorDisk{
						Mountpoint: disk.Mountpoint,
						DiskType:   disk.DiskType,
						Total:      disk.Total,
						Free:       disk.Free,
					})
				}
			}
		}
	}

	global.Logger.Info("非hdd磁盘下存在.lotusworker文件的磁盘列表：")
	for _, disk := range notHddDisks {
		global.Logger.Sugar().Info(disk.Mountpoint)
	}

	global.Logger.Info("-------------------------------\n")

	global.Logger.Info("hdd磁盘下存在.lotusworker文件的磁盘列表：")
	for _, disk := range hddDisks {
		global.Logger.Sugar().Info(disk.Mountpoint)
	}

	global.Logger.Info("-------------------------------\n")
	return notHddDisks, hddDisks, nil
}

func updateSectorMaps(notHddDisks []SectorDisk, sectorMaps map[int64]*configs.SectorInfo) {
	for _, disk := range notHddDisks {
		path := fmt.Sprintf("%s/.lotusworker/sealed", disk.Mountpoint)
		if disk.Mountpoint == "/" {
			path = "/.lotusworker/sealed"
		}

		files, err := ioutil.ReadDir(path)
		if err != nil {
			msg := fmt.Sprintf("获取%s下的文件失败,err:%s", disk.Mountpoint, err.Error())
			global.Logger.Error(msg)
			log.Println(msg)
			return
		}
		for _, file := range files {
			res := strings.Split(file.Name(), "-")
			if len(res) > 2 {
				num, err := strconv.ParseInt(res[2], 10, 64)
				if err != nil {
					global.Logger.Error(fmt.Sprintf("解析%s下%s的扇区文件失败,err:%s", disk.Mountpoint, file.Name(), err.Error()))
					continue
				}
				if v, ok := sectorMaps[num]; ok {
					v.SourcePath = disk.Mountpoint
					v.MinerId = res[1]
				} else {
					sectorMaps[num] = &configs.SectorInfo{
						SourcePath: disk.Mountpoint,
						MinerId:    res[1],
					}
				}
			}
		}
	}
}

func getNoHddSectorNumAndSectorSize(sectorMaps map[int64]*configs.SectorInfo) (uint64, uint64, error) {
	var noHddSectorNum uint64
	var sectorSize uint64
	for k, v := range sectorMaps {
		if !v.IsFinish {
			noHddSectorNum++
		}
		if sectorSize == 0 {
			sealPath := fmt.Sprintf("%s/.lotusworker/sealed/s-%s-%d", v.SourcePath, v.MinerId, k)
			if v.SourcePath == "/" {
				sealPath = fmt.Sprintf("/.lotusworker/sealed/s-%s-%d", v.MinerId, k)
			}
			size, err := utils.GetFileSize(sealPath)
			if err != nil {
				msg := fmt.Sprintf("获取%s扇区文件大小失败", sealPath)
				global.Logger.Error(msg)
				return noHddSectorNum, sectorSize, err
			}
			sectorSize += uint64(size)

			cachePath := fmt.Sprintf("%s/.lotusworker/cache/s-%s-%d", v.SourcePath, v.MinerId, k)
			if v.SourcePath == "/" {
				cachePath = fmt.Sprintf("/.lotusworker/cache/s-%s-%d", v.MinerId, k)
			}
			files, err := ioutil.ReadDir(cachePath)
			if err != nil {
				message := fmt.Sprintf("获取%s下的文件失败,err:%s", cachePath, err.Error())
				global.Logger.Error(message)
				return noHddSectorNum, sectorSize, err
			}
			for _, file := range files {
				file := fmt.Sprintf("%s/%s", cachePath, file.Name())
				size, err := utils.GetFileSize(file)
				if err != nil {
					msg := fmt.Sprintf("获取%s扇区文件大小失败", file)
					global.Logger.Error(msg)
					return noHddSectorNum, sectorSize, err
				}
				sectorSize += uint64(size)
			}
		}
	}
	msg := fmt.Sprintf("扇区大小:%d", sectorSize)
	global.Logger.Info(msg)

	return noHddSectorNum, sectorSize, nil
}

func copySector(sourcePath string, targetPath string, sector string) (isFinish bool) {
	// 1. 拷贝扇区sealed文件
	sealSPath := fmt.Sprintf("%s/.lotusworker/sealed/%s", sourcePath, sector)
	if sourcePath == "/" {
		sealSPath = fmt.Sprintf("/.lotusworker/sealed/%s", sector)
	}
	sealTPath := fmt.Sprintf("%s/.lotusworker/sealed/%s", targetPath, sector)
	msg := fmt.Sprintf("拷贝扇区:%s 源位置:%s ====> 目标位置:%s 拷贝中.....", sector, sealSPath, sealTPath)
	global.Logger.Info(msg)

	if err := os.RemoveAll(sealTPath); err != nil {
		global.Logger.Error(err.Error())
	}

	if err := copy.DirCopy(sealSPath, sealTPath, copy.Content, false); err != nil {
		msg = fmt.Sprintf("拷贝扇区:%s 源位置:%s 目标位置:%s 失败, err= %v\n", sector, sealSPath, sealTPath, err.Error())
		global.Logger.Error(msg)
		return false
	}

	// 2. 拷贝扇区cache文件夹
	cacheSPath := fmt.Sprintf("%s/.lotusworker/cache/%s/", sourcePath, sector)
	if sourcePath == "/" {
		cacheSPath = fmt.Sprintf("/.lotusworker/cache/%s/", sector)
	}
	cacheTPath := fmt.Sprintf("%s/.lotusworker/cache/%s/", targetPath, sector)
	msg = fmt.Sprintf("拷贝扇区:%s 源位置:%s ====> 目标位置:%s 拷贝中.....", sector, cacheSPath, cacheTPath)
	global.Logger.Info(msg)

	if err := os.RemoveAll(cacheTPath); err != nil {
		global.Logger.Error(err.Error())
	}

	if err := copy.DirCopy(cacheSPath, cacheTPath, copy.Content, false); err != nil {
		msg = fmt.Sprintf("拷贝扇区:%s 源位置:%s 目标位置:%s 失败, err= %v\n", sector, cacheSPath, cacheTPath, err.Error())
		global.Logger.Error(msg)
		return false
	}
	return true
}
