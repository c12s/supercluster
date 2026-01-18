package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

func writeErrorResp(err error, w http.ResponseWriter) {
	if err == nil {
		return
	} else if strings.Contains(err.Error(), "not found") {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte(err.Error()))
}

func writeResp(resp any, w http.ResponseWriter) {
	w.WriteHeader(http.StatusCreated)
	if resp == nil {
		return
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(respBytes)
}

func writeRespOK(resp any, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	if resp == nil {
		return
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(respBytes)
}

func readReq(req any, r *http.Request, w http.ResponseWriter) error {
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	return err
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
