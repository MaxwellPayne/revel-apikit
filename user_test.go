package apikit
import "github.com/revel/revel"

type ExampleUser struct {
	ID       uint64    `json:"id"`
	Username string    `json:"username"`
}

var mockUuser *ExampleUser = &ExampleUser{
	ID: 1,
	Username: "MaxwellPayne",
}

func (u *ExampleUser) CanBeViewedBy(other *User) bool {
	return true
}

func (u *ExampleUser) CanBeModifiedBy(other *User) bool {
	return other != nil && u.Username == u.Username
}

func (u *ExampleUser) Validate(v *revel.Validation) {

}

func (u *ExampleUser) UniqueID() uint64 {
	return u.ID
}

func (u *ExampleUser) EnableGET() bool {
	return true
}

func (u *ExampleUser) EnablePOST() bool {
	return true
}

func (u *ExampleUser) EnablePUT() bool {
	return true
}

func (u *ExampleUser) EnableDELETE() bool {
	return true
}

func (u *ExampleUser) Delete() error {
	return nil
}

func (u *ExampleUser) Save() error {
	return nil
}
