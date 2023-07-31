package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"web-api/util/i18n"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func SendSms(c *gin.Context, countryCode string, mobile string, otp string) (err error) {
	if countryCode == "+86" { // for china
		//msg := `【欧场】尊敬的用户，您的验证码：` + otp + `，如非本人操作，请忽略本短信`
		//url := `http://www.weiwebs.cn/msg/HttpSendSM?account=` + os.Getenv("CN_SMS_ACCOUNT") + `&pswd=` + os.Getenv("CN_SMS_PSWD") + `&mobile=` + mobile + `&msg=` + msg
		//var res *http.Response
		//var r []byte
		//res, err = http.Get(url)
		//r, err = io.ReadAll(res.Body)
		//re := strings.Split(string(r), ",")
		//if re[1] != "0" {
		//	err = errors.New(re[1])
		//}
		//res.Body.Close()

		data := url.Values{}
		data.Set("SpCode", os.Getenv("CN_SMS_SPCODE"))
		data.Set("LoginName", os.Getenv("CN_SMS_LOGINNAME"))
		data.Set("Password", os.Getenv("CN_SMS_PASSWORD"))
		data.Set("MessageContent", fmt.Sprintf(`您的验证码是 %s，5分钟有效，请尽快验证`, otp))
		data.Set("UserNumber", mobile)
		encoded := data.Encode()

		var req *http.Request
		req, err = http.NewRequest("POST", "https://api.huanxun58.com/sms/Api/ReturnJson/Send.do", strings.NewReader(encoded))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		if _, err = client.Do(req); err != nil {
			fmt.Println(err)
		}
	} else { // for the rest
		if os.Getenv("USE_BULKSMS") == "" || os.Getenv("USE_BULKSMS") != "true" {
			credential := common.NewCredential(
				os.Getenv("TENCENTCLOUD_SECRET_ID"),
				os.Getenv("TENCENTCLOUD_SECRET_KEY"),
			)

			cpf := profile.NewClientProfile()
			cpf.HttpProfile.ReqMethod = "POST"
			cpf.HttpProfile.ReqTimeout = 15
			cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
			cpf.SignMethod = "HmacSHA1"
			client, _ := sms.NewClient(credential, "ap-singapore", cpf)
			request := sms.NewSendSmsRequest()
			request.SmsSdkAppId = common.StringPtr(os.Getenv("TENCENTCLOUD_APP_ID"))
			request.TemplateId = common.StringPtr(os.Getenv("TENCENTCLOUD_TEMPLATE_ID"))
			request.TemplateParamSet = common.StringPtrs([]string{otp})
			request.PhoneNumberSet = common.StringPtrs([]string{countryCode + mobile})
			_, err = client.SendSms(request)
		} else {
			i18n := c.MustGet("i18n").(i18n.I18n)
			m := map[string]interface{}{
				"from": os.Getenv("BULKSMS_FROM"),
				"to": countryCode + mobile,
				"body": fmt.Sprintf(i18n.T("Your_request_otp"), otp),
				"encoding": "UNICODE",
			}
			var str []byte
			str, err = json.Marshal(m)
			if err != nil {
				return
			}

			var req *http.Request
			req, err = http.NewRequest("POST", "https://api.bulksms.com/v1/messages", bytes.NewReader(str))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Basic " + os.Getenv("BULKSMS_AUTH"))
			client := &http.Client{}
			_, err = client.Do(req)
			if err != nil {
				return
			}
		}
	}
	return
}
