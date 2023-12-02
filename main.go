package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/volkankocaali/elastic-go/handler"
	"log"
	"net/http"
)

const ElasticSearchEndpoint = "http://localhost:9201"

func main() {
	// create elastic search client
	client, err := elastic.NewClient(elastic.SetURL(ElasticSearchEndpoint))
	if err != nil {
		log.Fatal(err)
	}

	// check elastic search connection
	info, code, err := client.Ping(ElasticSearchEndpoint).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// start app server , db connection and configure routes
	startApp(client)
}

func startApp(client *elastic.Client) {
	// create MySQL connection
	db, err := sqlx.Connect("mysql", "root:root@tcp(localhost:3307)/db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	handlerProduct := handler.NewProductHandler(db, client)
	r := mux.NewRouter()

	setupProductEndpoints(r, client, handlerProduct)

	fmt.Printf("Server started at http://localhost%s\n", ":8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func setupProductEndpoints(r *mux.Router, client *elastic.Client, handlerProduct *handler.ProductHandler) {
	r.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		handlerProduct.HandleGetProducts(w, r)
	}).Methods("GET")

	r.HandleFunc("/products/filter", func(w http.ResponseWriter, r *http.Request) {
		handlerProduct.HandleGetFilter(w, r, client)
	}).Methods("GET")

	r.HandleFunc("/products/create-indexes", func(w http.ResponseWriter, r *http.Request) {
		handlerProduct.HandleElasticIndex(w, r, client)
	}).Methods("GET")
}
