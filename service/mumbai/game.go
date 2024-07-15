package mumbai

import (
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
	username := c.Merchant + c.Agent + user.IdAsString()

	res, err := client.LoginUser(username, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.

	// check whether there is error
	if err != nil {
		// check is it error with status code (EX002 - no account) , if yes then register new user.
		if err.Error() == string(ResponseCodeNotAccountFoundError) {
			// register new user.
			resp, err := client.RegisterUser(username, defaultPassword, extra.Ip)

			if err != nil {
				return "", err
			}

			// successfully register, and now login in the user again to get the url.
			res, err := client.LoginUser(resp.Result.UserName, defaultPassword, extra.Ip, tayaSubGameCode) // check for error code.

			if err != nil {
				return "", nil
			}

			return res.Result.GameCenterAddress, nil

		} else {
			return "", err
		}

	}

	// if no error meaning login successful , hence just return the url to front end.
	return res.Result.GameCenterAddress, nil

}
