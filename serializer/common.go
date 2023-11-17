package serializer

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"
	"web-api/conf/consts"
	"web-api/util/i18n"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

type TrackedErrorResponse struct {
	Response
	TrackID string `json:"track_id"`
}

const (
	CodeCheckLogin       = 401
	CodeNoRightErr       = 403
	CodeNotFound         = 404
	CodeGeneralError     = 50000
	CodeDBError          = 50001
	CodeEncryptError     = 50002
	CodeParamErr         = 40001
	CodeExistingUsername = 40002
	CodeSMSSent          = 40003
	CodeCaptchaInvalid   = 40004
	CodeOtpInvalid       = 40005
	CodeNoStream         = 100
)

func Err(c *gin.Context, service any, errCode int, msg string, err error) Response {
	res := Response{
		Code: errCode,
		Msg:  msg,
	}
	if err != nil && gin.Mode() != gin.ReleaseMode {
		res.Error = err.Error()
	}
	c.Set(consts.GinErrorKey, err)
	return res
}

func DBErr(c *gin.Context, service any, msg string, err error) Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	if msg == "" {
		msg = i18n.T("database_error")
	}
	return Err(c, service, CodeDBError, msg, err)
}

func ParamErr(c *gin.Context, service any, msg string, err error) Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	if msg == "" {
		msg = i18n.T("parameter_error")
	}
	return Err(c, service, CodeParamErr, msg, err)
}

func Url(original string) (new string) {
	if original == "" {
		return
	}
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

func EnsureErr(c *gin.Context, err error, res Response) Response {
	if res.Code != 0 {
		return res
	}
	return Err(c, "", CodeGeneralError, "", err)
}

func GeneralErr(c *gin.Context, err error) Response {
	i18n := c.MustGet("i18n").(i18n.I18n)
	return Err(c, "", CodeGeneralError, i18n.T("general_error"), err)
}

func UserSignature(userId int64) string {
	signatureHash := md5.Sum([]byte(fmt.Sprintf("%d%s", userId, os.Getenv("USER_SIGNATURE_SALT"))))
	return hex.EncodeToString(signatureHash[:])
}

func HouseClean(c *gin.Context, err error, res *Response) {
	if res.Code != 0 || res.Data != nil || res.Msg != "" {
		return
	}
	i18n := c.MustGet("i18n").(i18n.I18n)

	if err == nil {
		*res = Response{Msg: i18n.T("success")}
	}
	*res = Err(c, "", CodeGeneralError, "", err)
}
