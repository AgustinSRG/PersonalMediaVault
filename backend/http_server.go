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

	// Auth API
	router.HandleFunc("/api/auth/login", api_handleAuthLogin).Methods("POST")
	router.HandleFunc("/api/auth/logout", api_handleAuthLogout).Methods("POST")

	// Account API (changing credentials)
	router.HandleFunc("/api/account/username", api_changeUsername).Methods("POST")
	router.HandleFunc("/api/account/password", api_changePassword).Methods("POST")

	// Assets API (get encrypted media files)
	router.HandleFunc("/assets/b/{mid:[0-9]+}/{asset:[0-9]+}/{filename}", api_handleAssetGet).Methods("GET", "HEAD")
	router.HandleFunc("/assets/hls/{mid:[0-9]+}/{asset:[0-9]+}/{filename}", api_handleAssetVideoHLS).Methods("GET")

	// Media API
	router.HandleFunc("/api/media/{mid:[0-9]+}", api_getMedia).Methods("GET")

	router.HandleFunc("/api/upload", api_uploadMedia).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/edit", api_editMedia).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/delete", api_deleteMedia).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/encode", api_mediaRequestEncode).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/resolution/add", api_mediaAddResolution).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/resolution/remove", api_mediaRemoveResolution).Methods("POST")

	// Search API
	router.HandleFunc("/api/search", api_searchMedia).Methods("GET")

	// Tags API
	router.HandleFunc("/api/tags", api_getTags).Methods("GET")

	router.HandleFunc("/api/tags/add", api_tagMedia).Methods("POST")
	router.HandleFunc("/api/tags/remove", api_untagMedia).Methods("POST")

	// Albums API
	router.HandleFunc("/api/albums", api_getAlbums).Methods("GET")
	router.HandleFunc("/api/albums/{id:[0-9]+}", api_getAlbum).Methods("GET")

	router.HandleFunc("/api/albums", api_createAlbum).Methods("POST")

	router.HandleFunc("/api/albums/{id:[0-9]+}/delete", api_deleteAlbum).Methods("POST")

	router.HandleFunc("/api/albums/{id:[0-9]+}/rename", api_renameAlbum).Methods("POST")

	router.HandleFunc("/api/albums/{id:[0-9]+}/set", api_setAlbumList).Methods("POST")

	router.HandleFunc("/api/albums/{id:[0-9]+}/add", api_albumAddMedia).Methods("POST")
	router.HandleFunc("/api/albums/{id:[0-9]+}/remove", api_albumRemoveMedia).Methods("POST")

	// Config API
	router.HandleFunc("/api/config", api_getConfig).Methods("GET")
	router.HandleFunc("/api/config", api_setConfig).Methods("POST")

	// Tasks API
	router.HandleFunc("/api/tasks", api_getTasks).Methods("GET")
	router.HandleFunc("/api/tasks/{id:[0-9]+}", api_getTask).Methods("GET")
	router.HandleFunc("/api/tasks/{id:[0-9]+}/kill", api_killTask).Methods("POST")

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
