package wyclient

import (
	"encoding/json"
	"saas/common/utils/im"
	"strconv"
)

const (
	WY_MESSAGE_TYPE_TEXT = iota
	WY_MESSAGE_TYPE_PIC
	WY_MESSAGE_TYPE_CUSTOM = 100
)

// WyMessage 网易IM消息
type WyMessage struct {
}

func (self *WyMessage) Format() interface{} {
	str, _ := json.Marshal(self)
	return string(str)
}

// 获得一个新message对象
func newMessage(t int) im.Message {
	var o im.Message
	switch t {
	case WY_MESSAGE_TYPE_TEXT:
		o = &TextMessage{}
	case WY_MESSAGE_TYPE_PIC:
		o = &PicMessage{}
	case WY_MESSAGE_TYPE_CUSTOM:
		o = &CustomerMessage{}
	default:
		return nil
	}
	return o
}

// 文本消息
type TextMessage struct {
	WyMessage
	Msg string
}

func (self *TextMessage) GetType() string {
	return strconv.Itoa(WY_MESSAGE_TYPE_TEXT)
}

// 图片消息
type PicMessage struct {
	WyMessage
	Name string
	Md5  string
	Url  string
	W    string
	H    string
	Size string
}

func (self *PicMessage) GetType() string {
	return strconv.Itoa(WY_MESSAGE_TYPE_PIC)
}

// 自定义消息
type CustomerMessage struct {
	WyMessage
	Body interface{}
}

func (self *CustomerMessage) GetType() string {
	return strconv.Itoa(WY_MESSAGE_TYPE_CUSTOM)
}

func (self *CustomerMessage) Format() interface{} {
	str, _ := json.Marshal(self.Body)
	return string(str)
}
