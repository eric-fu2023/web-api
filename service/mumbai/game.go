package mumbai

import (
	"blgit.rfdev.tech/taya/game-service/mumbai"
	"errors"
	"fmt"
	"log"
	"os"
	"web-api/model"
	"web-api/util"
)

const defaultPassword = "qq123456"

func (c *Mumbai) GetGameUrl(user model.User, currency, tayaGameCode, tayaSubGameCode string, _ int64, extra model.Extra) (string, error) {
	log.Printf("Mumbai GetGameUrl ... \n")
	// creates the client so that we can call the login method.
	client, err := util.MumbaiFactory()
	if err != nil {
		return "", err
	}

	// try to login user and if there's is an error (EX002) meaning this user has not been created yet so we call register
	// and then login the user again to get the url.
	username := os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + os.Getenv("GAME_MUMBAI_AGENT_CODE") + fmt.Sprintf("%08s", user.IdAsString())
	log.Printf("Mumbai GetGameUrl mumbai username %s \n", username)

	res, err := client.LoginUser(username, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.
	if err != nil {
		// check is it error with status code (EX002 - no account) , if yes then register new user.
		if errors.Is(err, mumbai.ErrAccountNotFound) {
			log.Printf("Mumbai GetGameUrl mumbai username %s not found. registering... \n", username)
			// register new user.
			_, regErr := client.RegisterUser(username, defaultPassword, extra.Ip)
			if regErr != nil {
				log.Printf("Mumbai GetGameUrl mumbai username %s not found. register fail err: %v \n", username, regErr)
				return "", regErr
			}
			// successfully register, and now login in the user again to get the url.
			loginResp, loginErr := client.LoginUser(username, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.
			if loginErr != nil {
				log.Printf("Mumbai GetGameUrl mumbai username %s not found. register success and login fail err: %v\n", username, loginErr)
				return "", loginErr
			}
			return loginResp.Result.GameCenterAddress, nil
		}
		return "", err
	}

	return res.Result.GameCenterAddress, nil
}
