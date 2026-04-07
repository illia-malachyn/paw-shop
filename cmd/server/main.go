package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/illia-malachyn/paw-shop/internal/handler"
)

func main() {
	productHandler := handler.NewProductHandler()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/products", productHandler.HandleProducts)

	port := ":8080"
	fmt.Printf("PawShop server starting on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
