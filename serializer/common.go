package serializer

import (
	"github.com/gin-gonic/gin"
	"math"
	"os"
	"strings"
	"web-api/util"
)

// Response 基础序列化器
type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

// TrackedErrorResponse 有追踪信息的错误响应
type TrackedErrorResponse struct {
	Response
	TrackID string `json:"track_id"`
}

// 三位数错误编码为复用http原本含义
// 五位数错误编码为应用自定义错误
// 五开头的五位数错误编码为服务器端错误，比如数据库操作失败
// 四开头的五位数错误编码为客户端错误，有时候是客户端代码写错了，有时候是用户操作错误
const (
	// CodeCheckLogin 未登录
	CodeCheckLogin = 401
	// CodeNoRightErr 未授权访问
	CodeNoRightErr = 403
	CodeNotFound   = 404
	// General errors
	CodeGeneralError = 50000
	// CodeDBError 数据库操作失败
	CodeDBError = 50001
	// CodeEncryptError 加密失败
	CodeEncryptError = 50002
	//CodeParamErr 各种奇奇怪怪的参数错误
	CodeParamErr         = 40001
	CodeExistingUsername = 40002
)

// Err 通用错误处理
func Err(c *gin.Context, service any, errCode int, msg string, err error) Response {
	res := Response{
		Code: errCode,
		Msg:  msg,
	}
	// 生产环境隐藏底层报错
	if err != nil && gin.Mode() != gin.ReleaseMode {
		res.Error = err.Error()
	}
	fn := util.Log().Error
	if errCode == CodeParamErr {
		fn = util.Log().Info
	}
	fn(msg, err, c.Request.URL, c.Request.Header, util.MarshalService(service))
	return res
}

// DBErr 数据库操作失败
func DBErr(c *gin.Context, service any, msg string, err error) Response {
	if msg == "" {
		msg = "数据库操作失败"
	}
	return Err(c, service, CodeDBError, msg, err)
}

// ParamErr 各种参数错误
func ParamErr(c *gin.Context, service any, msg string, err error) Response {
	if msg == "" {
		msg = "参数错误"
	}
	return Err(c, service, CodeParamErr, msg, err)
}

func Url(original string) (new string) {
	if strings.HasPrefix(original, "http") {
		new = original
	} else {
		if strings.HasPrefix(original, "/") {
			new = os.Getenv("IMG_BASE_URL") + original
		} else {
			new = os.Getenv("IMG_BASE_URL") + "/" + original
		}
	}
	return
}

func FormatMarketValueCurrency(c *gin.Context, currency string) (new string) {
	if currency == "€" { // only for euro
		if c.MustGet("_language").(string) == "zh" {
			new = "欧"
		} else {
			new = "€"
		}
	}
	return
}

func AvgValuePerMatch(value int64, matches int64) (avg float64) {
	avg = math.Round(float64(value)/float64(matches)*10) / 10
	return
}

func RoundValue(numerator int64, denominator int64, dCoeff int64, dividedBy int64) (avg float64) {
	if denominator > 0 {
		avg = math.Round(float64(numerator)/float64(denominator)*float64(dCoeff)) / float64(dividedBy)
	}
	return
}
