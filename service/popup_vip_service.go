package service

type VipService struct {
	Type    int64  `form:"type" json:"type" binding:"required"`
	Message string `form:"message" json:"message" binding:"required"`
}

// func (service *WinLoseService) Get(c *gin.Context) (r serializer.Response, err error) {
// 	i18n := c.MustGet("i18n").(i18n.I18n)
// 	var userRef string
// 	u, isUser := c.Get("user")
// 	if isUser {
// 		user := u.(model.User)
// 		userRef = fmt.Sprintf(`%d`, user.ID)
// 	} else {
// 		deviceInfo, e := util.GetDeviceInfo(c)
// 		if e != nil || deviceInfo.Uuid == "" {
// 			r = serializer.ParamErr(c, service, i18n.T("missing_device_uuid"), e)
// 			return
// 		}
// 		userRef = deviceInfo.Uuid
// 	}
// 	var messages []ploutos.PrivateMessage
// 	q := model.DB.Model(ploutos.PrivateMessage{}).Where(`user_ref = ? OR user_ref = '0'`, userRef).
// 		Order(`created_at DESC, id DESC`).Limit(service.PageById.Limit)
// 	if service.PageById.IdFrom != 0 {
// 		q = q.Where(`id < ?`, service.PageById.IdFrom)
// 	}
// 	err = q.Find(&messages).Error
// 	if err != nil {
// 		r = serializer.Err(c, service, serializer.CodeGeneralError, i18n.T("general_error"), err)
// 		return
// 	}
// 	var data []serializer.PrivateMessage
// 	for _, message := range messages {
// 		data = append(data, serializer.BuildPrivateMessage(message))
// 	}

// 	r = serializer.Response{
// 		Data: data,
// 	}
// 	return
// }
