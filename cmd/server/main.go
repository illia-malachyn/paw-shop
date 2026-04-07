package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/illia-malachyn/paw-shop/internal/handler"
)

func main() {
	productHandler := handler.NewProductHandler()
	bundleHandler := handler.NewBundleHandler()
	discountHandler := handler.NewDiscountHandler()
	orderHandler := handler.NewOrderHandler()
	searchHandler := handler.NewSearchHandler()
	chatHandler := handler.NewChatHandler()
	cartHandler := handler.NewCartHandler()
	notificationHandler := handler.NewNotificationHandler()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/products", productHandler.HandleProducts)
	http.HandleFunc("/api/products/search", searchHandler.HandleSearch)
	http.HandleFunc("/api/bundles", bundleHandler.HandleBuild)
	http.HandleFunc("/api/bundles/templates", bundleHandler.HandleTemplates)
	http.HandleFunc("/api/bundles/clone", bundleHandler.HandleClone)
	http.HandleFunc("/api/discounts/apply", discountHandler.HandleApply)
	http.HandleFunc("/api/discounts/undo", discountHandler.HandleUndo)
	http.HandleFunc("/api/products/", discountHandler.HandleSubscribe) // /api/products/{id}/subscribe
	http.HandleFunc("/api/orders/batch", orderHandler.HandleBatch)
	http.HandleFunc("/api/orders/", orderHandler.HandleOrders)
	http.HandleFunc("/api/reports/", orderHandler.HandleReport)
	http.HandleFunc("/api/chat/", chatHandler.HandleChat)
	http.HandleFunc("/api/cart", cartHandler.HandleCart)
	http.HandleFunc("/api/cart/", cartHandler.HandleCart)
	http.HandleFunc("/api/notifications/send", notificationHandler.HandleNotify)
	http.HandleFunc("/api/logs", notificationHandler.HandleLogs)
	http.HandleFunc("/api/logs/stats", notificationHandler.HandleLogStats)

	port := ":8080"
	fmt.Printf("PawShop server starting on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
