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

	svcconf "github.com/AyushSenapati/reactive-micro/authzsvc/conf"
	svcep "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/endpoint"
	cl "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/logger"
	svcrepo "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/repo"
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/service"
	httptransport "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/transport/http"
	natstransport "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/transport/nats"
	kitep "github.com/go-kit/kit/endpoint"
	"github.com/nats-io/nats.go"
	"github.com/oklog/run"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var fs = flag.NewFlagSet("user", flag.ExitOnError)
var httpAddr = fs.String("http-addr", ":8083", "HTTP listen address")

// holds the name of all the endpoints that ther service supports
var allMethods = []string{"UpsertPolicy", "ListPolicy", "RemovePolicy", "RemovePolicyBySub"}

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

	// Get Mongo client to setup service repo
	mongoClient := getMongoClient(ctx, confObj)
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Info(ctx, fmt.Sprintf("mongo: error disconnecting client [%v]", err))
			return
		}
		logger.Info(ctx, "mongo: client disconnected")
	}()

	// Get NATS json encoded connection object
	nc := getNATSEncodedConn(confObj)
	defer func() {
		nc.Close()
		logger.Info(ctx, "nats: disconnected")
	}()
	logger.Info(ctx, "nats: connected")

	// initialize service repo
	repoObj := svcrepo.NewAuthzRepo(mongoClient)
	if repoObj == nil {
		logger.Info(ctx, "error in initialising service repository")
		return
	}

	// initialise service
	svcConfigs := []service.SvcConf{
		service.WithRepo(repoObj),
		service.WithNATSEncodedConn(nc),
	}
	svc := service.New(logger, svcConfigs...)
	if svc == nil {
		logger.Error(ctx, "error initialising service")
		return
	}

	// initialise endpoint
	eps := svcep.New(svc, getEndpointMW(confObj))

	g := &run.Group{}
	initEventHandler(logger, svc, nc, g) // initialise NATS transport
	initHttpHandler(logger, eps, g)      // initialise HTTP transport
	initCancelInterrupt(g)               // prepare listening OS interrupt signal
	err = g.Run()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("final err: %v", err))
	}
}

func getEndpointMW(c *svcconf.Config) (mw map[string][]kitep.Middleware) {
	mw = map[string][]kitep.Middleware{}
	return
}

func initHttpHandler(logger *cl.CustomLogger, endpoints svcep.Endpoints, g *run.Group) {
	options := defaultHttpOptions()

	// Add your http options here
	// ...

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

func initEventHandler(logger *cl.CustomLogger, svc service.IAuthzService, nc *nats.EncodedConn, g *run.Group) {
	eventHandler := natstransport.NewEventHandler(logger, nc, svc)
	g.Add(eventHandler.Execute, eventHandler.Interrupt)
}

func getMongoClient(ctx context.Context, c *svcconf.Config) *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(c.MongoURI))
	if err != nil {
		panic(err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	return client
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
