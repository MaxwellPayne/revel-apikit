package models
import (
	"github.com/MaxwellPayne/revel-apikit"
	"github.com/revel/revel"
	"errors"
)

// A model that will be provided by a RESTController
type User struct {
	ID            uint64    `json:"id"`
	Username      string    `json:"username"`
	FavoriteColor string    `json:"favorite_color"`
	Password      string    `json:"-"`
}

// The authentication mechanism for our RESTControllers
var AuthenticationHandler apikit.AuthenticationFunction = func(username, password string) apikit.User {
	for _, u := range usersDB {
		// simulate a 'query' through our lame usersDB
		if u.Username == username && u.Password == password {
			return u
		}
	}
	return nil
}

// Implementation of RESTObject interface
func (u *User) CanBeViewedBy(other apikit.User) bool {
	return true
}

func (u *User) CanBeCreatedBy(other apikit.User) bool {
	return u.CanBeModifiedBy(other)
}

func (u *User) CanBeModifiedBy(other apikit.User) bool {
	return other != nil && u.UniqueID() == other.UniqueID()
}

func (u *User) CanBeDeletedBy(other apikit.User) bool {
	return u.CanBeModifiedBy(other)
}

func (u *User) IsNewRecord() bool {
	return u.ID == 0
}

func (u *User) Validate(v *revel.Validation) {
	if u.ID == 0 {
		v.Error("0 is not a valid User ID")
	}
	v.MinSize(u.Username, 1).Message("Username cannot be blank")
}

func (u *User) UniqueID() uint64 {
	return u.ID
}

func (u *User) HasAdminPrivileges() bool {
	return false
}

func (u *User) Delete() error {
	// not actually deleting users in this example
	return nil
}

func (u *User) Save() error {
	v := revel.Validation{}
	u.Validate(&v)
	if v.HasErrors() {
		return errors.New(v.Errors[0].String())
	}

	// not actually persisting data in this example
	return nil
}


// Other User-related methods and data not-specific to RESTControllers
func GetUserByID(id uint64) *User {
	for _, u := range usersDB {
		if u.ID == id {
			return u
		}
	}
	return nil
}

var usersDB []*User = []*User{
	&User{
		ID: 1,
		Username: "MaxwellPayne",
		FavoriteColor: "Red",
		Password: "banana",
	},
	&User{
		ID: 2,
		Username: "SmokeyTheBear",
		FavoriteColor: "Blue",
		Password: "orange",
	},
}
