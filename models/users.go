package models

import (
	"errors"

	"github.com/iamtraining/gallery/hash"
	"github.com/iamtraining/gallery/rand"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided was invalid")
	ErrInvalodPassword = errors.New("models: incorrect password provided")
)

var userPwPepper = "secret-random-string"

const hmacSecretKey = "secret-hmac-key"

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null; unique_index"`
}

type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

func NewUserService(conninfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", conninfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	hmac := hash.NewHMAC(hmacSecretKey)

	return &UserService{
		db:   db,
		hmac: hmac,
	}, nil
}

func (us *UserService) Close() error {
	return us.db.Close()
}

func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id).First(&user)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (us *UserService) DestructiveReset() error {
	err := us.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return err
	}

	return us.AutoMigrate()
}

func (us *UserService) Create(u *User) error {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password+userPwPepper), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedBytes)
	u.Password = ""

	if u.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		u.Remember = token
	}

	u.RememberHash = us.hmac.Hash(u.Remember)

	return us.db.Create(u).Error
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)

	return &user, err
}

func (us *UserService) Update(u *User) error {
	if u.Remember != "" {
		u.RememberHash = us.hmac.Hash(u.Remember)
	}

	return us.db.Save(u).Error
}

func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (us *UserService) Authentificate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(password+userPwPepper),
	)

	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalodPassword
	default:
		return nil, err
	}

	return nil, nil
}

func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	rememberHash := us.hmac.Hash(token)
	err := first(us.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, err
}
