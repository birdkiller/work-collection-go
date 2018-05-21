package deptree

import (
	"fmt"
	"log"
	//"log"
	"strconv"

	//ldap "gopkg.in/ldap.v2"
	ldap "github.com/go-ldap/ldap"
)

// ldapDepTree DepTree的ldap实现，通过NewTree获得
type ldapDepTree struct {
	host   string
	port   int
	base   string
	user   string
	passwd string
}

/****************** For Comment *****************
// LdapOrgMap
var LdapOrgMap = map[string]string{
	"l": "Pid",
	"physicalDeliveryOfficeName": "Name",
	"businessCategory":           "Type",
	"street":                     "Mid",
	"st":                         "Id",
	"description":                "IsDefault",
}

// LdapStaffMap
var LdapStaffMap = map[string]string{
	"uid":              "Uid",
	"employeeNumber":   "Sid",
	"l":                "Pid",
	"o":                "Mid",
	"title":      "Positions",
}
************************************************/

// 从ldap.entry转化成orgnode
func ldap2orgnode(entry *ldap.Entry, node *OrgNode) {
	node.Mid = entry.GetAttributeValue("street")
	node.Pid = entry.GetAttributeValue("l")
	node.Id = entry.GetAttributeValue("st")
	node.Name = entry.GetAttributeValue("ou")
	node.Type, _ = strconv.Atoi(entry.GetAttributeValue("businessCategory"))
	node.IsDefault, _ = strconv.ParseBool(entry.GetAttributeValue("description"))

}

// 从ldap.entry转化成leafnode
func ldap2leafnode(entry *ldap.Entry, node *LeafNode) {
	node.Sid = entry.GetAttributeValue("employeeNumber")
	node.Mid = entry.GetAttributeValue("o")
	node.Pid = entry.GetAttributeValue("l")
	node.Uid = entry.GetAttributeValue("uid")
	node.Positions = entry.GetAttributeValues("title")

}

// ldapDepTree.connect 私有函数 连接ldap服务
func (self *ldapDepTree) connect() (*ldap.Conn, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", self.host, self.port))
	if err != nil {
		return nil, err
	}
	err = l.Bind(self.user, self.passwd)
	if err != nil {
		l.Close()
		return nil, err
	}
	return l, nil
}

// getTopTree 根据mid获取顶级树的dn
func (self *ldapDepTree) getTopTreeDn(mid string, conn *ldap.Conn) (string, error) {
	//log.Println(self.base)
	searchReq := ldap.NewSearchRequest(self.base, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(Objectclass=organizationalUnit)(st=%s))", mid), []string{"dn"}, nil)
	sr, err := conn.Search(searchReq)
	if err != nil {
		return "", err
	}
	if sr == nil || len(sr.Entries) == 0 {
		return "", fmt.Errorf("Can't find the top tree with this mid: %s", mid)
	}
	return sr.Entries[0].DN, nil
}

// getSubTree 根据顶级树dn和子树id获取子树（非叶子）的dn
func (self *ldapDepTree) getSubTreeDn(tree_dn string, id string, conn *ldap.Conn) (string, error) {
	searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(Objectclass=organizationalUnit)(st=%s))", id), []string{"dn"}, nil)
	sr, err := conn.Search(searchReq)
	if err != nil {
		return "", err
	}
	if sr == nil || len(sr.Entries) == 0 {
		return "", fmt.Errorf("Can't find the node with this id: %s", id)
	}
	return sr.Entries[0].DN, nil
}

// delTree 根据dn删除节点(递归)
func (self *ldapDepTree) delTree(dn string, conn *ldap.Conn) error {
	// 搜索该节点下层的叶子节点
	searchReq := ldap.NewSearchRequest(dn, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false, "(ObjectClass=posixAccount)", []string{"dn"}, nil)
	sr, err := conn.Search(searchReq)
	// 删除叶子
	for _, e := range sr.Entries {
		delReq := ldap.NewDelRequest(e.DN, nil)
		err = conn.Del(delReq)
		if err != nil {
			return err
		}
	}

	// 搜索该节点下首层非叶子节点
	searchReq = ldap.NewSearchRequest(dn, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false, "(ObjectClass=organizationalUnit)", []string{"dn"}, nil)
	sr, err = conn.Search(searchReq)
	// 删除子节点
	for _, e := range sr.Entries {
		err = self.delTree(e.DN, conn)
		if err != nil {
			return err
		}
	}
	// 删除自己
	delReq := ldap.NewDelRequest(dn, nil)
	err = conn.Del(delReq)
	return err
}

// getSubTree 根据ldap节点获取树信息(递归)
func (self *ldapDepTree) getSubTree(entry *ldap.Entry, conn *ldap.Conn) OrgTree {
	ret := OrgTree{
		SubTrees: []OrgTree{},
		SubLeafs: []LeafNode{},
	}
	ldap2orgnode(entry, &ret.OrgNode)

	// 搜索该节点下层的叶子节点
	searchReq := ldap.NewSearchRequest(entry.DN, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false, "(ObjectClass=posixAccount)",
		[]string{"uid", "l", "o", "title", "employeeNumber"}, nil)
	sr, _ := conn.Search(searchReq)
	// 处理叶子
	for _, e := range sr.Entries {
		leaf := LeafNode{}
		ldap2leafnode(e, &leaf)
		ret.SubLeafs = append(ret.SubLeafs, leaf)
	}

	// 搜索该节点下首层非叶子节点
	searchReq = ldap.NewSearchRequest(entry.DN, ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false, "(ObjectClass=organizationalUnit)",
		[]string{"l", "ou", "businessCategory", "street", "st", "description"},
		nil)
	sr, _ = conn.Search(searchReq)
	// 处理子节点
	for _, e := range sr.Entries {
		subtree := self.getSubTree(e, conn)
		ret.SubTrees = append(ret.SubTrees, subtree)
	}
	return ret
}

// AddOrgNode 新建组织节点
func (self *ldapDepTree) AddOrgNode(node OrgNode) (string, error) {
	// 获取ID
	id := node.Id
	mid := node.Mid
	name := node.Name
	pid := node.Pid
	var dn string // 待插入节点路径标识

	conn, err := self.connect()
	if conn == nil {
		return "", err
	}
	defer conn.Close()

	if pid == "" {
		//插入顶级节点(ID使用传入的mid)
		dn = fmt.Sprintf("ou=%s,%s", name, self.base)
		id = mid
	} else {
		// 搜索mid对应的树
		tree_dn, err := self.getTopTreeDn(mid, conn)
		if err != nil {
			return "", err
		}
		// 根据树dn和fid搜索父节点dn
		var parent_dn string
		if pid == mid {
			parent_dn = tree_dn
		} else {
			parent_dn, err = self.getSubTreeDn(tree_dn, pid, conn)
			if err != nil {
				return "", err
			}
		}
		// 生成dn
		if id == "" {
			id = GetId()
		}
		dn = fmt.Sprintf("ou=%s,%s", name, parent_dn)
		if err != nil {
			return "", err
		}
	}
	// 插入
	addReq := ldap.NewAddRequest(dn)
	addReq.Attribute("Objectclass", []string{"organizationalUnit"})
	if pid != "" {
		addReq.Attribute("l", []string{pid})
	}
	addReq.Attribute("street", []string{node.Mid})
	addReq.Attribute("ou", []string{node.Name})
	addReq.Attribute("businessCategory", []string{strconv.Itoa(node.Type)})
	addReq.Attribute("st", []string{id})
	addReq.Attribute("description", []string{strconv.FormatBool(node.IsDefault)})
	err = conn.Add(addReq)
	if err != nil {
		return "", err
	}
	return id, err
}

// ModifyOrgNode 修改组织信息
func (self *ldapDepTree) ModifyOrgNode(node OrgNode) error {
	id := node.Id
	mid := node.Mid
	if id == "" || mid == "" {
		return fmt.Errorf("invalid id or mid [%s,%s]", id, mid)
	}
	conn, err := self.connect()
	if conn == nil {
		return err
	}
	defer conn.Close()

	var dn string // 待更新节点路径标识
	if id == mid {
		// 顶级节点
		dn = fmt.Sprintf("ou=%s,%s", node.Name, self.base)
	} else {
		// 搜索mid对应的树
		tree_dn, err := self.getTopTreeDn(mid, conn)
		// 获得需更新节点的dn
		dn, err = self.getSubTreeDn(tree_dn, id, conn)
		if err != nil {
			return err
		}
	}

	newdn := fmt.Sprintf("ou=%s", node.Name)
	modDNReq := ldap.NewModifyDNRequest(dn, newdn, true, "")
	err = conn.ModifyDN(modDNReq)
	return err
}

// DelOrgNode 删除组织信息
func (self *ldapDepTree) DelOrgNode(mid string, id string) error {
	if id == "" || mid == "" {
		return fmt.Errorf("invalid id or mid [%s,%s]", id, mid)
	}
	conn, err := self.connect()
	if conn == nil {
		return err
	}
	defer conn.Close()

	var dn string // 待更新节点路径标识
	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return err
	}
	// 获得需删除节点的dn
	dn, err = self.getSubTreeDn(tree_dn, id, conn)
	if err != nil {
		return err
	}

	err = self.delTree(dn, conn)
	return err
}

// AddLeafNode 新增叶子节点(角色)
func (self *ldapDepTree) AddLeafNode(leaf LeafNode) error {
	mid := leaf.Mid
	pid := leaf.Pid
	uid := leaf.Uid
	sid := leaf.Sid
	positions := leaf.Positions
	conn, err := self.connect()
	if conn == nil {
		return err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return err
	}
	var parent_dn string
	if pid == mid {
		parent_dn = tree_dn
	} else {
		// 根据树dn和fid搜索父节点dn
		parent_dn, err = self.getSubTreeDn(tree_dn, pid, conn)
		if err != nil {
			return err
		}
	}
	// 生成dn
	dn := fmt.Sprintf("cn=%s,%s", uid, parent_dn)

	addReq := ldap.NewAddRequest(dn)
	addReq.Attribute("Objectclass", []string{"inetOrgPerson", "posixAccount"})
	addReq.Attribute("sn", []string{uid})
	addReq.Attribute("uid", []string{uid})
	addReq.Attribute("uidNumber", []string{"0"})
	addReq.Attribute("employeeNumber", []string{sid})
	addReq.Attribute("cn", []string{uid})
	addReq.Attribute("homeDirectory", []string{"/"})
	addReq.Attribute("gidNumber", []string{"0"})
	addReq.Attribute("l", []string{pid})
	addReq.Attribute("o", []string{mid})
	addReq.Attribute("street", []string{mid})
	// 如果包含角色数据
	if positions != nil {
		addReq.Attribute("title", positions)
	}

	err = conn.Add(addReq)
	return err

}

// ModifyLeafNode 修改叶子节点(角色信息) attrs -- {"positions":[]string{r1,r2,r3}}
func (self *ldapDepTree) ModifyLeafNode(leaf LeafNode) error {
	mid := leaf.Mid
	pid := leaf.Pid
	uid := leaf.Uid
	positions := leaf.Positions
	if positions == nil {
		// 如果无修改角色列表，则直接返回
		return nil
	}
	conn, err := self.connect()
	if conn == nil {
		return err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return err
	}
	// 根据树dn和fid搜索父节点dn
	var parent_dn string
	if pid == mid {
		parent_dn = tree_dn
	} else {
		parent_dn, err = self.getSubTreeDn(tree_dn, pid, conn)
		if err != nil {
			return err
		}
	}
	// 生成dn
	dn := fmt.Sprintf("cn=%s,%s", uid, parent_dn)

	modReq := ldap.NewModifyRequest(dn)
	modReq.Replace("title", positions)
	err = conn.Modify(modReq)
	return err

}

// DelStaff 删除员工(角色)
func (self *ldapDepTree) DelLeafNode(mid string,
	pid string,
	uid string) error {
	conn, err := self.connect()
	if conn == nil {
		return err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return err
	}
	// 根据树dn和fid搜索父节点dn
	var parent_dn string
	if pid == mid {
		parent_dn = tree_dn
	} else {
		parent_dn, err = self.getSubTreeDn(tree_dn, pid, conn)
		if err != nil {
			return err
		}
	}
	// 生成dn
	dn := fmt.Sprintf("cn=%s,%s", uid, parent_dn)

	delReq := ldap.NewDelRequest(dn, nil)
	err = conn.Del(delReq)
	return err
}

// GetLeafNodes
func (self *ldapDepTree) GetLeafNodes(mid string, oid string, uid string) ([]LeafNode, error) {
	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}

	// 根据oid搜索该树下的orgnode
	searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=organizationalUnit)(st=%s))", oid),
		[]string{"l", "ou", "businessCategory", "street", "st", "description"},
		nil)
	sr, err := conn.Search(searchReq)
	if err != nil || len(sr.Entries) <= 0 {
		return nil, err
	}
	org_dn := sr.Entries[0].DN

	// 根据uid搜索该树下的全部结果集
	searchReq = ldap.NewSearchRequest(org_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=posixAccount)(uid=%s))", uid),
		[]string{"uid", "l", "o", "title", "employeeNumber"}, nil)
	sr, err = conn.Search(searchReq)
	if err != nil {
		return nil, err
	}
	// 叶子信息
	ret := []LeafNode{}
	for _, e := range sr.Entries {
		oneleaf := LeafNode{}
		ldap2leafnode(e, &oneleaf)
		ret = append(ret, oneleaf)
	}
	return ret, nil
}

// GetLeafNodesByOrg
func (self *ldapDepTree) GetLeafNodesByOrg(mid string, oid string) ([]LeafNode, error) {
	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}
	// 根据oid搜索该树下的orgnode
	searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=organizationalUnit)(st=%s))", oid),
		[]string{"l", "ou", "businessCategory", "street", "st", "description"},
		nil)
	sr, err := conn.Search(searchReq)
	if err != nil || len(sr.Entries) <= 0 {
		return nil, err
	}
	org_dn := sr.Entries[0].DN

	// 搜索该org下的全部leafnode
	searchReq = ldap.NewSearchRequest(org_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, "(ObjectClass=posixAccount)",
		[]string{"uid", "l", "o", "title", "employeeNumber"}, nil)
	sr, err = conn.Search(searchReq)
	if err != nil {
		return nil, err
	}

	ret := []LeafNode{}
	for _, e := range sr.Entries {

		oneleaf := LeafNode{}
		ldap2leafnode(e, &oneleaf)
		ret = append(ret, oneleaf)
	}
	return ret, nil
}

// GetOrgNode
func (self *ldapDepTree) GetOrgNode(mid string, id string) (*OrgNode, error) {
	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}
	// 根据id搜索该树下的全部结果集
	searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=organizationalUnit)(st=%s))", id),
		[]string{"l", "ou", "businessCategory", "street", "st", "description"},
		nil)
	sr, err := conn.Search(searchReq)
	if err != nil || len(sr.Entries) <= 0 {
		return nil, err
	}

	org := OrgNode{}
	ldap2orgnode(sr.Entries[0], &org)
	return &org, nil
}

// GetOrgNodesByOrg
func (self *ldapDepTree) GetOrgNodesByOrg(mid string, oid string, dept int) ([]OrgNode, error) {
	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}
	// 根据oid搜索该树下的orgnode
	searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=organizationalUnit)(st=%s))", oid),
		[]string{"l", "ou", "businessCategory", "street", "st", "description"},
		nil)
	sr, err := conn.Search(searchReq)
	if err != nil || len(sr.Entries) <= 0 {
		return nil, err
	}
	org_dn := sr.Entries[0].DN

	// 搜索该org下的全部orgnode
	sign := ldap.ScopeWholeSubtree
	if dept == 1 {
		sign = ldap.ScopeSingleLevel
	}
	searchReq = ldap.NewSearchRequest(org_dn, sign,
		ldap.NeverDerefAliases,
		0, 0, false, "(ObjectClass=organizationalUnit)",
		[]string{"l", "ou", "businessCategory", "street", "st", "description"}, nil)
	sr, err = conn.Search(searchReq)
	if err != nil {
		return nil, err
	}

	ret := []OrgNode{}
	for _, e := range sr.Entries {

		oneorg := OrgNode{}
		ldap2orgnode(e, &oneorg)
		ret = append(ret, oneorg)
	}
	return ret, nil
}

// GetSubTree 取树形结构
func (self *ldapDepTree) GetSubTree(mid string, id string) (*OrgTree, error) {
	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}
	// 根据id搜索该树下的组织节点
	searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=organizationalUnit)(st=%s))", id),
		[]string{"l", "ou", "businessCategory", "street", "st", "description"},
		nil)
	sr, err := conn.Search(searchReq)
	if err != nil || len(sr.Entries) <= 0 {
		return nil, err
	}
	// 根据命中的节点取出子树
	subtree := self.getSubTree(sr.Entries[0], conn)
	return &subtree, nil
}

// GetUsersByPosition 根据角色查询UID列表
func (self *ldapDepTree) GetUsersByPosition(mid string, pid string, positionid string) ([]LeafNode, error) {
	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}

	// 根据树tree_dn和pid搜索父节点dn
	var parent_dn string
	if mid != pid {
		parent_dn, err = self.getSubTreeDn(tree_dn, pid, conn)
		if err != nil {
			return nil, err
		}
	} else {
		parent_dn = tree_dn
	}

	// 根据uid搜索该树下的全部结果集
	searchReq := ldap.NewSearchRequest(parent_dn, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false, fmt.Sprintf("(&(ObjectClass=posixAccount)(title=%s))", positionid),
		[]string{"uid", "l", "o", "title", "employeeNumber"}, nil)
	sr, err := conn.Search(searchReq)
	if err != nil {
		return nil, err
	}
	// 添加叶子信息
	ret := []LeafNode{}
	for _, e := range sr.Entries {
		oneleaf := LeafNode{}
		ldap2leafnode(e, &oneleaf)
		ret = append(ret, oneleaf)
	}
	return ret, nil
}

// GetParents 根据节点id获得全部父节点信息(路径) 从近到远
func (self *ldapDepTree) GetParents(mid string, id string) ([]OrgNode, error) {
	nodelist := []OrgNode{}

	conn, err := self.connect()
	if conn == nil {
		return nil, err
	}
	defer conn.Close()

	// 搜索mid对应的树
	tree_dn, err := self.getTopTreeDn(mid, conn)
	if err != nil {
		return nil, err
	}
	// 根据id搜索该树下的组织节点
	searchid := id
	for {
		log.Println(searchid)
		searchReq := ldap.NewSearchRequest(tree_dn, ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0, 0, false, fmt.Sprintf("(&(ObjectClass=organizationalUnit)(st=%s))", searchid),
			[]string{"l", "ou", "businessCategory", "street", "st", "description"},
			nil)
		sr, err := conn.Search(searchReq)
		if err != nil || len(sr.Entries) <= 0 {
			return nil, err
		}
		entry := sr.Entries[0]
		node := OrgNode{}
		ldap2orgnode(entry, &node)
		nodelist = append(nodelist, node)
		if node.Id == mid || node.Pid == "" {
			// 已经搜索到根目录，退出循环
			break
		}
		searchid = node.Pid
	}
	return nodelist, nil

}
