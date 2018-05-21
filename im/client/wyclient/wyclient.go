package wyclient

import (
	"crypto/sha1"
	"encoding/json"
	//log "envkeeper/core/logger"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"saas/common/utils/im"
)

type WyImClient struct {
	host     string
	appkey   string
	secret   string
	isconfig bool
	im.ImClient
}

type Rsp struct {
	Code float64
	Info interface{}
	Desc string
}

type CreateGroupRsp struct {
	Rsp
	Tid    string
	Faccid interface{}
}

type AddToGroupRsp struct {
	Rsp
	Faccid map[string]interface{}
}

type GetGroupInfoRsp struct {
	Rsp
	Tinfo map[string]interface{}
}

type GetUserInfoRsp struct {
	Rsp
	Uinfos []interface{}
}

// WyImClient.GetId 返回wyclient
func (self *WyImClient) GetId() string {
	return "wyclient"
}

// WyImClient.Init 如果已配置则返回，否则加载配置
// c map[string]string:
//   host string
//   appkey string
//   appsecret string
func (self *WyImClient) Init(c map[string]interface{}) {
	if self.isconfig == true {
		return
	}
	config := c
	host, ok := config["host"]
	if ok != true {
		panic(fmt.Errorf("缺少IM配置项:host"))
	}
	self.host = "https://" + host.(string)
	appkey, ok := config["appkey"]
	if ok != true {
		panic(fmt.Errorf("缺少IM配置项:appkey"))
	}
	self.appkey = appkey.(string)
	secret, ok := config["appsecret"]
	if ok != true {
		panic(fmt.Errorf("缺少IM配置项:appsecret"))
	}
	self.secret = secret.(string)

	self.isconfig = true
}

// WyImClient.Register 注册网易用户
// extmsg: 需传入map[string]string
//         props - 自定义属性 json
//         icon - 头像url
//         token - 自定义token 不传会生成
//         sign - 签名
//         email - 电邮
//         birth - 生日
//         mobile - 手机
//         gender - 性别 0-未知 1-男 2-女
func (self *WyImClient) Register(uid string,
	nick string,
	extmsg map[string]string) (*im.User, error) {
	req := url.Values{}
	req.Set("accid", uidfilter(uid))
	req.Set("name", nick)
	//log.Println(nick)
	//rsp_rg := RegistUserInfo{}
	for k, m := range extmsg {
		req.Set(k, m)
	}
	rsp := Rsp{}
	self.post("nimserver/user/create.action", req, nil, &rsp)
	//log.Println(rsp)
	if rsp.Code == 200 {
		//log.Printf("%t\n", rsp.Info)
		rsp_info := rsp.Info.(map[string]interface{})
		token, _ := rsp_info["token"].(string)
		accid, _ := rsp_info["accid"].(string)
		name, _ := rsp_info["name"].(string)
		user := im.User{
			Auth:     token,
			Id:       accid,
			Nickname: name,
		}
		return &user, nil
	} else {
		log.Println(rsp.Desc)
		return nil, fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.SendMessage 发送消息
// extmsg:
func (self *WyImClient) SendMessage(
	from string,
	to []string,
	targettype int,
	message im.Message,
	atlist []string,
	extmsg map[string]string) error {
	req := url.Values{}
	rsp := Rsp{}
	for k, m := range extmsg {
		req.Set(k, m)
	}
	if atlist != nil && len(atlist) > 0 {
		atlist, _ := json.Marshal(atlist)
		req.Set("forcepushlist", string(atlist))
	}
	if len(to) > 1 && targettype == 0 {
		//如果是发给个人，且群发则调用群发接口
		req.Set("fromAccid", from)
		if len(to) > 500 {
			to = to[:500]
		}
		tolist, _ := json.Marshal(to)
		req.Set("toAccids", string(tolist))
		req.Set("body", message.Format().(string))
		req.Set("type", message.GetType())
		self.post("nimserver/msg/sendBatchMsg.action", req, nil, &rsp)
	} else {
		//否则调用单发接口（网易不支持群发给群组，故群组ID仅取第一个）
		req.Set("from", from)
		req.Set("ope", strconv.Itoa(targettype))
		req.Set("to", to[0])
		req.Set("body", message.Format().(string))
		req.Set("type", message.GetType())
		self.post("nimserver/msg/sendMsg.action", req, nil, &rsp)
	}

	if rsp.Code == 200 {
		return nil
	} else {
		return fmt.Errorf(rsp.Desc)
	}

}

// WyImClient.SendSysNotice 发送系统通知
// extmsg:
func (self *WyImClient) SendSysNotice(
	from string,
	to []string,
	targettype int,
	message interface{},
	extmsg map[string]string) error {
	req := url.Values{}
	rsp := Rsp{}
	bbody, err := json.Marshal(message)
	if err != nil {
		return err
	}
	body := string(bbody)
	if len(to) > 1 && targettype == 0 {
		//如果是发给个人，且群发则调用群发接口
		req.Set("fromAccid", from)
		if len(to) > 500 {
			to = to[:500]
		}
		tolist, _ := json.Marshal(to)
		req.Set("toAccids", string(tolist))
		req.Set("attach", body)
		self.post("nimserver/msg/sendBatchAttachMsg.action", req, nil, &rsp)
	} else {
		//否则调用单发接口（网易不支持群发给群组，故群组ID仅取第一个）
		req.Set("from", from)
		req.Set("msgtype", strconv.Itoa(targettype))
		req.Set("to", to[0])
		req.Set("attach", body)
		self.post("nimserver/msg/sendAttachMsg.action", req, nil, &rsp)
	}

	if rsp.Code == 200 {
		return nil
	} else {
		return fmt.Errorf(rsp.Desc)
	}

}

// WyImClient.CreataGroup 创建群组
// extmsg:
//   msg string 欢迎信息
//   magree string 0 入群无需被拉人同意，1 需被拉人同意
//   joinmode string 0 SDK无需验证，1 需验证
func (self *WyImClient) CreateGroup(name string,
	owner string,
	members []string,
	extmsg map[string]string) (string, error) {
	req := url.Values{}
	req.Set("tname", name)
	req.Set("owner", owner)
	memberlist, _ := json.Marshal(members)
	req.Set("members", string(memberlist))
	// 欢迎信息
	msg, exsist := extmsg["msg"]
	if exsist != true {
		msg = "欢迎入群！"
	}
	req.Set("msg", msg)
	// 拉人是否需要被邀请人同意
	magree, exsist := extmsg["magree"]
	if exsist != true {
		magree = "0"
	}
	req.Set("magree", magree)
	// SDK操作是否需要验证
	joinmode, exsist := extmsg["joinmode"]
	if exsist != true {
		joinmode = "0"
	}
	req.Set("joinmode", joinmode)

	// 群图片
	icon, exsist := extmsg["icon"]
	req.Set("icon", icon)

	// 群扩展信息
	custom, exsist := extmsg["custom"]
	if exsist == true {
		req.Set("custom", custom)
	}

	rsp := CreateGroupRsp{}
	self.post("nimserver/team/create.action", req, nil, &rsp)
	log.Println(rsp)
	if rsp.Code == 200 {
		//log.Printf("%t\n", rsp.Info)
		tid := rsp.Tid
		return tid, nil
	} else {
		log.Println(rsp.Desc)
		return "", fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.AddToGroup 拉人进群
// extmsg:
//   magree string 0入群无需被拉人同意，1需被拉人同意
//   msg string 入群欢迎信息
func (self *WyImClient) AddToGroup(groupid string,
	owner string,
	members []string,
	extmsg map[string]interface{}) error {
	req := url.Values{}
	rsp := AddToGroupRsp{}
	req.Set("tid", groupid)
	req.Set("owner", owner)
	memberlist, _ := json.Marshal(members)
	req.Set("members", string(memberlist))
	magree, exsist := extmsg["magree"]
	if exsist != true {
		magree = "0"
	}
	req.Set("magree", magree.(string))
	msg, exsist := extmsg["msg"]
	if exsist != true {
		msg = fmt.Sprintf("欢迎%v入群！", strings.Join(members, ","))
	}
	req.Set("msg", msg.(string))
	self.post("nimserver/team/add.action", req, nil, &rsp)
	//log.Println(rsp)
	if rsp.Code == 200 {
		//log.Printf("%t\n", rsp.Info)
		if rsp.Faccid != nil {
			faccid := rsp.Faccid
			log.Printf("拉人入群%v，部分用户未入群：[%v]，原因：%v\n",
				groupid,
				strings.Join(faccid["accid"].([]string), ","),
				faccid["msg"].(string))
		} else {
			log.Printf("拉人入群%v，用户[%v]，成功！\n",
				groupid,
				strings.Join(members, ","))
		}
		return nil
	} else {
		return fmt.Errorf(rsp.Desc)
	}

}

// WyImClient.KickFromGroup 踢人出群
func (self *WyImClient) KickFromGroup(groupid string,
	owner string,
	members []string,
	extmsg map[string]interface{}) error {
	req := url.Values{}
	var rsp Rsp
	req.Set("tid", groupid)
	req.Set("owner", owner)
	var succ int
	var errmsg string
	for _, member := range members {
		req.Set("member", member)
		self.post("nimserver/team/kick.action", req, nil, &rsp)
		if rsp.Code == 200 {
			log.Printf("从%v群踢出用户[%v]成功！\n", groupid, member)
			succ = 1
		} else {
			log.Printf("从%v群踢出用户[%v]失败！原因：%v\n",
				groupid,
				member,
				rsp.Desc)
			errmsg += rsp.Desc
		}

	}
	if succ == 1 {
		return nil
	} else {
		return fmt.Errorf(errmsg)
	}
}

// WyImClient.DeleteGroup 删除群组
func (self *WyImClient) DeleteGroup(groupid string,
	owner string,
	extmsg map[string]interface{}) error {
	req := url.Values{}
	rsp := Rsp{}
	req.Set("tid", groupid)
	req.Set("owner", owner)
	self.post("nimserver/team/remove.action", req, nil, &rsp)
	//log.Println(rsp)
	if rsp.Code == 200 {
		log.Printf("删除聊天群%v成功！\n", groupid)
		return nil
	} else {
		log.Printf("删除聊天群%v失败！原因：%v\n",
			groupid,
			rsp.Desc)
		return fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.ChangeOwner 移交群主
func (self *WyImClient) ChangeOwner(groupid string,
	owner string,
	newowner string,
	leave int) error {
	req := url.Values{}
	rsp := Rsp{}
	req.Set("tid", groupid)
	req.Set("owner", owner)
	req.Set("newowner", newowner)
	req.Set("leave", strconv.Itoa(leave))
	self.post("nimserver/team/changeOwner.action", req, nil, &rsp)
	//log.Println(rsp)
	if rsp.Code == 200 {
		log.Printf("移交群主%v:%v->%v成功！\n", groupid, owner, newowner)
		return nil
	} else {
		log.Printf("移交群%v群主失败！原因：%v\n",
			groupid,
			rsp.Desc)
		return fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.GetGroupInfo
func (self *WyImClient) GetGroupInfo(groupid string,
	extmsg map[string]interface{}) (*im.Group, error) {
	req := url.Values{}
	rsp := GetGroupInfoRsp{}
	req.Set("tid", groupid)
	self.post("nimserver/team/queryDetail.action", req, nil, &rsp)
	//log.Println(rsp)
	if rsp.Code == 200 {
		ginfo := im.Group{Id: strconv.Itoa((int)(rsp.Tinfo["tid"].(float64)))}
		if rsp.Tinfo["tname"] != nil {
			ginfo.Name = rsp.Tinfo["tname"].(string)
		}
		if rsp.Tinfo["icon"] != nil {
			ginfo.Icon = rsp.Tinfo["icon"].(string)
		}
		// owner
		owner := rsp.Tinfo["owner"].(map[string]interface{})
		log.Println(owner)
		cardstr, _ := owner["custom"]
		card := im.UserCard{}
		if cardstr != nil {
			json.Unmarshal([]byte(cardstr.(string)), &card)
		}
		ginfo.Owner = im.User{
			Id:       owner["accid"].(string),
			Nickname: owner["nick"].(string),
			Card:     card,
		}
		// admins
		ginfo.Admins = []im.User{}
		tadmins := rsp.Tinfo["admins"].([]interface{})
		for _, ta := range tadmins {
			tmp_a := ta.(map[string]interface{})
			cardstr, _ := tmp_a["custom"]
			card := im.UserCard{}
			if cardstr != nil {
				json.Unmarshal([]byte(cardstr.(string)), &card)
			}
			ginfo.Admins = append(ginfo.Admins, im.User{
				Id:       tmp_a["accid"].(string),
				Nickname: tmp_a["nick"].(string),
				Card:     card,
			})
		}
		// members
		ginfo.Members = []im.User{}
		tmembers := rsp.Tinfo["members"].([]interface{})
		for _, tm := range tmembers {
			tmp_m := tm.(map[string]interface{})
			cardstr, _ := tmp_m["custom"]
			card := im.UserCard{}
			if cardstr != nil {
				json.Unmarshal([]byte(cardstr.(string)), &card)
			}
			ginfo.Members = append(ginfo.Members, im.User{
				Id:       tmp_m["accid"].(string),
				Nickname: tmp_m["nick"].(string),
				Card:     card,
			})
		}

		return &ginfo, nil
	} else {
		log.Printf("获取聊天群%v信息失败！原因：%v\n",
			groupid,
			rsp.Desc)
		return nil, fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.GetUserInfo
func (self *WyImClient) GetUserInfo(uid string) (*im.User, error) {
	req := url.Values{}
	rsp := GetUserInfoRsp{}
	juids, _ := json.Marshal([]string{uid})
	req.Set("accids", string(juids))
	self.post("nimserver/user/getUinfos.action", req, nil, &rsp)
	//log.Println(rsp)
	if rsp.Code == 200 {
		if len(rsp.Uinfos) == 0 {
			return nil, fmt.Errorf("IM端不存在用户数据")
		}
		user := im.User{}
		uinfo := rsp.Uinfos[0].(map[string]interface{})
		user.Id = uinfo["accid"].(string)
		if uinfo["name"] != nil {
			user.Nickname = uinfo["name"].(string)
		}
		return &user, nil
	} else {
		log.Printf("获取用户%v信息失败！原因：%v\n",
			uid,
			rsp.Desc)
		return nil, fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.SetGroupUserCard
func (self *WyImClient) SetGroupUserCard(groupid string, owner string, member string, card im.UserCard) error {
	req := url.Values{}
	rsp := Rsp{}
	extmsg, _ := json.Marshal(card)
	req.Set("tid", groupid)
	req.Set("owner", owner)
	req.Set("accid", member)
	req.Set("nick", card.Alias)
	req.Set("custom", string(extmsg))
	self.post("nimserver/team/updateTeamNick.action", req, nil, &rsp)
	if rsp.Code == 200 {
		return nil
	} else {
		log.Printf("更新用户%v群名片信息失败！原因：%v\n",
			member,
			rsp.Desc)
		return fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.UpdateGroupInfo
func (self *WyImClient) UpdateGroupInfo(groupid string,
	owner string,
	name string,
	icon string,
	extmsg map[string]string) error {
	req := url.Values{}
	rsp := Rsp{}
	req.Set("tid", groupid)
	req.Set("owner", owner)
	if name != "" {
		req.Set("tname", name)
	}
	if icon != "" {
		req.Set("icon", icon)
	}
	for key, val := range extmsg {
		req.Set(key, val)
	}
	self.post("nimserver/team/update.action", req, nil, &rsp)
	if rsp.Code == 200 {
		return nil
	} else {
		log.Printf("更新群信息失败！原因：%v\n",
			rsp.Desc)
		return fmt.Errorf(rsp.Desc)
	}
}

// WyImClient.GetExtend 获得附加功能
func (self *WyImClient) GetExtend(key string) *func(args ...interface{}) {
	return nil
}

// WyImClient.NewMessage 新建消息
func (self *WyImClient) NewMessage(t string) im.Message {
	it, err := strconv.Atoi(t)
	if err != nil {
		return nil
	}
	return newMessage(it)
}

// WyImClient.checksum 网易校验：appsecret+nonce+curtime进行SHA1，结果保存至16进制小写字符串
func (self *WyImClient) checksum(nonce string, curtime uint32) string {
	str := self.secret + nonce + strconv.Itoa(int(curtime))
	encrypt := sha1.New()
	io.WriteString(encrypt, str)
	return fmt.Sprintf("%x", encrypt.Sum(nil))
}

// WyImClient.post 发送网易IM接受的post请求
func (self *WyImClient) post(url string, data url.Values, header map[string]string, rsptype interface{}) int {
	timestamp := uint32(time.Now().Unix())
	nonce := strconv.Itoa(int(timestamp))

	datastr := data.Encode()
	request, _ := http.NewRequest("POST", self.host+"/"+url, strings.NewReader(datastr))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	request.Header.Set("AppKey", self.appkey)
	request.Header.Set("Nonce", nonce)
	request.Header.Set("CurTime", strconv.Itoa(int(timestamp)))
	request.Header.Set("CheckSum", self.checksum(nonce, timestamp))
	//log.Println(request.Header)

	for k, v := range header {
		request.Header.Set(k, v)
	}

	rsp, err := http.DefaultClient.Do(request)
	log.Println(rsp.Status)
	if err != nil {
		log.Printf("请求失败:%v\n", err.Error())
		return -1
	}
	if rsp.Body != nil {
		defer rsp.Body.Close()
	}

	jdata, err := ioutil.ReadAll(rsp.Body)
	log.Println(string(jdata))
	if err != nil {
		log.Printf("包体读取失败:%v\n", err.Error())
	}
	if rsp != nil {
		json.Unmarshal(jdata, rsptype)
	}

	return rsp.StatusCode
}

// 初始化
func init() {
	im.Register("wyclient", new(WyImClient))
}
