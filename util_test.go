package apikit
import (
	"testing"
	"fmt"
)

var _ = fmt.Println

func TestEmbedsRESTController(t *testing.T) {
	if embedsRESTController(1) {
		t.Error("non-structs should not embed REST controllers")
	}

	if embedsRESTController(ExampleUser{}) {
		t.Error("ExampleUser does not embed REST controller")
	}

	if !embedsRESTController(ExampleUserController{}) {
		t.Error("ExampleUserController does embed REST controller")
	}

	if !embedsRESTController(&ExampleUserController{}) {
		t.Error("Ptr to ExampleUserController does embed REST controller")
	}
}