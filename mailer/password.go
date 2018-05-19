package mailer

import (
	"crypto/tls"
	"fmt"

	"github.com/liverecord/lrs"
	"github.com/matcornic/hermes"
	"gopkg.in/gomail.v2"
)

type reset struct {
}

func (r *reset) Name() string {
	return "reset"
}

func (r *reset) Email(user lrs.User, password string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: user.Name,
			Intros: []string{
				"We have generated a new password for you.",
			},
			Dictionary: []hermes.Entry{
				{
					Key:   "New Password",
					Value: password,
				},
			},
			Outros: []string{
				"You can always change your password later in the settings.",
			},
			Signature: "Thanks",
		},
	}
}

// SendPasswordReset sends new password
func SendPasswordReset(Cfg *lrs.Config, user lrs.User, password string) {

	h := hermes.Hermes{
		Product: hermes.Product{
			Name:        Cfg.Name,
			Link:        Cfg.SiteURL(),
			Logo:        "https://avatars3.githubusercontent.com/u/31494987?s=100&v=4",
			Copyright:   "LiveRecord",
			TroubleText: "",
		},
	}
	e := new(reset)
	// h.Theme = new(hermes.Flat)
	res, err := h.GenerateHTML(e.Email(user, password))
	fmt.Println(err)
	fmt.Println(res)

	m := gomail.NewMessage()
	m.SetHeader("From", Cfg.SMTP.From)
	m.SetHeader("To", user.Email)

	m.SetHeader("Subject", "New Password")

	m.SetBody("text/html", res)

	d := gomail.NewDialer(Cfg.SMTP.Host, Cfg.SMTP.Port, Cfg.SMTP.Username, Cfg.SMTP.Password)
	if Cfg.SMTP.InsecureTLS {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	d.SSL = Cfg.SMTP.SSL

	if err := d.DialAndSend(m); err != nil {
		panic(err)

	}
}
