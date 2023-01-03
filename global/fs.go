package global

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net/netip"
	"os"
	"strings"
)

const (
	// ImagePath go-cqhttp使用的图片缓存目录
	ImagePath = "data/images"
	// VoicePath go-cqhttp使用的语音缓存目录
	VoicePath = "data/voices"
	// VideoPath go-cqhttp使用的视频缓存目录
	VideoPath = "data/videos"
	// CachePath go-cqhttp使用的缓存目录
	CachePath = "data/cache"
	// DumpsPath go-cqhttp使用错误转储目录
	DumpsPath = "dumps"
	// HeaderAmr AMR文件头
	HeaderAmr = "#!AMR"
	// HeaderSilk Silkv3文件头
	HeaderSilk = "\x02#!SILK_V3"
)

func SetRoot(path string) string {
	log.Info("切换工作目录到" + path)
	if err := os.Chdir(path); err != nil {
		log.Fatal(err)
	}
	path, _ = os.Getwd()
	return path
}

// PathExists 判断给定path是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || errors.Is(err, os.ErrExist)
}
func ReadAllText(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Error(err)
		return ""
	}
	return string(b)
}
func DelFile(path string) bool {
	err := os.Remove(path)
	if err != nil {
		// 删除失败
		log.Error(err)
		return false
	}
	// 删除成功
	log.Info(path + "删除成功")
	return true
}
func ReadAddrFile(path string) []netip.AddrPort {
	d, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	str := string(d)
	lines := strings.Split(str, "\n")
	var ret []netip.AddrPort
	for _, l := range lines {
		addr, err := netip.ParseAddrPort(l)
		if err == nil {
			ret = append(ret, addr)
		}
	}
	return ret
}
