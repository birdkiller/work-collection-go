package im

// User IM用户结构体
type User struct {
	Id        string
	Nickname  string // 昵称
	Staffname string // 员工姓名
	Icon      string // 头像
	Auth      string
	Card      UserCard // 用户名片
}

// Group 群组结构体
type Group struct {
	Id      string
	Name    string
	Icon    string
	Owner   User
	Admins  []User
	Members []User
}

const (
	POSITION_NAME_CUSTOMER   = "业主"
	POSITION_NAME_DESIGNER   = "设计师"
	POSITION_NAME_MANAGER    = "项目经理"
	POSITION_NAME_SUPERVISOR = "工程监理"
	POSITION_TYPE_CUSTOMER   = iota
	POSITION_TYPE_DESIGNER
	POSITION_TYPE_MANAGER
	POSITION_TYPE_SUPERVISOR
)

// UserCard 群内用户名片
type UserCard struct {
	Uid          string
	Type         int    // 用户类型、自定义、可参考 saas/common/core/modelbase.ROLE_TYPE_USER
	Alias        string // 群内别名
	Position     string // 群内岗位信息
	PositionType int    // 群内岗位编号
}

const (
	CHANGE_OWNER_TYPE_LEAVE int = iota + 1
	CHANGE_OWNER_TYPE_STAY
)

const (
	MESSAGE_TARGET_USER int = iota
	MESSAGE_TARGET_GROUP
)

// Message IM通用消息接口
type Message interface {
	// 获取消息类型
	GetType() string
	// 格式化
	Format() interface{}
}
