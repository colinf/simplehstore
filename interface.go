package db

import "net/http"

// Database interfaces

type IList interface {
	Add(value string) error
	GetAll() ([]string, error)
	GetLast() (string, error)
	GetLastN(n int) ([]string, error)
	Remove() error
	Clear() error
}

type ISet interface {
	Add(value string) error
	Has(value string) (bool, error)
	GetAll() ([]string, error)
	Del(value string) error
	Remove() error
	Clear() error
}

type IHashMap interface {
	Set(owner, key, value string) error
	Get(owner, key string) (string, error)
	Has(owner, key string) (bool, error)
	Exists(owner string) (bool, error)
	GetAll() ([]string, error)
	DelKey(owner, key string) error
	Del(key string) error
	Remove() error
	Clear() error
}

type IKeyValue interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Del(key string) error
	Remove() error
	Clear() error
}

// Interface for making it possible to depend on different versions of the permission package, or other packages that implement userstates.
type IUserState interface {
	UserRights(req *http.Request) bool
	HasUser(username string) bool
	BooleanField(username, fieldname string) bool
	SetBooleanField(username, fieldname string, val bool)
	IsConfirmed(username string) bool
	IsLoggedIn(username string) bool
	AdminRights(req *http.Request) bool
	IsAdmin(username string) bool
	UsernameCookie(req *http.Request) (string, error)
	SetUsernameCookie(w http.ResponseWriter, username string) error
	AllUsernames() ([]string, error)
	Email(username string) (string, error)
	PasswordHash(username string) (string, error)
	AllUnconfirmedUsernames() ([]string, error)
	ConfirmationCode(username string) (string, error)
	AddUnconfirmed(username, confirmationCode string)
	RemoveUnconfirmed(username string)
	MarkConfirmed(username string)
	RemoveUser(username string)
	SetAdminStatus(username string)
	RemoveAdminStatus(username string)
	addUserUnchecked(username, passwordHash, email string)
	AddUser(username, password, email string)
	SetLoggedIn(username string)
	SetLoggedOut(username string)
	Login(w http.ResponseWriter, username string)
	Logout(username string)
	Username(req *http.Request) string
	CookieTimeout(username string) int64
	SetCookieTimeout(cookieTime int64)
	PasswordAlgo() string
	SetPasswordAlgo(algorithm string) error
	HashPassword(username, password string) string
	CorrectPassword(username, password string) bool
	AlreadyHasConfirmationCode(confirmationCode string) bool
	FindUserByConfirmationCode(confirmationcode string) (string, error)
	Confirm(username string)
	ConfirmUserByConfirmationCode(confirmationcode string) error
	SetMinimumConfirmationCodeLength(length int)
	GenerateUniqueConfirmationCode() (string, error)

	// Related to the database backend
	Users() *IHashMap
	Host() *IHost
}

// A database host
type IHost interface {
	Close()
}
