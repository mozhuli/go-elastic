package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mozhuli/go-elastic/pkg/app/routers"
	//"github.com/mozhuli/go-elastic/pkg/app/utils"
	//"fmt"
)

func main() {
	/*
		Routing using mux
	*/

	//utils.GetNumCpu()
	r := mux.NewRouter()
	r.HandleFunc("/set", routers.SetHandler)
	r.HandleFunc("/get", routers.GetHandler)
	r.HandleFunc("/map", routers.MappingHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8000", r)

}
