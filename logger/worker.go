package logger

import (
	"github.com/sirupsen/logrus"
)

type worker struct {
	ch    *chan message
	entry *logrus.Entry
}

func (self worker) work() {
	if self.ch == nil || self.entry == nil {
		panic("日志worker初始化错误！")
	}
	for {
		m := <-*self.ch
		switch m.level {
		case logrus.ErrorLevel:
			self.entry.WithFields(m.data).Errorln(m.message)
		case logrus.DebugLevel:
			self.entry.WithFields(m.data).Debug(m.message)
		case logrus.InfoLevel:
			self.entry.WithFields(m.data).Println(m.message)
		case logrus.FatalLevel:
			m.data["level-upgrade"] = m.level
			self.entry.WithFields(m.data).Errorln(m.message)
		case logrus.PanicLevel:
			m.data["level-upgrade"] = m.level
			self.entry.WithFields(m.data).Errorln(m.message)
		default:
			self.entry.WithFields(m.data).Println(m.message)
		}
	}
}

func GetWorker(key string) *worker {
	return nil
}
