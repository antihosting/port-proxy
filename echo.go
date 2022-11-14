/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package main

import (
	"io/ioutil"
	"net/http"
)

type echoHandler struct {
}

func(t *echoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}


