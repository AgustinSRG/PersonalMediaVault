// HTTP Server

package main

import (
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

func RunHTTPServer() {
	router := mux.NewRouter()

	// API routes

	// Static frontend

	frontend_path := os.Getenv("FRONTEND_PATH")

	if frontend_path == "" {
		frontend_path = "../frontend/dist/"
	}

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(frontend_path)))

	// Run server

	var wg sync.WaitGroup

	wg.Add(2)

	go runHTTPServer(&wg, router)
	go runHTTPSecureServer(&wg, router)

	wg.Wait()
}

func runHTTPSecureServer(wg *sync.WaitGroup, router *mux.Router) {
	defer wg.Done()

	bind_addr := os.Getenv("BIND_ADDRESS")

	// Setup HTTPS server
	var ssl_port int
	ssl_port = 443
	customSSLPort := os.Getenv("SSL_PORT")
	if customSSLPort != "" {
		sslp, e := strconv.Atoi(customSSLPort)
		if e == nil {
			ssl_port = sslp
		}
	}

	certFile := os.Getenv("SSL_CERT")
	keyFile := os.Getenv("SSL_KEY")

	if certFile != "" && keyFile != "" {
		// Listen
		LogInfo("[SSL] Listening on " + bind_addr + ":" + strconv.Itoa(ssl_port))
		errSSL := http.ListenAndServeTLS(bind_addr+":"+strconv.Itoa(ssl_port), certFile, keyFile, router)

		if errSSL != nil {
			LogError(errSSL)
		}
	}
}

func runHTTPServer(wg *sync.WaitGroup, router *mux.Router) {
	defer wg.Done()

	bind_addr := os.Getenv("BIND_ADDRESS")

	// Setup HTTP server
	var tcp_port int
	tcp_port = 80
	customTCPPort := os.Getenv("HTTP_PORT")
	if customTCPPort != "" {
		tcpp, e := strconv.Atoi(customTCPPort)
		if e == nil {
			tcp_port = tcpp
		}
	}

	// Listen
	LogInfo("[HTTP] Listening on " + bind_addr + ":" + strconv.Itoa(tcp_port))
	errHTTP := http.ListenAndServe(bind_addr+":"+strconv.Itoa(tcp_port), router)

	if errHTTP != nil {
		LogError(errHTTP)
	}
}
