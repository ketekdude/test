package middleware

import (
	"fmt"
	"net/http"
)

func HandlerWrapper(handler http.HandlerFunc, chains ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	return executeChain(handler, chains...)
}

func MiddlewareMetrics(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("MiddlewareMetrics Run")
		next(w, r)
	}
}

func executeChain(endHandler http.HandlerFunc, chains ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	if len(chains) == 0 {
		return endHandler
	}
	return chains[0](executeChain(endHandler, chains[1:]...))
}
