package mumbai

import (
	"blgit.rfdev.tech/taya/game-service/mumbai"
	"errors"
	"fmt"
	"web-api/model"
	"web-api/util"
)

const defaultPassword = "qq123456"

func (c *Mumbai) GetGameUrl(user model.User, currency, tayaGameCode, tayaSubGameCode string, _ int64, extra model.Extra) (string, error) {
	// creates the client so that we can call the login method.
	client, err := util.MumbaiFactory()

	if err != nil {
		return "", err
	}

	// try to login user and if there's is an error (EX002) meaning this user has not been created yet so we call register
	// and then login the user again to get the url.
	username := c.Merchant + c.Agent + fmt.Sprintf("%08s", user.IdAsString())
	res, err := client.LoginUser(username, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.

	// check whether there is error
	if err != nil {
		// check is it error with status code (EX002 - no account) , if yes then register new user.
		if errors.Is(err, mumbai.ErrAccountNotFound) {
			// register new user.
			_, regErr := client.RegisterUser(username, defaultPassword, extra.Ip)
			if regErr != nil {
				return "", regErr
			}
			// successfully register, and now login in the user again to get the url.
			loginResp, loginErr := client.LoginUser(username, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.
			if loginErr != nil {
				return "", nil
			}
			return loginResp.Result.GameCenterAddress, nil
		}
		return "", err
	}

	// if no error meaning login successful , hence just return the url to front end.
	return res.Result.GameCenterAddress, nil

}
