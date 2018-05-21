/*
	package logger 日志模块(非通用)
	配置文件 business/common/config/log.json
	{
	"common":{
		"logpath":"."         // 日志文件路径(相对路径相对可执行文件目录，
			或绝对路径，注意：window系统绝对路径以[盘符]://开头)
		"logfile":"main.log", // 日志文件名,为空则输出到标准输出
		"level": "debug"},       // 日志级别 panic<fatal<error<warning<info<debug 参考logrus.Level
		"formatter": "json/text/pretty", // 输出格式json/text/pretty
	"模块1":{
		...
		"ignoreinputs":["handler1"], // 模块配置中忽略特定的handler输入 仅适用于gin
		"ignoreoutputs":["handler2"]  // 模块配置中忽略特定的handler输出 仅适用于gin
		...
	},
	"模块2":{...}...
	}
	使用：
		import 会默认按照配置文件初始化
		或手动初始化：LoadConfig初始化主配置，NewLogger初始化模块日志
		GetLogger获得一个模块日志
		全部logger均实现标准库Logger的全部函数，可以进行直接切换
	其他：
		标准库log的输出在初始化后会被重定向至主日志对象，标准库输出将以info级别进行输出。
		日志切分/压缩等以后将由writer实现，但不建议使用，建议使用运维管理工具进行维护。
	example：
		import (
			"saas/common/core/db/logger"
			"github.com/sirupsen/logrus"
			)
		...

		var log *logrus.Entry

		func init(){
			...
			log = logger.GetLogger("im")
			...
			log.Println(...)
			log.Errorln(...)
			log.Info(...)
			log.Debug(...)
			log.Panic(...)
			log.WithFields(logrus.Fields{"f1":"v1", "f2":"v2"...}).Debug(...)
			...
		}

*/
package logger

import (
	"io/ioutil"
	stdlog "log"
	"os"
	. "saas/common/core/logger/config"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Entry
}

var mainlog *logrus.Entry
var backends map[string]*worker

var config *Config

// debug
func (self Logger) Debug(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Debug(args...)
}

func (self Logger) Debugf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Debugf(format, args...)
}

func (self Logger) Debugln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Debugln(args...)
}

func (self Logger) Error(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Error(args...)
}

// error
func (self Logger) Errorf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Errorf(format, args...)
}

func (self Logger) Errorln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Errorln(args...)
}

// fatal
func (self Logger) Fatal(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Fatal(args...)
}

func (self Logger) Fatalf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Fatalf(format, args...)
}

func (self Logger) Fatalln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Fatalln(args...)
}

// panic
func (self Logger) Panic(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Panic(args...)
}

func (self Logger) Panicf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Panicf(format, args...)
}

func (self Logger) Panicln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Panicln(args...)
}

// warning&warn
func (self Logger) Warn(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Warn(args...)
}

func (self Logger) Warnf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Warnf(format, args...)
}

func (self Logger) Warnln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Warnln(args...)
}

func (self Logger) Warning(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Warning(args...)
}

func (self Logger) Warningf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Warningf(format, args...)
}

func (self Logger) Warningln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Warningln(args...)
}

// print
func (self Logger) Print(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Print(args...)
}

func (self Logger) Printf(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Printf(format, args...)
}

func (self Logger) Println(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Println(args...)
}

// info
func (self Logger) Info(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Info(args...)
}

func (self Logger) Infof(format string, args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Infof(format, args...)
}

func (self Logger) Infoln(args ...interface{}) {
	self.Entry.WithField("Goid", getGoid()).Infoln(args...)
}

// 初始化 path = "common/config/log.json"
func Init(path string) {
	backends = map[string]*worker{}
	configpath, err := getAbs(path)
	if err != nil {
		stdlog.Panicln(err.Error())
	}
	f, err := os.OpenFile(configpath, os.O_RDONLY, os.ModeType)
	if err != nil {
		stdlog.Panicln(err.Error())
	}
	defer f.Close()
	buff, err := ioutil.ReadAll(f)
	config := LoadJson(buff)
	// 初始化common日志
	LoadConfig(config)
	// 循环初始化模块日志
	for _, k := range config.Keys() {
		if k != "common" {
			NewLogger(k, config.GetConfig(k), nil)
		}
	}

}

// 主（common）日志加载配置
func LoadConfig(c *Config) {
	config = c
	if mainlog == nil {
		// 初始化主日志
		log := logrus.New()
		mainlog = log.WithFields(logrus.Fields{ROUTE_KEY: "common"})
	}

	// 级别
	level := config.GetValue("common", "level")
	mainlog.Logger.SetLevel(getLevel(level))
	// 格式
	format := config.GetValue("common", "formatter")
	switch format {
	case "json":
		mainlog.Logger.Formatter = &logrus.JSONFormatter{
			TimestampFormat: LOG_TIME_FORMAT,
		}
	case "text":
		mainlog.Logger.Formatter = &logrus.TextFormatter{
			TimestampFormat: LOG_TIME_FORMAT,
		}
	case "pretty":
		mainlog.Logger.Formatter = &PrettyFormatter{}
	default:
		// 默认json
		mainlog.Logger.Formatter = &logrus.JSONFormatter{
			TimestampFormat: LOG_TIME_FORMAT,
		}
	}

	// 输出
	logpath := config.GetValue("common", "logpath")
	if logpath == "" {
		logpath = "."
	}
	logfile := config.GetValue("common", "logfile")
	if logfile != "" {
		logfile, _ := getAbs(logpath + "/" + logfile)
		mainlog.Logger.Out = FileWriter{
			Filepath:     logfile,
			AutoFlushLag: 2,
			Cachesize:    0,
		}
	} else {
		// 默认标准输出
		mainlog.Logger.Out = os.Stdout
	}

	// 指定系统输出
	stdlog.SetOutput(mainlog.Writer())
}

// 新增一个日志对象 key-标识 c-配置 context-上下文
func NewLogger(key string, c *Config, context map[string]interface{}) Logger {
	// 查找是否已经注册
	_, ok := backends[key]
	if ok {
		return GetLogger(key)
	}
	// 新建日志对象
	log := logrus.New()
	// 读取配置 先读取C，再读取config
	if c == nil {
		c = &Config{}
	}
	// 级别
	log.SetLevel(logrus.DebugLevel) //输出级别固定，触发级别由配置决定
	acceptlevels := []logrus.Level{}
	toplevel := c.GetValue("level")
	if toplevel == "" {
		toplevel = config.GetValue(key, "level")
	}
	if toplevel == "" {
		toplevel = config.GetValue("common", "level")
	}
	if toplevel != "" {
		topint := int(getLevel(toplevel))
		for i := 0; i <= topint; i++ {
			acceptlevels = append(acceptlevels, logrus.Level(i))
		}
	}
	if len(acceptlevels) == 0 {
		for i := 0; i <= int(logrus.DebugLevel); i++ {
			acceptlevels = append(acceptlevels, logrus.Level(i))
		}
	}

	// 格式
	format := c.GetValue("formatter")
	if format == "" {
		format = config.GetValue(key, "formatter")
	}
	if format == "" {
		format = config.GetValue("common", "formatter")
	}
	switch format {
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: LOG_TIME_FORMAT,
		}
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: LOG_TIME_FORMAT,
		}
	case "pretty":
		log.Formatter = &PrettyFormatter{}
	default:
		// 默认json
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: LOG_TIME_FORMAT,
		}
	}

	// 输出 logfile 不会使用common的配置，而会直接输出到标准输出
	logpath := c.GetValue("logpath")
	if logpath == "" {
		logpath = config.GetValue(key, "logpath")
	}
	if logpath == "" {
		logpath = config.GetValue("common", "logpath")
	}
	if logpath == "" {
		logpath = "."
	}

	logfile := c.GetValue("logfile")

	if logfile == "" {
		logfile = config.GetValue(key, "logfile")
	}
	if logfile != "" {
		logfile, _ := getAbs(logpath + "/" + logfile)
		log.Out = FileWriter{
			Filepath:     logfile,
			AutoFlushLag: 2,
			Cachesize:    0,
		}
	} else {
		// 默认标准输出
		log.Out = os.Stdout
	}
	// 初始化信道 100个缓冲区
	ch := make(chan message, 100)
	// 组装worker
	w := worker{
		ch:    &ch,
		entry: log.WithFields(logrus.Fields(context)),
	}
	backends[key] = &w
	// 启动worker
	go w.work()

	// 注册hook
	hook := Modulehook{
		Key: key,
		Ls:  acceptlevels,
	}
	mainlog.Logger.AddHook(hook)

	return GetLogger(key)
}

// 获得日志对象
func GetLogger(name string) Logger {
	if mainlog == nil {
		Init("common/config/log.json")
	}
	if name == "" {
		return Logger{Entry: mainlog}
	} else {
		return Logger{Entry: mainlog.WithFields(logrus.Fields{ROUTE_KEY: name})}
	}
}

// 获取配置信息
func GetConfig() *Config {
	return config
}
