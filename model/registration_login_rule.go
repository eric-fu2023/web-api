package model

import (
	ploutos "blgit.rfdev.tech/taya/ploutos-object"
	"gorm.io/gorm"
)

type RegistrationCount struct {
	DeviceCount int `json:"device_count"`
	IPCount     int `json:"ip_count"`
}

func GetRegistrationLoginRule() (rule ploutos.RegistrationLoginRules, err error) {

	err = DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Table("registration_login_rules").First(&rule).Error
		return
	})

	return
}

func GetRegistrationDeviceIPCount(deviceId, ip string) (deviceCount, ipCount int) {

	var registrationCount RegistrationCount

	query := `
        SELECT 
            COUNT(CASE WHEN registration_device_uuid = ? THEN 1 END) AS device_count,
            COUNT(CASE WHEN registration_ip = ? THEN 1 END) AS ip_count
        FROM users
    `

	err := DB.Raw(query, deviceId, ip).Scan(&registrationCount).Error
	if err != nil {
		return
	}

	deviceCount = registrationCount.DeviceCount
	ipCount = registrationCount.IPCount

	return
}
