package gin

import (
	"fmt"
	"saas/common/core/logger"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	"encoding/json"
	"io/ioutil"
)

func Middleware() gin.HandlerFunc {
	blogger := logger.GetLogger("bill")
	cfg := logger.GetConfig()
	return func(c *gin.Context) {

		module := strings.ToLower(c.Param("module"))
		handler := strings.ToLower(c.Param("handler"))
		var input interface{}
		var rsp interface{}
		timestamp := time.Now().UnixNano()
		// LOG
		log := logger.GetLogger(module)
		input = getInput(c)
		ignoreinputs := cfg.GetSlice(module, "ignoreinputs")
		if len(ignoreinputs) > 0 && inSlice(handler, ignoreinputs) {
			// 如果配置了该请求忽略，则不写入input
			log.WithFields(logrus.Fields{"input": "..."}).
				Infoln(fmt.Sprintf("%s/%s called", module, handler))
		} else {
			log.WithFields(logrus.Fields{"input": input}).
				Infoln(fmt.Sprintf("%s/%s called", module, handler))
		}
		// bill
		defer func() {
			go func() {
				bill(blogger, module, handler, input, rsp,
					(time.Now().UnixNano()-timestamp)/time.Millisecond.Nanoseconds(), c)
			}()
		}()

		// GO PASS
		c.Next()
		rsp = getOutput(c)
		ignoreoutputs := cfg.GetSlice(module, "ignoreoutputs")
		// LOG
		if len(ignoreoutputs) > 0 && inSlice(handler, ignoreoutputs) {
			// 如果配置了该请求忽略，则不写入output
			log.WithFields(logrus.Fields{"output": "..."}).
				Infoln(fmt.Sprintf("%s/%s returned", module, handler))
		} else {
			log.WithFields(logrus.Fields{"output": rsp}).
				Infoln(fmt.Sprintf("%s/%s returned", module, handler))
		}
	}
}

func bill(logger logger.Logger, module, action string, request, response interface{},
	cost int64, c *gin.Context) {
	logger.WithFields(logrus.Fields{
		"module": module,
		"action": action,
		"source": c.Request.RemoteAddr,
		"header": c.Request.Header,
		"rsp":    response,
		"req":    request,
		"cost":   fmt.Sprintf("%vms", cost),
	}).Info("")
}

func getInput(c *gin.Context) map[string]interface{} {
	input := map[string]interface{}{}
	if strings.Index(c.Request.Header.Get("Content-Type"), "multipart/form-data") >= 0 {
		// 如果是文件
		input["FileData"] = "......"
		return input
	}
	inputbody, _ := c.Request.GetBody()
	inputbyte, _ := ioutil.ReadAll(inputbody)
	err := json.Unmarshal(inputbyte, &input)
	if err != nil {
		// 如果非json
		input["NoJSON"] = string(inputbyte[:100]) + "..."
	}
	return input
}

func getOutput(c *gin.Context) interface{} {
	out, ok := c.Get("output")
	if ok && out != nil {
		return out
	}
	return nil
}
