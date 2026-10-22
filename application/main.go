package main

//An example application which serves as a reciever of messages sent from a unikernel

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	conn := Conn()
	defer conn.Close()

	subject := "novaapp3+novaapp1/Root/folder0"

	err := conn.Publish(subject, []byte("Message to stored procedure!"))
	if err != nil {
		log.Fatal(err)
	}

}
func Conn() *nats.Conn {
	conn, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
