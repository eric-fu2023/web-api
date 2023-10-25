package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Shufti struct {
	ClientId  string
	SecretKey string
}

func (s *Shufti) VerifyDocument(id int64, firstName, middleName, lastName, dob string, document, face []byte) (isAccepted bool, reason string, err error) {
	reference := fmt.Sprintf(`%d-%d`, id, time.Now().Unix())
	documentBase64 := base64Encode(document)
	faceBase64 := base64Encode(face)

	url := "https://api.shuftipro.com"
	method := "POST"

	payloadMap := map[string]interface{}{
		"reference":         reference,
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
	if v, exists := j["event"]; exists {
		if vv, ok := v.(string); ok {
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
