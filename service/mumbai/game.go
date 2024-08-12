package mumbai

import (
	"errors"
	"fmt"
	"log"
	"os"

	"web-api/model"
	"web-api/util"

	"blgit.rfdev.tech/taya/game-service/mumbai"
	"blgit.rfdev.tech/taya/game-service/mumbai/api"
)

const defaultPassword = "qq123456"

// login, if not found => register user, then login again.
func (c *Mumbai) LoginWithCreateUser(mbUserName string, password string, clientIP string, gameCode string) (api.LoginResponseBody, error) {
	client, err := util.MumbaiFactory()
	if err != nil {
		return api.LoginResponseBody{}, err
	}

	res, err := client.LoginUser(mbUserName, password, clientIP, gameCode)
	switch {
	case err == nil:
		return res, nil
	case errors.Is(err, mumbai.ErrAccountNotFound): // try again once
		_, regErr := client.RegisterUser(mbUserName, password, clientIP)
		if regErr != nil {
			log.Printf("Mumbai GetOrCreateUser mumbai username %s not found. register fail err: %v \n", mbUserName, regErr)
			return api.LoginResponseBody{}, regErr
		}
		res, err = client.LoginUser(mbUserName, password, clientIP, gameCode)
		if err != nil {
			return api.LoginResponseBody{}, err
		}
		return res, nil
	default:
		return api.LoginResponseBody{}, err
	}
}

func (c *Mumbai) GetGameUrl(user model.User, currency, tayaGameCode, tayaSubGameCode string, _ int64, extra model.Extra) (string, error) {
	// creates the client so that we can call the login method.
	client, err := util.MumbaiFactory()
	if err != nil {
		return "", err
	}
	log.Printf("Mumbai GetGameUrl ... GAME_MUMBAI_MODE_NAME %s client %#v \n", os.Getenv("GAME_MUMBAI_MODE_NAME"), client)

	// try to login user and if there's is an error (EX002) meaning this user has not been created yet so we call register
	// and then login the user again to get the url.
	username := os.Getenv("GAME_MUMBAI_MERCHANT_CODE") + os.Getenv("GAME_MUMBAI_AGENT_CODE") + fmt.Sprintf("%08s", user.IdAsString())
	log.Printf("Mumbai GetGameUrl mumbai username %s \n", username)

	res, err := c.LoginWithCreateUser(username, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.
	if err != nil {
		return "", err
	}
	return res.Result.GameCenterAddress, nil
}
