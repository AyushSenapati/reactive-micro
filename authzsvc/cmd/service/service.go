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
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/endpoint"
	svcrepo "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/repo"
	"github.com/AyushSenapati/reactive-micro/authzsvc/pkg/service"
	httptransport "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/transport/http"
	natstransport "github.com/AyushSenapati/reactive-micro/authzsvc/pkg/transport/nats"
	kitep "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/nats-io/nats.go"
	"github.com/oklog/run"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var fs = flag.NewFlagSet("user", flag.ExitOnError)
var httpAddr = fs.String("http-addr", ":8083", "HTTP listen address")

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

	// Get Mongo client to setup service repo
	mongoClient := getMongoClient(ctx, confObj)
	defer func() {
		fmt.Printf("mongo: disconnecting client...")
		if err := mongoClient.Disconnect(ctx); err != nil {
			fmt.Printf(" failed [%v]\n", err)
		}
		fmt.Println(" done")
	}()

	// Get NATS json encoded connection object
	nc := getNATSEncodedConn(confObj)
	defer func() {
		nc.Close()
		fmt.Println("nats: disconnected")
	}()

	// initialize service repo
	repoObj := svcrepo.NewAuthzRepo(mongoClient)
	if repoObj == nil {
		fmt.Println("failed initialising service repository")
		return
	}

	// initialise service
	svcConfigs := []service.SvcConf{
		service.WithRepo(repoObj),
		service.WithNATSEncodedConn(nc),
	}
	svc := service.New(svcConfigs...)
	if svc == nil {
		fmt.Println("error initialising service")
		return
	}

	// initialise endpoint
	eps := endpoint.New(svc, getEndpointMW(confObj))

	g := &run.Group{}
	initEventHandler(svc, nc, g) // initialise NATS transport
	initHttpHandler(eps, g)      // initialise HTTP transport
	initCancelInterrupt(g)       // prepare listening OS interrupt signal
	err = g.Run()
	if err != nil {
		fmt.Println("final err:", err)
	}
}

func getEndpointMW(c *svcconf.Config) (mw map[string][]kitep.Middleware) {
	mw = map[string][]kitep.Middleware{}
	return
}

func initHttpHandler(endpoints endpoint.Endpoints, g *run.Group) {
	options := defaultHttpOptions()

	// Add your http options here
	// ...

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

func defaultHttpOptions() map[string][]kithttp.ServerOption {
	options := map[string][]kithttp.ServerOption{
		"ListPolicy": {kithttp.ServerErrorEncoder(httptransport.ErrorEncoder)},
	}
	return options
}

func initEventHandler(svc service.IAuthzService, nc *nats.EncodedConn, g *run.Group) {
	eventHandler := natstransport.NewEventHandler(nc, svc)
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
	fmt.Println("mongo: connected")
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
	fmt.Println("nats: connected")
	return encodedConn
}
