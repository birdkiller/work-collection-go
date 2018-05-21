package logger

import (
	"encoding/json"
	"fmt"
	"log"
)

type Config map[string]interface{}

// 取配置 keys - 配置路径
func (self *Config) Get(keys ...string) interface{} {
	var o interface{}
	o = map[string]interface{}(*self)
	for _, k := range keys {
		ot, ok := o.(map[string]interface{})
		if ok != true {
			return nil
		}
		o, _ = ot[k]
		if o == nil {
			return nil
		}
	}
	return o
}

// *Config.GetValue获得配置值 keys - 配置路径
func (self *Config) GetValue(keys ...string) string {
	v := self.Get(keys...)
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// *Config.GetSlice获取数组值 keys - 配置路径
func (self *Config) GetSlice(keys ...string) (out []string) {
	out = []string{}
	s := self.Get(keys...)
	if s == nil {
		return out
	}
	sl, ok := s.([]interface{})
	if ok {
		for _, v := range sl {
			out = append(out, fmt.Sprintf("%v", v))
		}
		return out
	} else {
		return out
	}
}

// 获取子配置对象
func (self *Config) GetConfig(keys ...string) *Config {
	c := self.Get(keys...)
	if c == nil {
		return &Config{}
	}
	config, ok := c.(map[string]interface{})
	if ok {
		subc := Config(config)
		return &subc
	} else {
		return &Config{}
	}

}

func (self *Config) Keys() []string {
	out := []string{}
	for k, _ := range *self {
		out = append(out, k)
	}
	return out
}

// Load 加载配置
func Load(c map[string]interface{}) *Config {
	config := Config(c)
	return &config
}

// LoadFromJson
func LoadJson(c []byte) *Config {
	config := Config{}
	err := json.Unmarshal(c, &config)
	if err != nil {
		log.Println("配置加载错误：", err)
	}
	return &config
}
