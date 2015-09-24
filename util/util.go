package util

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSONDecode(r *http.Request) (map[string]interface{}, error) {
	req := make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func JSONResponse(w http.ResponseWriter, r map[string]interface{}) {
	log.Printf("Responding with JSON: %v\n", r)
	if b, err := json.Marshal(r); err == nil {
		w.Write(b)
	} else {
		log.Printf("Error writing JSON response: %v\n", err)
	}
}
