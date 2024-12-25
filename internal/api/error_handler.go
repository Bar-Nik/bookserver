package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func handleError(w http.ResponseWriter, status int, err error) {
	result := map[string]interface{}{
		"error":  err.Error(),
		"status": http.StatusText(status),
	}
	data, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	w.WriteHeader(status)
	_, err = w.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
}
