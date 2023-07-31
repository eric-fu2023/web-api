package util

import (
	"web-api/util/i18n"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mailgun/mailgun-go/v4"
	"os"
	"time"
)

func SendEmail(c *gin.Context, email string, otp string) (err error) {
	i18n := c.MustGet("i18n").(i18n.I18n)
	mg := mailgun.NewMailgun(os.Getenv("MAILGUN_DOMAIN"), os.Getenv("MAILGUN_PRIVATE_KEY"))
	sender := os.Getenv("MAILGUN_SENDER")
	subject := i18n.T("Otp_email_subject")
	recipient := email
	message := mg.NewMessage(sender, subject, "", recipient)
	body := fmt.Sprintf(i18n.T("Otp_html_email"), otp)
	message.SetHtml(body)
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	_, _, err = mg.Send(ctx, message)
	return
}
