package main

//An example of event unikernel which expects messages from trigger

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	conn := Conn()
	defer conn.Close()

	topic := "mytopic"

	_, err := conn.Subscribe(topic, func(message *nats.Msg) {
		fmt.Printf("RECEIVED MESSAGE FROM TRIGGER: %s\n", string(message.Data))
	})
	if err != nil {
		log.Fatal(err)
	}
	select {}
}

func Conn() *nats.Conn {
	conn, err := nats.Connect("10.0.2.2:4222")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
