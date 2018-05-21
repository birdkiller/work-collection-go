package deptree

import (
	"strings"

	"github.com/golibs/uuid"
)

var idChannel chan string

// SetId 循环设置ID等待使用
func SetId(c chan string) {
	defer close(c)

	for {
		uuid := strings.Replace(uuid.Rand().Hex(), "-", "", -1)
		c <- uuid
	}
}

// GetId 获取一个ID
func GetId() string {
	uuid := <-idChannel
	return uuid
}

func init() {
	idChannel = make(chan string, 10)

	go SetId(idChannel)
}
