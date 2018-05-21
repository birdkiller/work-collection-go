package logger

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// getLevel由字符串配置获得level对象
func getLevel(ls string) logrus.Level {
	for _, l := range logrus.AllLevels {
		if l.String() == ls {
			return l
		}
	}
	// 默认debug模式
	return logrus.DebugLevel
}

// 取绝对路径(相对可执行文件)
func getAbs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	mainpath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Dir(mainpath) + "/" + path)
}

// 保证目录存在 (没有就创建)
func ensurePath(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0777)
	} else {
		return err
	}
}

// 获取goid
func getGoid() string {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	if n <= 0 {
		return ""
	}
	fields := strings.Split(string(buf), " ")
	if len(fields) >= 2 {
		return fields[1]
	} else {
		return ""
	}
}
