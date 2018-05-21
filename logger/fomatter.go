package logger

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	LOG_TIME_FORMAT = "2006-01-02 15:04:05.999"
)

// PrettyFormatter 比较漂亮的格式化
type PrettyFormatter struct {
}

// 美化输出
func (self *PrettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	level := entry.Level.String()
	timestamp := entry.Time.Format(LOG_TIME_FORMAT)
	message := entry.Message
	fields := []string{}
	for k, v := range entry.Data {
		if k != ROUTE_KEY {
			// 不显示routekey
			fields = append(fields, fmt.Sprintf("%s:%v", k, v))
		}
	}

	output := fmt.Sprintf("%s [%s] %s data: %s\n",
		timestamp,
		level,
		message,
		"{"+strings.Join(fields, "; ")+"}")
	return []byte(output), nil
}
