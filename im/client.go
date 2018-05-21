package im

import (
	"fmt"
)

var clientContainer = make(map[string]ImClient)

// ImClient IM客户端接口
type ImClient interface {
	// GetId 获取本客户端ID
	GetId() string
	// Init 实现加载配置
	Init(config map[string]interface{})
	// Register 实现注册用户
	Register(uid string, nick string, extmsg map[string]string) (*User, error)
	// SendMessage 实现发送消息
	SendMessage(
		from string,
		to []string,
		targettype int, // 0对个人 1对群组
		message Message,
		atlist []string,
		extmsg map[string]string) error

	// SendSysNotice 实现发送系统通知
	SendSysNotice(
		from string,
		to []string,
		targettype int, // 0对个人 1对群组
		message interface{},
		extmsg map[string]string) error

	// CreateGroup 创建群组
	CreateGroup(name string,
		owner string,
		members []string,
		extmsg map[string]string) (string, error)

	// AddToGroup 将用户加入群
	AddToGroup(groupid string, owner string, members []string, extmsg map[string]interface{}) error
	// KickFromGroup 将用户踢出群
	KickFromGroup(groupid string, owner string, members []string, extmsg map[string]interface{}) error
	// DeleteGroup 删除群
	DeleteGroup(groupid string, owner string, extmsg map[string]interface{}) error
	// ChangeOwner 移交群主 leave - 移交后是否退群，1离开，2留下
	ChangeOwner(groupid string, owner string, newowner string, leave int) error
	// GetGroupInfo获取聊天群信息
	GetGroupInfo(groupid string, extmsg map[string]interface{}) (*Group, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(uid string) (*User, error)
	// SetGroupUserCard 设置群用户名片
	SetGroupUserCard(groupid string, owner string, member string, card UserCard) error
	// UpdateGroupInfo 更新群信息
	UpdateGroupInfo(groupid string, owner string, name string, icon string, extmsg map[string]string) error
	// 设置扩展功能函数
	//SetExtend(key string, caller *func(args map[string]interface{}))
	// 获取扩展功能函数
	GetExtend(key string) *func(args ...interface{})

	// 新建一个消息 t-类型
	NewMessage(t string) Message
}

// Register 注册一个IM客户端
func Register(key string, c ImClient) error {
	_, exsist := clientContainer[key]
	if exsist {
		return fmt.Errorf(fmt.Sprintf("已存在key:%v", key))
	}
	clientContainer[key] = c
	return nil
}

// GetClient 获得一个IM客户端对象
func GetClient(key string, config map[string]interface{}) (ImClient, bool) {
	o, exsist := clientContainer[key]
	if exsist != true {
		return nil, exsist
	}
	o.Init(config)
	return o, true
}
