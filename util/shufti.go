package util

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"web-api/conf/consts"
	"web-api/model"
)

type Shufti struct {
	ClientId  string
	SecretKey string
}

func (s *Shufti) VerifyDocument(ctx context.Context, id int64, firstName, middleName, lastName, dob, nationality string, document, face []byte) (isAccepted bool, reason string, err error) {
	now := time.Now()
	reference := fmt.Sprintf(`%d-%d`, id, now.Unix())
	documentBase64 := base64Encode(document)
	faceBase64 := base64Encode(face)

	url := "https://api.shuftipro.com"
	method := "POST"

	payloadMap := map[string]interface{}{
		"reference":         reference,
		"country":           consts.CountryISO[nationality],
		"verification_mode": "any",
		"face": map[string]interface{}{
			"proof": faceBase64,
		},
		"document": map[string]interface{}{
			"proof":           documentBase64,
			"supported_types": []string{"id_card", "driving_license", "passport"},
			"name": map[string]interface{}{
				"first_name":  firstName,
				"middle_name": middleName,
				"last_name":   lastName,
			},
			"dob": dob,
		},
	}
	payloadByte, err := json.Marshal(payloadMap)
	if err != nil {
		return
	}
	payload := bytes.NewReader(payloadByte)

	basicAuth := toBase64([]byte(fmt.Sprintf(`%s:%s`, s.ClientId, s.SecretKey)))
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf(`Basic %s`, basicAuth))

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	var j map[string]interface{}
	err = json.Unmarshal(body, &j)
	if err != nil {
		return
	}

	event := ""
	if v, exists := j["event"]; exists {
		if vv, ok := v.(string); ok {
			event = vv
			if strings.Contains(vv, "accepted") {
				isAccepted = true
			} else {
				if rea, e := j["declined_reason"]; e {
					if rr, o := rea.(string); o {
						reason = rr
					}
				}
			}
		}
	}

	go logKycEvent(ctx, id, now, reference, event, string(body), res.StatusCode)
	return
}

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func base64Encode(image []byte) string {
	mime := http.DetectContentType(image)
	var base64Encoded string
	switch mime {
	case "image/jpeg":
		base64Encoded += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoded += "data:image/png;base64,"
	}
	base64Encoded += toBase64(image)
	return base64Encoded
}

func logKycEvent(ctx context.Context, id int64, now time.Time, reference, event, responseBody string, httpCode int) {
	kycEvent := model.KycEvent{
		KycEvent: ploutos.KycEvent{
			KycId:        id,
			DateTime:     now.Format(time.DateTime),
			Reference:    reference,
			Event:        event,
			HttpCode:     httpCode,
			ResponseBody: responseBody,
		},
	}
	err := model.LogKycEvent(kycEvent)
	if err != nil {
		GetLoggerEntry(ctx).Errorf("LogKycEvent error: %s", err.Error())
	}
}
