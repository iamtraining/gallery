package models

import (
	"errors"
	"regexp"
	"strings"

	"github.com/iamtraining/gallery/hash"
	"github.com/iamtraining/gallery/rand"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound              = errors.New("models: resource not found")
	ErrIDInvalid             = errors.New("models: ID provided was invalid")
	ErrPasswordRequired      = errors.New("models: password is required")
	ErrPasswordShort         = errors.New("models: password must be at least 8 characters long")
	ErrPasswordInvalid       = errors.New("models: incorrect password provided")
	ErrEmailRequired         = errors.New("models: email address is required")
	ErrEmailInvalid          = errors.New("models: email address is not valid")
	ErrEmailAlreadyTaken     = errors.New("models: email address is already taken")
	ErrRememberTokenRequired = errors.New("models: remember token is required")
	ErrRememberTokenTooShort = errors.New("models: remember token must be at least 32 bytes")
)

var userPwPepper = "secret-random-string"

var _ UserDB = &userGorm{}
var _ UserService = &userService{}

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

type userService struct {
	UserDB
}

type UserService interface {
	Authentificate(email, password string) (*User, error)
	UserDB
}

type userGorm struct {
	db *gorm.DB
}

type userValidator struct {
	UserDB
	hmac        hash.HMAC
	emailRegexp *regexp.Regexp
}

type userValFunc func(*User) error

type UserDB interface {
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// close a db connection
	Close() error

	//migration helpers
	AutoMigrate() error
	DestructiveReset() error

	// altering users methods
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
}

func NewUserService(conninfo string) (UserService, error) {
	ug, err := newUserGorm(conninfo)
	if err != nil {
		return nil, err
	}

	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)

	return &userService{
		UserDB: uv,
	}, nil
}

func newUserGorm(conninfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", conninfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	return &userGorm{
		db: db,
	}, nil
}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB: udb,
		hmac:   hmac,
		emailRegexp: regexp.MustCompile(
			`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`,
		),
	}
}

func (ug *userGorm) Close() error {
	return ug.db.Close()
}

func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ug *userGorm) DestructiveReset() error {
	err := ug.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return err
	}

	return ug.AutoMigrate()
}

func (ug *userGorm) Create(u *User) error {
	return ug.db.Create(u).Error
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)

	return &user, err
}

func (ug *userGorm) Update(u *User) error {
	return ug.db.Save(u).Error
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (us *userService) Authentificate(email, password string) (*User, error) {
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
		return nil, ErrPasswordInvalid
	default:
		return nil, err
	}
}

func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user)
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) Create(user *User) error {
	if err := runUserValFuncs(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.emailNorms,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvailable,
	); err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	if err := runUserValFuncs(user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.emailNorms,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvailable,
	); err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValFuncs(&user,
		uv.idCheck(id),
	)
	if err != nil {
		return err
	}

	return uv.UserDB.Delete(id)
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	return nil
}

func runUserValFuncs(user *User, funcs ...userValFunc) error {
	for _, fn := range funcs {
		if err := fn(user); err != nil {
			return err
		}
	}

	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)

	return nil
}

func (uv *userValidator) setRememberUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token

	return nil
}

func (uv *userValidator) idCheck(n uint) userValFunc {
	return userValFunc(func(u *User) error {
		if u.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

func (uv *userValidator) emailNorms(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)

	return nil
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}

	err := runUserValFuncs(&user,
		uv.emailNorms)
	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}

	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}

	if uv.emailRegexp.MatchString(user.Email) {
		return ErrEmailInvalid
	}

	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	ok, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	if user.ID != ok.ID {
		return ErrEmailAlreadyTaken
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}

	if len(user.Password) < 8 {
		return ErrPasswordShort
	}

	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}

	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}

	if n < 32 {
		return ErrRememberTokenTooShort
	}

	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberTokenRequired
	}

	return nil
}
