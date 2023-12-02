package handler

import (
	"context"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"log"
	"net/http"
	"strconv"
)

// ProductHandler struct contains dependencies for the HandleGetProducts and HandleElasticIndex functions.
type ProductHandler struct {
	DB     *sqlx.DB
	Client *elastic.Client
}

func NewProductHandler(db *sqlx.DB, client *elastic.Client) *ProductHandler {
	return &ProductHandler{
		DB:     db,
		Client: client,
	}
}

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

func (h *ProductHandler) HandleGetProducts(w http.ResponseWriter, r *http.Request) {
	// handle GET products using h.DB

	// insertProductTable(db)
	// insertProductData(db)

	// return get all products data
	products := getProducts(h.DB)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(products)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (h *ProductHandler) HandleElasticIndex(w http.ResponseWriter, r *http.Request, client *elastic.Client) {
	// handle Elastic index using h.DB and h.Client
	products := getProducts(h.DB)

	for _, product := range products {
		_, err := client.Index().
			Index("products_index").
			BodyJson(product).
			Do(context.Background())

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (h *ProductHandler) HandleGetFilter(w http.ResponseWriter, r *http.Request, client *elastic.Client) {
	category := r.URL.Query().Get("category")
	priceStr := r.URL.Query().Get("price")
	name := r.URL.Query().Get("name")

	// string price to float float64
	maxPrice, err := strconv.ParseFloat(priceStr, 64)

	if err != nil {
		maxPrice = 0.0
	}

	filteredProducts, err := filterProduct(client, category, name, maxPrice)
	if err != nil {
		http.Error(w, "Error while filtering products", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filteredProducts)
}

func getProducts(db *sqlx.DB) []Product {
	var products []Product
	err := db.Select(&products, "select * from products")
	if err != nil {
		log.Fatal(err)
	}

	return products
}

func filterProduct(client *elastic.Client, category string, name string, maxPrice float64) ([]Product, error) {
	query := elastic.NewBoolQuery()

	if category != "" {
		query = query.Must(elastic.NewMatchQuery("category", category))
	}

	if name != "" {
		query = query.Must(elastic.NewMatchQuery("name", name))
	}

	if maxPrice > 0.0 {
		query = query.Filter(elastic.NewRangeQuery("price").Lte(maxPrice))
	}

	searchResult, err := client.Search().
		Index("products_index").
		Query(query).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	var products []Product
	for _, hit := range searchResult.Hits.Hits {
		var product Product
		err := json.Unmarshal(hit.Source, &product)
		if err != nil {
			log.Println("Error unmarshalling JSON:", err)
		}
		products = append(products, product)
	}

	return products, nil
}

func insertProductTable(db *sqlx.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255),
			category VARCHAR(50),
			price INT
		)
	`)

	if err != nil {
		log.Fatal(err)
	}
}

func insertProductData(db *sqlx.DB) {
	for i := 0; i < 1000; i++ {
		name := gofakeit.BeerName()
		category := gofakeit.BeerStyle()
		price := gofakeit.Price(100, 1000)

		_, err := db.Exec("INSERT INTO products (name, category, price) VALUES (?, ?, ?)",
			name, category, price)
		if err != nil {
			log.Fatal(err)
		}
	}
}
