package rest

import (
	"log"
	"net/http"
	"test/load-test/pkg/middleware"
)

func Init() {
	http.HandleFunc("/run-test", middleware.HandlerWrapper(RunK6Test, middleware.MiddlewareMetrics))
	http.HandleFunc("/test", middleware.HandlerWrapper(Test))
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
