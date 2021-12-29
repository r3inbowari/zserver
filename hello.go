package zserver

import "net/http"

func Hello(w http.ResponseWriter, _ *http.Request) {
	ResponseCommon(w, `hello`, "ok", 1, http.StatusOK, 0)
}
