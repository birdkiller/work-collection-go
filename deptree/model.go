package deptree

// Node 组织节点
type OrgNode struct {
	Mid       string // 商户ID
	Pid       string // 父节点ID
	Id        string // ID
	Type      int    // 1-商户 2-分公司 3-部门
	Name      string // 名称
	IsDefault bool   // 是否默认生成
}

// Leaf 叶节点
type LeafNode struct {
	Mid       string   // 商户ID
	Pid       string   // 父节点ID
	Sid       string   // staff id
	Uid       string   // uid
	Positions []string // 岗位ID列表
}

// OrgTree组织树
type OrgTree struct {
	OrgNode             // 本节点属性
	SubTrees []OrgTree  // 子树
	SubLeafs []LeafNode // 子叶
}

var TYPE_SHOP int = 1
var TYPE_SUBCOM int = 2
var TYPE_DEP int = 3
