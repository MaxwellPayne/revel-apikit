package apikit
import (
	"github.com/revel/revel"
	"reflect"
	"path"
	"io/ioutil"
	"strings"
	"regexp"
	"errors"
	"github.com/robfig/pathtree"
)

// Register the RESTControllers
func RegisterRESTControllers(controllers []RESTController) {
	revel.MainRouter = revel.NewRouter(path.Join(revel.BasePath, "conf", "routes"))
	revel.MainRouter.Refresh()

	for _, c := range controllers {
		revel.RegisterController(c,
			[]*revel.MethodType{
				&revel.MethodType{
					Name: "Get",
					Args: []*revel.MethodArg{
						{"id", reflect.TypeOf((*uint64)(nil))},
					},
				},
				&revel.MethodType{
					Name: "Post",
				},
				&revel.MethodType{
					Name: "Put",
				},
				&revel.MethodType{
					Name: "Delete",
					Args: []*revel.MethodArg{
						{"id", reflect.TypeOf((*uint64)(nil))},
					},
				},
			},
		)
	}

	restcontrollerPath := path.Join(revel.BasePath, "conf", "restcontroller-routes")
	restcontrollerData, err := ioutil.ReadFile(restcontrollerPath)
	if err != nil {
		panic(err)
	}

	restcontrollerRoutes, err := parseRoutes(restcontrollerPath, "", string(restcontrollerData))
	if err != nil {
		panic(err)
	}

	revel.MainRouter.Routes = append(revel.MainRouter.Routes, restcontrollerRoutes...)
	updateTree(revel.MainRouter)
}

func parseRoutes(routesPath, joinedPath, content string) ([]*revel.Route, error) {
	var routes []*revel.Route

	// For each line..
	for n, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		const modulePrefix = "module:"

		// Handle included routes from modules.
		// e.g. "module:testrunner" imports all routes from that module.
		if strings.HasPrefix(line, modulePrefix) {
			continue
		}

		// A single route
		method, path, action, fixedArgs, found := parseRouteLine(line)
		if !found {
			continue
		}

		// this will avoid accidental double forward slashes in a route.
		// this also avoids pathtree freaking out and causing a runtime panic
		// because of the double slashes
		if strings.HasSuffix(joinedPath, "/") && strings.HasPrefix(path, "/") {
			joinedPath = joinedPath[0 : len(joinedPath)-1]
		}
		//path = strings.Join([]string{revel.BasePath, joinedPath, path}, "")
		if strings.Contains(action, ":") {
			return nil, errors.New("revel-apikit does not yet support catchall (:) actions")
		}

		// This will import the module routes under the path described in the
		// routes file (joinedPath param). e.g. "* /jobs module:jobs" -> all
		// routes' paths will have the path /jobs prepended to them.
		// See #282 for more info
		if strings.Contains(line, "*") {
			return nil, errors.New("revel-apikit does not yet support wildcard (*) routes")
		}

		route := revel.NewRoute(method, path, action, fixedArgs, routesPath, n)
		routes = append(routes, route)
	}
	return routes, nil
}

// Groups:
// 1: method
// 4: path
// 5: action
// 6: fixedargs
var routePattern *regexp.Regexp = regexp.MustCompile(
	"(?i)^(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD|WS|\\*)" +
	"[(]?([^)]*)(\\))?[ \t]+" +
	"(.*/[^ \t]*)[ \t]+([^ \t(]+)" +
	`\(?([^)]*)\)?[ \t]*$`)

func parseRouteLine(line string) (method, path, action, fixedArgs string, found bool) {
	var matches []string = routePattern.FindStringSubmatch(line)
	if matches == nil {
		return
	}
	method, path, action, fixedArgs = matches[1], matches[4], matches[5], matches[6]
	found = true
	return
}

func updateTree(router *revel.Router) error {
	router.Tree = pathtree.New()
	for _, route := range router.Routes {
		err := router.Tree.Add(route.TreePath, route)

		// Allow GETs to respond to HEAD requests.
		if err == nil && route.Method == "GET" {
			err = router.Tree.Add(treePath("HEAD", route.Path), route)
		}

		// Error adding a route to the pathtree.
		if err != nil {
			return errors.New(err.Error())
		}
	}
	return nil
}

func treePath(method, path string) string {
	if method == "*" {
		method = ":METHOD"
	}
	return "/" + method + path
}
