package imsb_api

import (
	"blgit.rfdev.tech/taya/game-service/imsb/callback"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"web-api/service/imsb"
	"web-api/util"
)

func GetBalance(c *gin.Context) {
	var req callback.EncryptedRequest
	if err := c.ShouldBind(&req); err == nil {
		var decrypted callback.GetBalanceRequest
		e := DecryptRequest(req, &decrypted)
		if e != nil {
			c.String(200, ErrorResponse(c, req, e))
			return
		}
		res, e := imsb.GetBalanceCallback(c, decrypted, req)
		if e != nil {
			c.String(200, ErrorResponse(c, req, e))
			return
		}
		c.String(200, EncryptResponse(res))
	} else {
		c.String(200, ErrorResponse(c, req, err))
	}
}

func DeductBalance(c *gin.Context) {
	var req callback.EncryptedRequest
	if err := c.ShouldBind(&req); err == nil {
		var decrypted callback.WagerDetail
		e := DecryptRequest(req, &decrypted)
		if e != nil {
			c.String(200, ErrorResponse(c, req, e))
			return
		}
		res, e := imsb.OnBalanceDeduction(c, decrypted, req)
		if e != nil {
			c.String(200, ErrorResponse(c, req, e))
			return
		}
		c.String(200, EncryptResponse(res))
	} else {
		c.String(200, ErrorResponse(c, req, err))
	}
}

func UpdateBalance(c *gin.Context) {
	var req callback.EncryptedRequest
	if err := c.ShouldBind(&req); err == nil {
		var decrypted callback.UpdateBalanceRequest
		e := DecryptRequest(req, &decrypted)
		if e != nil {
			c.String(200, ErrorResponse(c, req, e))
			return
		}
		res, e := imsb.UpdateBalanceCallback(c, decrypted, req)
		if e != nil {
			c.String(200, ErrorResponse(c, req, e))
			return
		}
		c.String(200, EncryptResponse(res))
	} else {
		c.String(200, ErrorResponse(c, req, err))
	}
}

func DecryptRequest(req callback.EncryptedRequest, obj any) error {
	req.BalancePackage = strings.Replace(req.BalancePackage, " ", "+", -1)
	fmt.Println(req.BalancePackage)
	client := util.IMFactory.NewClient()
	decrypted, err := client.Decrypt(req.BalancePackage)
	if err != nil {
		return err
	}
	err = json.Unmarshal(decrypted, &obj)
	if err != nil {
		return err
	}
	return nil
}

func ErrorResponse(c *gin.Context, req any, err error) string {
	r := callback.BaseResponse{
		StatusCode: -100,
		StatusDesc: err.Error(),
	}
	util.Log().Error(r.StatusDesc, c.Request.URL, c.Request.Header, util.MarshalService(req))
	return EncryptResponse(r)
}

func EncryptResponse(data any) string {
	j, e := json.Marshal(data)
	if e != nil {
		util.Log().Error("imsb response error: ", e)
	}
	client := util.IMFactory.NewClient()
	encrypted, e := client.Encrypt(string(j))
	if e != nil {
		util.Log().Error("imsb response error: ", e)
	}
	return encrypted
}
