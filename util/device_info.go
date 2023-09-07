package util

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
)

var (
	ErrDeviceInfoEmpty = errors.New("device info is empty")
)

type DeviceInfo struct {
	Platform string `json:"platform"`
}

func GetDeviceInfo(c *gin.Context) (DeviceInfo, error) {
	d := DeviceInfo{}
	if c.GetHeader("Device-Info") == "" {
		return DeviceInfo{}, ErrDeviceInfoEmpty
	}
	if err := json.Unmarshal([]byte(c.GetHeader("Device-Info")), &d); err != nil {
		return DeviceInfo{}, err
	}
	return d, nil
}
