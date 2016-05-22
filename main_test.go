package apikit
import (
	"testing"
	"os"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"
	"github.com/revel/revel"
	"path"
	"runtime"
)

const (
	testPort int = 62937
)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	revel.ConfPaths = []string{path.Join(cwd, "conf")}
	revel.Config = revel.NewEmptyConfig()
	conf, err := revel.LoadConfig("app.conf")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	revel.Config = conf

	revel.RegisterController((*ExampleUserController)(nil),
		[]*revel.MethodType{
			&revel.MethodType{
				Name: "Ok",
			},
		},
	)

	revel.RegisterController((*NonModelProviderConformingController)(nil),
		[]*revel.MethodType{
			&revel.MethodType{
				Name: "Ok",
			},
		},
	)

	revel.MainRouter = revel.NewRouter(path.Join(cwd, "conf", "routes"))
	revel.MainRouter.Refresh()

	go Run(testPort)
	time.Sleep(time.Millisecond * 100)
	os.Exit(m.Run())
}

// This method handles all requests.  It dispatches to handleInternal after
// handling / adapting websocket connections.
func handle(w http.ResponseWriter, r *http.Request) {
	if maxRequestSize := int64(revel.Config.IntDefault("http.maxrequestsize", 0)); maxRequestSize > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	}

	upgrade := r.Header.Get("Upgrade")
	if upgrade == "websocket" || upgrade == "Websocket" {
		websocket.Handler(func(ws *websocket.Conn) {
			//Override default Read/Write timeout with sane value for a web socket request
			ws.SetDeadline(time.Now().Add(time.Hour * 24))
			r.Method = "WS"
			handleInternal(w, r, ws)
		}).ServeHTTP(w, r)
	} else {
		handleInternal(w, r, nil)
	}
}

func handleInternal(w http.ResponseWriter, r *http.Request, ws *websocket.Conn) {
	var (
		req  = revel.NewRequest(r)
		resp = revel.NewResponse(w)
		c    = revel.NewController(req, resp)
	)
	req.Websocket = ws

	revel.Filters[0](c, revel.Filters[1:])

	if c.Result != nil {
		c.Result.Apply(req, resp)
	} else if c.Response.Status != 0 {
		c.Response.Out.WriteHeader(c.Response.Status)
	}
	// Close the Writer if we can
	if w, ok := resp.Out.(io.Closer); ok {
		w.Close()
	}
}

// Run the server.
// This is called from the generated main file.
// If port is non-zero, use that.  Else, read the port from app.conf.
func Run(port int) {
	address := revel.HttpAddr
	if port == 0 {
		port = revel.HttpPort
	}

	var network = "tcp"
	var localAddress string

	// If the port is zero, treat the address as a fully qualified local address.
	// This address must be prefixed with the network type followed by a colon,
	// e.g. unix:/tmp/app.socket or tcp6:::1 (equivalent to tcp6:0:0:0:0:0:0:0:1)
	if port == 0 {
		parts := strings.SplitN(address, ":", 2)
		network = parts[0]
		localAddress = parts[1]
	} else {
		localAddress = address + ":" + strconv.Itoa(port)
	}

	revel.Server = &http.Server{
		Addr:         localAddress,
		Handler:      http.HandlerFunc(handle),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	runStartupHooks()

	// Load templates
	revel.MainTemplateLoader = revel.NewTemplateLoader(revel.TemplatePaths)
	revel.MainTemplateLoader.Refresh()

	// Replace revel's PanicFilter with our own
	revel.Filters = revel.Filters[1:]
	revel.Filters = append([]revel.Filter{testingPanicFilter}, revel.Filters...)

	// The "watch" config variable can turn on and off all watching.
	// (As a convenient way to control it all together.)
	if revel.Config.BoolDefault("watch", true) {
		revel.MainWatcher = revel.NewWatcher()
		revel.Filters = append([]revel.Filter{revel.WatchFilter}, revel.Filters...)
	}

	// If desired (or by default), create a watcher for templates and routes.
	// The watcher calls Refresh() on things on the first request.
	if revel.MainWatcher != nil && revel.Config.BoolDefault("watch.templates", true) {
		//////MainWatcher.Listen(MainTemplateLoader, MainTemplateLoader.paths...)
		revel.MainWatcher.Listen(revel.MainTemplateLoader)
	}

	// add the type injection filter
	var modelFilter revel.Filter = CreateRESTControllerInjectionFilter(authenticationFunc)
	filterCount := len(revel.Filters)
	revel.Filters = append(revel.Filters[:filterCount-1],
		append([]revel.Filter{modelFilter}, revel.Filters[filterCount-1:]...)...)

	go func() {
		//////time.Sleep(100 * time.Millisecond)
		fmt.Printf("Listening on %s...\n", localAddress)
	}()

	if revel.HttpSsl {
		if network != "tcp" {
			// This limitation is just to reduce complexity, since it is standard
			// to terminate SSL upstream when using unix domain sockets.
			revel.ERROR.Fatalln("SSL is only supported for TCP sockets. Specify a port to listen on.")
		}
		revel.ERROR.Fatalln("Failed to listen:",
			revel.Server.ListenAndServeTLS(revel.HttpSslCert, revel.HttpSslKey))
	} else {
		listener, err := net.Listen(network, localAddress)
		if err != nil {
			revel.ERROR.Fatalln("Failed to listen:", err)
		}
		revel.ERROR.Fatalln("Failed to serve:", revel.Server.Serve(listener))
	}
}

func runStartupHooks() {
	for _, hook := range startupHooks {
		hook()
	}
}

var startupHooks []func()

func testingPanicFilter(c *revel.Controller, fc []revel.Filter) {
	defer func() {
		if err := recover(); err != nil {
			//revel.ERROR.Print(err, "\n", string(debug.Stack()))
			c.Response.Out.WriteHeader(http.StatusInternalServerError)
			c.Response.Out.Write([]byte("An unexpected error ocurred"))
		}
	}()

	fc[0](c, fc[1:])
}

func authenticationFunc(username, password string) User {
	return nil
}
