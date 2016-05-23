# revel-apikit
A way to rapidly develop REST APIs using the [Revel](https://github.com/revel/revel) framework

### Goal
Writing controllers for a Revel REST API can get repetitive. 
You might find yourself writing `Get()`, `Create()`, `Update()`, and `Delete()` for every model in your database.

This package tries to abstract away that CRUD boilerplate. 
The following is an example of how you might implement a controller for just one of your models, a `User`:
```Go
type UserController struct
  *revel.Controller
  authenticatedUser *User
}

func (c *UserController) Get(id uint64) revel.Result {
	if user := GetUserByID(id); user == nil {
		return ApiMessage{
			StatusCode: http.StatusNotFound,
			Message: fmt.Sprint("User with ID ", id, " not found"),
		}
	} else if !user.CanBeViewedBy(c.authenticatedUser) {
		return ApiMessage{
			StatusCode: http.StatusUnauthorized,
			Message: fmt.Sprint("Unauthorized to view User with ID ", id),
		}
	} else {
		return c.RenderJson(user)
	}
}

func (c *UserController) Post() revel.Result {
	newUser := User{}
	err := json.NewDecoder(c.Request.Body).Decode(&newUser)
	if err != nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: "Improperly formatted request body",
			}
	}
	
	if !newUser.CanBeModifiedBy(c.authenticatedUser) {
		return ApiMessage{
			StatusCode: http.StatusUnauthorized,
			Message: "Not authorized to post this User",
		}
	}
	if err = newUser.Save(); err != nil {
		return ApiMessage{
			StatusCode: http.StatusBadRequest,
			Message: err.Error(),
		}
	} else {
		return c.RenderJson(newUser)
	}
}

func (c *UserController) Put() revel.Result {
	updatedUser := User{}
	err := json.NewDecoder(c.Request.Body).Decode(&updatedUser)
	if err != nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: "Improperly formatted request body",
			}
	}

	if !updatedUser.CanBeModifiedBy(c.authenticatedUser) {
		return ApiMessage{
			StatusCode: http.StatusUnauthorized,
			Message: "Not authorized to modify this User",
		}
	}
	if err = updatedUser.Save(); err != nil {
		return ApiMessage{
			StatusCode: http.StatusBadRequest,
			Message: err.Error(),
		}
	} else {
		return c.RenderJson(updatedUser)
	}
}

func (c *UserController) Delete(id uint64) revel.Result {
	if user := GetUserByID(id); user == nil {
		return ApiMessage{
			StatusCode: http.StatusNotFound,
			Message: fmt.Sprint("User with ID ", id, " not found"),
		}
	} else {
		if !user.CanBeModifiedBy(c.authenticatedUser) {
			return ApiMessage{
				StatusCode: http.StatusUnauthorized,
				Message: "Not authorized to delete this User",
			}
		}
		if err := user.Delete(); err != nil {
			return ApiMessage{
				StatusCode: http.StatusBadRequest,
				Message: err.Error(),
			}
		} else {
			return ApiMessage{
				StatusCode: http.StatusOK,
				Message: "Success",
			}
		}
	}
}
```

The best way to solve this problem would be to create a generic struct,
but since Go [does not currently support generics](https://golang.org/doc/faq#generics),
this package tries to emulate a generic class using interfaces. 

With this package, you can gain all of the above functionality by embeding a `RESTController`
within a Revel controller that implements the `ModelProvider` interface. 
The `ModelProvider` interface serves as a workaround for Go's lack of generics. 
Using it gives you all of the above functionality with only the following code:
```Go
type UserController struct {
  *revel.Controller
  apikit.RESTController
}

func (c *UserController) ModelFactoryFunc() func() apikit.RESTObject {
	return func() apikit.RESTObject {
		return &User{}
	}
}

func (c *UserController) GetModelByIDFunc() func(id uint64) apikit.RESTObject {
	return func(id uint64) apikit.RESTObject {
		if u := GetUserByID(id); u == nil {
			return nil
		} else {
			return u
		}
	}
}
```

Doing this will allow your `UserController` to serve the following Revel routes:
```
# UserController
GET     /user/:id                               UserController.Get
DELETE  /user/:id                               UserController.Delete
POST    /user                                   UserController.Post
PUT     /user                                   UserController.Put
```

#### Limitations
- `RESTController` instances cannot have Actions other than those provided by `RESTController`
- `conf/restcontroller-routes` cannot use catchall `:` Actions
- `conf/restcontroller-routes` cannot use wildcard `*` paths

