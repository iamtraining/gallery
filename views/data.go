package views

import (
	"log"

	"github.com/iamtraining/gallery/models"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"
	AlertMsg        = "Something went wrong. Please try again."
)

type Data struct {
	Alert *Alert
	Body  interface{}
	User  *models.User
}

type Alert struct {
	Level   string
	Message string
}

func (d *Data) SetAlert(err error) {
	var msg string
	if pub, ok := err.(PublicError); ok {
		msg = pub.Public()
	} else {
		log.Println(err)
		msg = AlertMsg
	}

	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}

func (d *Data) CreateErrorAlert(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}
