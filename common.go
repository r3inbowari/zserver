package zserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RequestResult struct {
	Total   int         `json:"total"`
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
	Message string      `json:"msg"`
}

func ResponseCommon(w http.ResponseWriter, data interface{}, msg string, total int, tag int, code int) {
	var rq RequestResult
	rq.Data = data
	rq.Total = total
	rq.Code = code
	rq.Message = msg
	jsonStr, err := json.Marshal(rq)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	w.WriteHeader(tag)
	_, _ = fmt.Fprintf(w, string(jsonStr))
}
