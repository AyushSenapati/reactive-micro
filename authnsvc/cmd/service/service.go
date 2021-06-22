package service

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	svcconf "github.com/AyushSenapati/reactive-micro/authnsvc/conf"
	svcep "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/endpoint"
	svcpe "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/lib/policy-enforcer"
	cl "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/logger"
	svcrepo "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/repo"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/service"
	httptransport "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/transport/http"
	natstransport "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/transport/nats"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	kitep "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/nats-io/nats.go"
	"github.com/oklog/run"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var fs = flag.NewFlagSet("authnsvc", flag.ExitOnError)
var httpAddr = fs.String("http-addr", "0.0.0.0:8081", "HTTP listen address")

// holds the name of the protected methods
var securedMethods = []string{"ListAccount", "DeleteAccount"}

// holds the name of all the endpoints that ther service supports
var allMethods = []string{"Login", "CreateAccount", "ListAccount", "DeleteAccount"}

// holds the database table names that the service is dealing with
var allResourceTypes = []string{"accounts", "roles"}

func Run() {
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	confObj, err := svcconf.Load("")

	if err != nil {
		fmt.Println("error in loading svc configuration", err)
		os.Exit(1)
	}

	serializedConf, _ := json.Marshal(confObj)
	fmt.Println(string(serializedConf))

	logger := cl.NewLogger(confObj.Env)
	logger.Configure(
		cl.WithSvcName(confObj.SVCName),
		cl.WithTimeStamp(),
	)

	// Get NATS json encoded connection object
	nc := getNATSEncodedConn(confObj)
	defer func() {
		nc.Close()
		logger.Info(ctx, "nats: disconnected")
	}()
	logger.Info(ctx, "nats: connected")

	// get gorm client to setup service repo
	db := getDBConn(confObj.GetDSN())

	// initialise service repo
	repoObj := svcrepo.NewBasicUserRepo(db)
	if repoObj == nil {
		logger.Info(ctx, "error in initialising service repository")
		return
	}

	// intialise policy enforcer
	c := cache.New(5*time.Minute, 10*time.Minute)
	ps, err := svcpe.NewCachedPolicyStorageMW(confObj.AuthzSvcUrl, allResourceTypes, c)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error initialising policy storage [%v]", err))
		return
	}

	// initialise service
	svcConfigs := []service.SvcConf{
		service.WithRepo(repoObj),
		service.WithNATSEncodedConn(nc),
		service.WithPolicyStorage(ps),
	}
	svc := service.New(logger, getServiceMiddleware(confObj, ps), svcConfigs...)
	if svc == nil {
		logger.Error(ctx, "error initialising service")
		return
	}

	// initialise endpoint
	eps := svcep.New(svc, getEndpointMW(confObj))

	g := &run.Group{}
	initEventHandler(logger, svc, nc, g)
	initHttpHandler(logger, eps, g)
	initCancelInterrupt(g)
	err = g.Run()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("final err: %v", err))
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

func initHttpHandler(logger *cl.CustomLogger, endpoints svcep.Endpoints, g *run.Group) {
	options := defaultHttpOptions()

	// Add your http options here

	// Extract token from the request header and put into the context
	for _, method := range securedMethods {
		options[method] = append(
			options[method], kithttp.ServerBefore(kitjwt.HTTPToContext()))
	}

	httpHandler := httptransport.NewHTTPHandler(endpoints, options)
	nl, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		logger.Error(context.TODO(), "transport [HTTP]: err during listing on specified address")
		return
	}
	g.Add(func() error {
		logger.Info(context.TODO(), fmt.Sprintf("transport [HTTP]: listening at %s", *httpAddr))
		return http.Serve(nl, httpHandler)
	}, func(err error) {
		logger.Error(context.TODO(), fmt.Sprintf("transport [HTTP]: %v", err))
		nl.Close()
		logger.Info(context.TODO(), "transport [HTTP]: closed the HTTP listener")
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

func initEventHandler(logger *cl.CustomLogger, svc service.IAuthNService, nc *nats.EncodedConn, g *run.Group) {
	eventHandler := natstransport.NewEventHandler(logger, nc, svc)
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
	return encodedConn
}
