package main

import (
	"fmt"

	"github.com/iamtraining/gallery/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1111"
	dbname   = "gallery_dev"
)

func main() {
	psql := fmt.Sprintf("host=%s port =%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	us, err := models.NewUserService(psql)
	if err != nil {
		panic(err)
	}
	defer us.Close()
	us.AutoMigrate()

	user := models.User{
		Name:     "Michael Scott",
		Email:    "michael@dundermifflin.com",
		Password: "bestboss",
	}
	err = us.Create(&user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", user)
	if user.Remember == "" {
		panic("invalid remember token")
	}

	user2, err := us.ByRemember(user.Remember)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", user2)
}
