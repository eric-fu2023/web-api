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
	baseUser := model.User{User: ploutos.User{BASE: ploutos.BASE{ID: 0234567}}}

	newNotif_Promo := ploutos.Notification{
		BASE_BY: ploutos.BASE_BY{
			CreatedBy: baseUser.ID,
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

	e1 := model.DB.Debug().Create(&newNotif_Promo).Error
	if e1 != nil {
		panic(e1)
	}

	log.Printf("notif.ID %d err %#v\n", newNotif_Promo.ID, e1)

	newPromoUserNotif := ploutos.UserNotification{
		UserId:         baseUser.ID,
		Text:           "?",
		NotificationId: newNotif_Promo.ID,
		IsRead:         false,
	}

	e2 := model.DB.Debug().Create(&newPromoUserNotif).Error
	if e2 != nil {
		panic(e1)
	}
	//log.Printf("notif %#v err %#v newnotif %#v err %#v\n", newNotif_Promo, e1, newPromoUserNotif, e2)

	var userNotifications []ploutos.UserNotification
	err := model.DB.Scopes(model.ByUserId(baseUser.ID), model.SortByCreated).Find(&userNotifications).Error
	if err != nil {
		panic(err)
	}

	log.Printf("userNotifications %#v\n", userNotifications)
	log.Printf("len(userNotifications) %d \n", len(userNotifications))

	log.Printf("model.DB.Delete(&newNotif_Promo).Error %v\n", model.DB.Delete(&newNotif_Promo).Error)
	log.Printf("model.DB.Delete(&newPromoUserNotif).Error %v\n", model.DB.Delete(&newPromoUserNotif).Error)

	var notifications2 []ploutos.UserNotification
	err = model.DB.Scopes(model.ByUserId(baseUser.ID), model.SortByCreated).Find(&notifications2).Error
	if err != nil {
		panic(err)
	}

	log.Printf("after del len(notifications2) %d \n", len(notifications2))
	{ // T_000006 General Notification, can create, mark as read
		newNotif_general := ploutos.Notification{
			BASE_BY: ploutos.BASE_BY{
				CreatedBy: baseUser.ID,
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

		e1 := model.DB.Debug().Create(&newNotif_general).Error
		if e1 != nil {
			panic(e1)
		}

		log.Printf("notif.ID %d err %#v\n", newNotif_general.ID, e1)

		newUserNotif_general := ploutos.UserNotification{
			UserId:         baseUser.ID,
			Text:           "general content?",
			NotificationId: newNotif_general.ID,
			IsRead:         false,
		}

		e2 := model.DB.Debug().Create(&newUserNotif_general).Error
		if e2 != nil {
			panic(e1)
		}

		getnotif, err := notificationservice.FindGeneralOne(context.TODO(), baseUser, ploutos.NotificationCategoryTypeNotification, newNotif_general.ID, newUserNotif_general.ID)
		if err != nil {
			panic(err)
		}
		log.Printf("getnotif %#v\n", getnotif)

		_, err = notificationservice.MarkNotificationAsRead(context.TODO(), baseUser, notificationservice.UserNotificationMarkReadForm{
			UserNotificationId: newUserNotif_general.ID,
			NotificationId:     newNotif_general.ID,
			CategoryType:       ploutos.NotificationCategoryTypeNotification,
		})
		if err != nil {
			panic(err)
		}
		getnotifaftermark, err := notificationservice.FindGeneralOne(context.TODO(), baseUser, ploutos.NotificationCategoryTypeNotification, newNotif_general.ID, newUserNotif_general.ID)
		if err != nil {
			panic(err)
		}
		log.Printf("getnotif.IsRead %#v\n", getnotifaftermark.IsRead)

	}

	{ // T_000006 given a user's targetless notification will not exist in `user_notification` table, then row should be created and read as mark.

		newNotif_Promo2 := ploutos.Notification{
			BASE_BY: ploutos.BASE_BY{
				CreatedBy: baseUser.ID,
			},
			Title:             "22222TESTTESTNotificationCategoryTypePromotion",
			Content:           "TESTTEST",
			Target:            0,
			Vip:               []int32{1, 2, 3, 4, 5},
			Category:          ploutos.NotificationCategoryTypePromotion,
			CategoryContentID: 0,
			PushEnable:        false,
			PushTitle:         "",
			PushContent:       "newNotif_Promo2 content",
			PushType:          0,
			PushTypeContentID: 0,
			SendAt:            time.Time{},
			ExpiredAt:         time.Time{},
			ImageUrl:          "",
			ShortContent:      "",
		}

		e1 := model.DB.Debug().Create(&newNotif_Promo2).Error
		if e1 != nil {
			panic(e1)
		}

		baseUser2 := model.User{User: ploutos.User{BASE: ploutos.BASE{ID: 02345672222}}}
		newUserNotif_Promo2Id, err := notificationservice.MarkNotificationAsRead(context.TODO(), baseUser2, notificationservice.UserNotificationMarkReadForm{
			NotificationId: newNotif_Promo2.ID,
			CategoryType:   ploutos.NotificationCategoryTypeNotification,
		})

		if err != nil {
			panic(err)
		}

		log.Printf("T_000006 %d \n", newUserNotif_Promo2Id)
	}
}
