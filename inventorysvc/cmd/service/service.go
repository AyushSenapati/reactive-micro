package service

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	svcconf "github.com/AyushSenapati/reactive-micro/inventorysvc/conf"
	"github.com/nats-io/nats.go"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	svcep "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/endpoint"
	svcpe "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/lib/policy-enforcer"
	svcrepo "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/repo"
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/service"
	httptransport "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/transport/http"
	natstransport "github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/transport/nats"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	kitep "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/oklog/run"
)

var fs = flag.NewFlagSet("user", flag.ExitOnError)
var httpAddr = fs.String("http-addr", ":8084", "HTTP listen address")

// holds the name of the protected methods
var securedMethods = []string{"CreateMerchant", "ListMerchant", "CreateProduct", "ListProduct"}

// holds the name of all the endpoints that ther service supports
var allMethods = []string{"CreateMerchant", "ListMerchant", "CreateProduct", "ListProduct"}

// holds the database table names that the service is dealing with
var allResourceTypes = []string{"merchants", "products", "reserved_products"}

func Run() {
	fs.Parse(os.Args[1:])

	confObj, err := svcconf.Load("")

	if err != nil {
		fmt.Println("error in loading svc configuration", err)
		os.Exit(1)
	}

	serializedConf, _ := json.Marshal(confObj)
	fmt.Println(string(serializedConf))

	// Get NATS json encoded connection object
	nc := getNATSEncodedConn(confObj)
	defer func() {
		nc.Close()
		fmt.Println("nats: disconnected")
	}()

	// get gorm client to setup service repo
	db := getDBConn(confObj.GetDSN())

	// initialise service repo
	repoObj := svcrepo.NewBasicOrderRepo(db)

	// intialise policy enforcer
	c := cache.New(5*time.Minute, 10*time.Minute)
	ps, err := svcpe.NewCachedPolicyStorageMW(confObj.AuthzSvcUrl, allResourceTypes, c)
	if err != nil {
		fmt.Println("error initialising policy storage. err:", err)
		return
	}

	// initialise service
	svcConfigs := []service.SvcConf{
		service.WithRepo(repoObj),
		service.WithNATSEncodedConn(nc),
		service.WithPolicyStorage(ps),
	}
	svc := service.New(getServiceMiddleware(confObj, ps), svcConfigs...)
	if svc == nil {
		fmt.Println("error initialising service")
		return
	}

	// initialise endpoint
	eps := svcep.New(svc, getEndpointMW(confObj))

	g := &run.Group{}
	initEventHandler(svc, nc, g)
	initHttpHandler(eps, g)
	initCancelInterrupt(g)
	err = g.Run()
	if err != nil {
		fmt.Println("final err:", err)
	}
}

func getServiceMiddleware(c *svcconf.Config, ps svcpe.PolicyStorage) (mw []service.Middleware) {
	mw = []service.Middleware{}
	// Append your middleware here

	// add authorization middleware
	pe, err := svcpe.NewPolicyEnforcer(ps)
	if err != nil {
		fmt.Println("error initialising policy enforcer, err:", err)
		return
	}
	mw = append(mw, service.NewAuthzMW(pe))

	return
}

func getEndpointMW(c *svcconf.Config) (mw map[string][]kitep.Middleware) {
	mw = map[string][]kitep.Middleware{}

	// enforce jwt token parsing middleware on secure endpoints
	for _, method := range securedMethods {
		mw[method] = append(
			mw[method], svcep.NewJWTTokenParsingMW(c.Auth.SecretKey))
	}

	return
}

func initHttpHandler(endpoints svcep.Endpoints, g *run.Group) {
	options := defaultHttpOptions()

	// Add your http options here
	// ...

	// Extract token from the request header and put into the context
	for _, method := range securedMethods {
		options[method] = append(
			options[method], kithttp.ServerBefore(kitjwt.HTTPToContext()))
	}

	httpHandler := httptransport.NewHTTPHandler(endpoints, options)
	nl, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		fmt.Println("transport err: err during listing on specified address")
		return
	}
	g.Add(func() error {
		fmt.Println("listening at", *httpAddr)
		return http.Serve(nl, httpHandler)
	}, func(err error) {
		fmt.Println("err:", err)
		nl.Close()
		fmt.Println("closed the HTTP listener")
	})
}

func initCancelInterrupt(g *run.Group) {
	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})
}

func initEventHandler(svc service.IInventoryService, nc *nats.EncodedConn, g *run.Group) {
	eventHandler := natstransport.NewEventHandler(nc, svc)
	g.Add(eventHandler.Execute, eventHandler.Interrupt)
}

func getDBConn(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func getNATSEncodedConn(c *svcconf.Config) *nats.EncodedConn {
	opts := []nats.Option{nats.Name(c.SVCName)}
	conn, err := nats.Connect(c.NATSUrl, opts...)
	if err != nil {
		panic(err)
	}
	encodedConn, err := nats.NewEncodedConn(conn, "json")
	if err != nil {
		panic(err)
	}
	fmt.Println("nats: connected")
	return encodedConn
}
