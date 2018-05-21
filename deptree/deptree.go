package deptree

// DepTree 组织架构树操作接口
// 说明：依赖包 - gopkg.in/ldap.v2
//             - github.com/golibs/uuid
type DepTree interface {
	// AddOrgNode 新增组织节点 node 节点信息 需包含Mid Pid(顶级节点可省略) Name(同一节点下需保证唯一) 信息 Id可选
	// 返回节点id
	AddOrgNode(node OrgNode) (string, error)
	// ModifyOrgNode 修改组织节点 node需包含完整的Mid Id 信息 只能更新Name信息，传空不更新
	ModifyOrgNode(node OrgNode) error
	// DelOrgNode 删除组织节点 id-节点ID
	DelOrgNode(mid string, id string) error
	// AddLeafNode 新增叶子节点
	AddLeafNode(leaf LeafNode) error
	// ModifyLeafNode 修改叶子节点
	ModifyLeafNode(leaf LeafNode) error
	// DelLeafNode 删除叶子节点 pid-父节点ID uid-uid
	DelLeafNode(mid string, pid string, uid string) error
	// GetLeafNodes 根据mid，pid, uid取叶子节点信息
	GetLeafNodes(mid string, pid string, uid string) ([]LeafNode, error)
	// GetLeafNodesByOrg 根据组织节点，取所有叶子节点信息
	GetLeafNodesByOrg(mid string, pid string) ([]LeafNode, error)
	// GetOrgNode 取组织节点信息
	GetOrgNode(mid string, id string) (*OrgNode, error)
	// GetOrgNode 取组织节点下的全部节点ID dept
	GetOrgNodesByOrg(mid string, pid string, dept int) ([]OrgNode, error)
	// GetSubTree 取树形结构 id表示树根的ID（商户ID、分公司ID或部门ID）
	GetSubTree(mid string, id string) (*OrgTree, error)
	// GetUsersByPosition 根据岗位查询UID列表
	GetUsersByPosition(mid string, pid string, positionid string) ([]LeafNode, error)
	// GetParents 根据节点id获得全部父节点信息 从近到远
	GetParents(mid string, id string) ([]OrgNode, error)
}

// NewTree 目前仅返回ldap结构树对象，扩展视后续需求开发
func NewTree(config map[string]interface{}) DepTree {
	host := config["Host"]
	port := config["Port"]
	base := config["Base"]
	user := config["User"]
	passwd := config["Password"]
	if host == nil || port == nil ||
		base == nil || user == nil ||
		passwd == nil {
		return nil
	}
	return &ldapDepTree{
		host:   host.(string),
		port:   int(port.(float64)),
		base:   base.(string),
		user:   user.(string),
		passwd: passwd.(string),
	}
}
