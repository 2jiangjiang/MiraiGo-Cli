package MiraiGo

import (
	"MiraiGo-Cli/global"
	"MiraiGo-Cli/internal/base"
	"crypto/md5"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
	"time"
)

func Version() string {
	return "1.0.0"
}
func Login(account string, password string) bool {
	base.Parse()
	base.Init()
	base.Account.Uin, _ = strconv.ParseInt(account, 0, 64)
	base.Account.Password = password

	rotateOptions := []rotatelogs.Option{
		rotatelogs.WithRotationTime(time.Hour * 24),
	}
	rotateOptions = append(rotateOptions, rotatelogs.WithMaxAge(base.LogAging))
	if base.LogForceNew {
		rotateOptions = append(rotateOptions, rotatelogs.ForceNewFile())
	}
	w, err := rotatelogs.New(path.Join("logs", "%Y-%m-%d.log"), rotateOptions...)
	if err != nil {
		log.Errorf("rotatelogs init err: %v", err)
		panic(err)
	}

	consoleFormatter := global.LogFormat{EnableColor: base.LogColorful}
	fileFormatter := global.LogFormat{EnableColor: false}
	log.AddHook(global.NewLocalHook(w, consoleFormatter, fileFormatter, global.GetLogLevel(base.LogLevel)...))

	if (base.Account.Uin == 0 || (base.Account.Password == "" && !base.Account.Encrypt)) && !global.PathExists("session.token") {
		log.Warn("账号密码未配置, 将使用二维码登录.")
		if !base.FastStart {
			log.Warn("将在 5秒 后继续.")
			time.Sleep(time.Second * 5)
		}
	}

	log.Info("当前版本:", "debug")
	if base.Debug {
		log.SetLevel(log.DebugLevel)
		log.Warnf("已开启Debug模式.")
		// log.Debugf("开发交流群: 192548878")
	}
	if !global.PathExists("device.json") {
		log.Warn("虚拟设备信息不存在, 将自动生成随机设备.")
		client.GenRandomDevice()
		_ = os.WriteFile("device.json", client.SystemDeviceInfo.ToJson(), 0o644)
		log.Info("已生成设备信息并保存到 device.json 文件.")
	} else {
		log.Info("将使用 device.json 内的设备信息运行Bot.")
		if err := client.SystemDeviceInfo.ReadJson([]byte(global.ReadAllText("device.json"))); err != nil {
			log.Fatalf("加载设备信息失败: %v", err)
		}
	}

	base.PasswordHash = md5.Sum([]byte(base.Account.Password))

	Cli = newClient()
	if base.Account.Uin != 0 && base.PasswordHash != [16]byte{} {
		Cli.Uin = base.Account.Uin
		Cli.PasswordMd5 = base.PasswordHash
	}
	log.Info("开始尝试登录并同步消息...")
	log.Infof("使用协议: %s", client.SystemDeviceInfo.Protocol)
	log.Infof("正在尝试登录:%d", Cli.Uin)
	if err := commonLogin(); err != nil {
		log.Fatalf("登录时发生致命错误: %v", err)
		return false
	}
	return true
}
func newClient() *client.QQClient {
	c := client.NewClientEmpty()
	c.UseFragmentMessage = base.ForceFragmented
	c.OnServerUpdated(func(bot *client.QQClient, e *client.ServerUpdatedEvent) bool {
		if !base.UseSSOAddress {
			log.Infof("收到服务器地址更新通知, 根据配置文件已忽略.")
			return false
		}
		log.Infof("收到服务器地址更新通知, 将在下一次重连时应用. ")
		return true
	})
	if global.PathExists("address.txt") {
		log.Infof("检测到 address.txt 文件. 将覆盖目标IP.")
		addr := global.ReadAddrFile("address.txt")
		if len(addr) > 0 {
			c.SetCustomServer(addr)
		}
		log.Infof("读取到 %v 个自定义地址.", len(addr))
	}
	c.SetLogger(protocolLogger{})
	return c
}

type protocolLogger struct{}

const fromProtocol = "Protocol -> "

func (p protocolLogger) Info(format string, arg ...any) {
	log.Infof(fromProtocol+format, arg...)
}

func (p protocolLogger) Warning(format string, arg ...any) {
	log.Warnf(fromProtocol+format, arg...)
}

func (p protocolLogger) Debug(format string, arg ...any) {
	log.Debugf(fromProtocol+format, arg...)
}

func (p protocolLogger) Error(format string, arg ...any) {
	log.Errorf(fromProtocol+format, arg...)
}

func (p protocolLogger) Dump(data []byte, format string, arg ...any) {
	if !global.PathExists(global.DumpsPath) {
		_ = os.MkdirAll(global.DumpsPath, 0o755)
	}
	dumpFile := path.Join(global.DumpsPath, fmt.Sprintf("%v.dump", time.Now().Unix()))
	message := fmt.Sprintf(format, arg...)
	log.Errorf("出现错误 %v. 详细信息已转储至文件 %v 请连同日志提交给开发者处理", message, dumpFile)
	_ = os.WriteFile(dumpFile, data, 0o644)
}
