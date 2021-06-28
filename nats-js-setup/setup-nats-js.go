package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

var fs = flag.NewFlagSet("setup-nats-js", flag.ExitOnError)
var natsURL = fs.String("nats-url", "", "nats url")
var strConfDir = fs.String("streams-dir", "/nats-js/stream-configs", "path to all streams config dir")
var conConfDir = fs.String("consumers-dir", "/nats-js/consumer-configs", "path to all consumers config dir")

func main() {
	fs.Parse(os.Args[1:])
	conFiles, err := ioutil.ReadDir(*conConfDir)
	if err != nil {
		log.Fatal(err)
	}

	conConfigFnames := []string{}

	for _, fo := range conFiles {
		if !fo.IsDir() && strings.Contains(fo.Name(), ".json") {
			conConfigFnames = append(conConfigFnames, path.Join(*conConfDir, fo.Name()))
		}
	}

	streamsCreated := 0
	failedCount := 0
	streamHash := map[string]bool{}

	for _, fname := range conConfigFnames {
		_, f := path.Split(fname)
		streamName := strings.Split(f, ".")[0]

		_, found := streamHash[streamName]
		if !found {
			err := runNatsStrAdd(streamName, path.Join(*strConfDir, fmt.Sprintf("%s.json", streamName)))
			if err != nil {
				log.Printf("error adding stream-%s [%v]", streamName, err)
			} else {
				streamHash[streamName] = true // stream created
				streamsCreated++
			}
		}

		err = runNatsConAdd(streamName, fname)
		if err != nil {
			log.Printf("error adding consumer [%v]\n", err)
			failedCount++
			continue
		}
	}

	fmt.Printf("\nNOTE:these counters show successful execution. if stream/consumer already exists they'll remain untouched\n")
	fmt.Println("streams added", streamsCreated)
	fmt.Printf("consumers added [%d/%d]\n", len(conConfigFnames)-failedCount, len(conConfigFnames))
}

func runNatsConAdd(streamName, conConfigName string) (err error) {
	cmd := exec.Command(
		"nats", "-s", *natsURL, "con", "add", streamName, "--config", conConfigName,
	)
	log.Println("cmd:", cmd.String())

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	return
}

func runNatsStrAdd(streamName, strConfigName string) (err error) {
	cmd := exec.Command(
		"nats", "-s", *natsURL, "str", "add", streamName, "--config", strConfigName,
	)
	log.Println("cmd:", cmd.String())

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	return
}
