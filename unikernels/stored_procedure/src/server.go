package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type PostDTO struct {
	Parm1 string `json:"parm1"`
}

func helloPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Post!\n")
	var postDTO PostDTO
	err := json.NewDecoder(r.Body).Decode(&postDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rez := postDTO.Parm1
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(rez))
}

func helloGet(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Get!\n")
	jsonResponse("povratna iz geta", w)
}

func jsonResponse(object interface{}, w http.ResponseWriter) {
	resp, err := json.Marshal(object)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", helloGet).Methods("GET")
	r.HandleFunc("/", helloPost).Methods("POST")
	fmt.Println("Listening on :8080...")
	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",
	}
	log.Fatal(srv.ListenAndServe())
}
