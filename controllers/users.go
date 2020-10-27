package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/iamtraining/gallery/models"
	"github.com/iamtraining/gallery/rand"
	"github.com/iamtraining/gallery/views"
)

type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        models.UserService
}

type RegisterForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView: views.NewView(
			"bootstrap",
			"users/new",
		),
		LoginView: views.NewView(
			"bootstrap",
			"users/login",
		),
		us: us,
	}
}

// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, r, nil)

}

// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var data views.Data

	var form RegisterForm

	if err := parseForm(r, &form); err != nil {
		log.Println(err)
		data.SetAlert(err)
		u.NewView.Render(w, r, data)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil {
		data.SetAlert(err)
		u.NewView.Render(w, r, data)
		return
	}

	err := u.signIn(w, &user)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var data views.Data

	var form LoginForm

	if err := parseForm(r, &form); err != nil {
		data.SetAlert(err)
		u.LoginView.Render(w, r, data)
		return
	}

	user, err := u.us.Authentificate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			data.CreateErrorAlert("invalid email address")
		default:
			data.SetAlert(err)
		}
		u.LoginView.Render(w, r, data)
		return
	}

	err = u.signIn(w, user)
	if err != nil {
		data.SetAlert(err)
		u.LoginView.Render(w, r, data)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}

		user.Remember = token
		err = u.us.Update(user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	return nil
}

func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusFound)
		return
	}

	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusFound)
		return
	}

	fmt.Fprintln(w, user)
}
