package wyclient

import (
	"regexp"
)

// uid禁用字符匹配对象
var reg_uid_forbiddent *regexp.Regexp

func init() {
	reg_uid_forbiddent = regexp.MustCompile(`[^A-Za-z0-9_\.\-]`)
}

// uid将全部非法字符转化为下划线
func uidfilter(uid string) (accid string) {
	accid = reg_uid_forbiddent.ReplaceAllString(uid, "_")
	return
}
