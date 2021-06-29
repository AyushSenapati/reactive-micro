# Setup NATS Jetstream
The supported events are fired on their respective streams and consumers consume those events from the stream based on some filters. The streams and the consumers are present in the NATS cluster.  
The streams must exist before the publishers publish events on those streams. To configure NATS Jetstream the given Dockerfile can be used to build an image on the top of natsio/natsbox (having tools to communicate with NATS), which contains a binary called `setup-nats-js`.  

To build the image, being in this directory run
```
$ docker build . -t reactive-micro/setup-nats-js:latest
```

Once the reactive-micro/setup-nats-js:latest is built, run the following command to create streams and consumers in NATS JS
```
$ docker run --rm reactive-micro/setup-nats-js:latest /bin/sh
```
This will drop you inside the container, where setup-nats-js command in available.  
`consumer-configs/`: contains configuration of all the consumers.  
`stream-configs/`: contains configuration of all the streams. 
```
$ setup-nats-js -nats-uri NATS_URL -streams-dir STREAM_CONFIG -consumers-dir CONSUMER_CONFIG
```
If you want to add new streams/consumers, add their config files in their respective folders, build the image again and run the above command to configure NATS JS.
