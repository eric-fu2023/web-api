package main

import (
	"context"
	"log"
	"time"

	"web-api/conf"
	"web-api/model"
	notificationservice "web-api/service/notification"

	ploutos "blgit.rfdev.tech/taya/ploutos-object"
)

func main() {
	NotificationModule()
}

// basic create, del and mark as read
func NotificationModule() {
	conf.Init()
	log.Println("config initialized")
	var userId int64 = 0234567

	newPromoNotif := ploutos.Notification{
		BASE_BY: ploutos.BASE_BY{
			CreatedBy: userId,
			UpdatedBy: 0,
			DeletedBy: 0,
		},
		Title:             "TESTTESTNotificationCategoryTypePromotion",
		Content:           "TESTTEST",
		Target:            0,
		Vip:               []int32{1, 2, 3, 4, 5},
		Category:          ploutos.NotificationCategoryTypePromotion,
		CategoryContentID: 0,
		PushEnable:        false,
		PushTitle:         "",
		PushContent:       "TESTTESTPUSHPUSH",
		PushType:          0,
		PushTypeContentID: 0,
		SendAt:            time.Time{},
		ExpiredAt:         time.Time{},
		ImageUrl:          "",
		ShortContent:      "",
	}

	e1 := model.DB.Debug().Create(&newPromoNotif).Error
	if e1 != nil {
		panic(e1)
	}

	log.Printf("notif.ID %d err %#v\n", newPromoNotif.ID, e1)

	newPromoUserNotif := ploutos.UserNotification{
		UserId:         userId,
		Text:           "?",
		NotificationId: newPromoNotif.ID,
		IsRead:         false,
	}

	e2 := model.DB.Debug().Create(&newPromoUserNotif).Error
	if e2 != nil {
		panic(e1)
	}
	//log.Printf("notif %#v err %#v newnotif %#v err %#v\n", newPromoNotif, e1, newPromoUserNotif, e2)

	var userNotifications []ploutos.UserNotification
	err := model.DB.Scopes(model.ByUserId(userId), model.SortByCreated).Find(&userNotifications).Error
	if err != nil {
		panic(err)
	}

	log.Printf("userNotifications %#v\n", userNotifications)
	log.Printf("len(userNotifications) %d \n", len(userNotifications))

	log.Printf("model.DB.Delete(&newPromoNotif).Error %v\n", model.DB.Delete(&newPromoNotif).Error)
	log.Printf("model.DB.Delete(&newPromoUserNotif).Error %v\n", model.DB.Delete(&newPromoUserNotif).Error)

	var notifications2 []ploutos.UserNotification
	err = model.DB.Scopes(model.ByUserId(userId), model.SortByCreated).Find(&notifications2).Error
	if err != nil {
		panic(err)
	}

	log.Printf("after del len(notifications2) %d \n", len(notifications2))
	{ // General
		newGeneralNotif := ploutos.Notification{
			BASE_BY: ploutos.BASE_BY{
				CreatedBy: userId,
				UpdatedBy: 0,
				DeletedBy: 0,
			},
			Title:             "TESTTESTNotificationCategoryTypeNotification",
			Content:           "TESTTEST",
			Target:            0,
			Vip:               []int32{1, 2, 3, 4, 5},
			Category:          ploutos.NotificationCategoryTypeNotification,
			CategoryContentID: 0,
			PushEnable:        false,
			PushTitle:         "",
			PushContent:       "TESTTESTPUSHPUSH",
			PushType:          0,
			PushTypeContentID: 0,
			SendAt:            time.Time{},
			ExpiredAt:         time.Time{},
			ImageUrl:          "",
			ShortContent:      "",
		}

		e1 := model.DB.Debug().Create(&newGeneralNotif).Error
		if e1 != nil {
			panic(e1)
		}

		log.Printf("notif.ID %d err %#v\n", newGeneralNotif.ID, e1)

		newUserNotif_general := ploutos.UserNotification{
			UserId:         userId,
			Text:           "general content?",
			NotificationId: newGeneralNotif.ID,
			IsRead:         false,
		}

		e2 := model.DB.Debug().Create(&newUserNotif_general).Error
		if e2 != nil {
			panic(e1)
		}

		baseUser := model.User{User: ploutos.User{BASE: ploutos.BASE{ID: userId}}}
		getnotif, err := notificationservice.FindGeneralOne(context.TODO(), baseUser, ploutos.NotificationCategoryTypeNotification, newGeneralNotif.ID, newUserNotif_general.ID)

		if err != nil {
			panic(err)
		}
		log.Printf("getnotif %#v\n", getnotif)

		_, err = notificationservice.MarkNotificationAsRead(context.TODO(), baseUser, notificationservice.UserNotificationMarkReadForm{
			UserNotificationId: newUserNotif_general.ID,
			NotificationId:     newGeneralNotif.ID,
			CategoryType:       ploutos.NotificationCategoryTypeNotification,
		})
		if err != nil {
			panic(err)
		}
		getnotifaftermark, err := notificationservice.FindGeneralOne(context.TODO(), baseUser, ploutos.NotificationCategoryTypeNotification, newGeneralNotif.ID, newUserNotif_general.ID)
		if err != nil {
			panic(err)
		}
		log.Printf("getnotif.IsRead %#v\n", getnotifaftermark.IsRead)

	}
}
