package logger

import (
	"github.com/sirupsen/logrus"
)

const ROUTE_KEY string = "*routekey"

type message struct {
	message string
	level   logrus.Level
	data    logrus.Fields
}

// Modulehook an impletement of logrus.Hook
type Modulehook struct {
	Key string
	Ls  []logrus.Level
}

func (self Modulehook) Levels() []logrus.Level {
	return self.Ls
}

func (self Modulehook) Fire(entry *logrus.Entry) error {
	key, _ := entry.Data[ROUTE_KEY]
	if key != nil && key.(string) == self.Key {
		worker, ok := backends[key.(string)]
		if ok {
			c := worker.ch
			*c <- message{
				data:    entry.Data,
				level:   entry.Level,
				message: entry.Message,
			}
		}
	}

	return nil
}
