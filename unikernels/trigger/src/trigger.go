package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
)

type EventDTO struct {
	EventTopic string `json:"event"`
	AddEvent   bool   `json:"add"`
}

type EventHandler struct {
	EventTopics []string
	Conn        *nats.Conn
	mu          sync.RWMutex
}

func (handler *EventHandler) RegisterEvent(w http.ResponseWriter, r *http.Request) {
	var eventDTO EventDTO
	err := json.NewDecoder(r.Body).Decode(&eventDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	handler.mu.Lock()
	defer handler.mu.Unlock()

	topic := eventDTO.EventTopic
	if eventDTO.AddEvent {
		handler.EventTopics = append(handler.EventTopics, topic)
	} else {
		for i, t := range handler.EventTopics {
			if t == topic {
				handler.EventTopics = append(handler.EventTopics[:i], handler.EventTopics[i+1:]...)
				break
			}
		}
	}

	fmt.Println("Registration finished!")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(topic))
}

func (handler *EventHandler) trigger(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			handler.mu.RLock()
			topics := append([]string(nil), handler.EventTopics...)
			handler.mu.RUnlock()

			for _, topic := range topics {
				if err := handler.Conn.Publish(topic, []byte("message from trigger!")); err != nil {
					log.Println("publish failed:", err)
				}
			}

			fmt.Println("messages sent!")

		case <-ctx.Done():
			log.Println("trigger stopped")
			return
		}
	}
}

func main() {
	handler := EventHandler{EventTopics: []string{}, Conn: Conn()}

	r := mux.NewRouter()
	r.HandleFunc("/", handler.RegisterEvent).Methods("POST")
	ctx, cancel := context.WithCancel(context.Background())

	go handler.trigger(ctx)

	srv := &http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			fmt.Errorf("Cannot run server")
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	signal.Notify(sigCh, syscall.SIGTERM)

	//When we receive an interrupt or kill, if we don't have any current connections the code will terminate.
	//But if we do the code will stop receiving any new connections and wait for maximum of 30 seconds to finish all current requests.
	//After that the code will terminate.
	_ = <-sigCh

	timeoutContext, cancel1 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel1()
	cancel()
	//Try to shutdown gracefully
	if srv.Shutdown(timeoutContext) != nil {
		fmt.Errorf("Cannot gracefully shutdown...")
	}

}

func Conn() *nats.Conn {
	conn, err := nats.Connect("10.0.2.2:4222")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
