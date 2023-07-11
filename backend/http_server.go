// HTTP Server

package main

import (
	"crypto/tls"
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	CORS_INSECURE_MODE_ENABLED = false // Insecure CORS mode for development and testing
)

// Logging middleware to log requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		LogRequest(r)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// CORS middleware, only applied when CORS_INSECURE_MODE_ENABLED = true
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		corsOrigin := r.Header.Get("origin")

		if corsOrigin == "" {
			corsOrigin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		allowMethods := r.Header.Get("access-control-request-method")

		if allowMethods != "" {
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)
		}

		allowHeaders := r.Header.Get("access-control-request-headers")

		if allowHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		}

		w.Header().Set("Access-Control-Max-Age", "86400")

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// Handler for the OPTIONS method, only when CORS_INSECURE_MODE_ENABLED = true
func corsHeadInsecure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// Adds cache TTL to static asset requests
func cacheTTLAdd(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "max-age=31536000")
		}
		next.ServeHTTP(w, r)
	})
}

// Runs HTTP server
// Creates mux router and launches the listener
// NOTE: This method locks the thread forever, run with coroutine: go RunHTTPServer(p, b)
// port - Port to listen
// bindAddr - Bind address
// isTest - True for testing, false for actual server
// Returns the router
func RunHTTPServer(port string, bindAddr string, isTest bool) *mux.Router {
	router := mux.NewRouter()

	// Logging middleware
	router.Use(loggingMiddleware)

	if CORS_INSECURE_MODE_ENABLED {
		LogWarning("CORS insecure mode enabled. Use this only for development")
		router.Use(corsMiddleware)
	}

	// API routes

	// Auth API
	router.HandleFunc("/api/auth/login", api_handleAuthLogin).Methods("POST")
	router.HandleFunc("/api/auth/logout", api_handleAuthLogout).Methods("POST")

	// Account API (changing credentials)
	router.HandleFunc("/api/account", api_getAccountContext).Methods("GET")
	router.HandleFunc("/api/account/username", api_getAccountContext).Methods("GET")
	router.HandleFunc("/api/account/username", api_changeUsername).Methods("POST")
	router.HandleFunc("/api/account/password", api_changePassword).Methods("POST")

	// Admin API (Manage accounts)
	router.HandleFunc("/api/admin/accounts", api_getAccounts).Methods("GET")
	router.HandleFunc("/api/admin/accounts", api_createAccount).Methods("POST")
	router.HandleFunc("/api/admin/accounts/delete", api_deleteAccount).Methods("POST")
	router.HandleFunc("/api/admin/launcher/{tag}", api_checkLauncherTag).Methods("GET")

	// Assets API (get encrypted media files)
	router.HandleFunc("/assets/b/{mid:[0-9]+}/{asset:[0-9]+}/{filename}", api_handleAssetGet).Methods("GET", "HEAD")
	router.HandleFunc("/assets/p/{mid:[0-9]+}/{asset:[0-9]+}/{filename}", api_handleAssetVideoPreviews).Methods("GET")

	// Media API
	router.HandleFunc("/api/media/{mid:[0-9]+}", api_getMedia).Methods("GET")
	router.HandleFunc("/api/media/{mid:[0-9]+}/albums", api_getMediaAlbums).Methods("GET")
	router.HandleFunc("/api/media/{mid:[0-9]+}/size_stats", api_getMediaSizeStats).Methods("GET")

	router.HandleFunc("/api/upload", api_uploadMedia).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/edit/title", api_editMediaTitle).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/edit/description", api_editMediaDescription).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/edit/extra", api_editMediaExtraParams).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/edit/thumbnail", api_editMediaThumbnail).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/edit/notes", api_setImageNotes).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/edit/time_slices", api_editMediaTimelineSplices).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/delete", api_deleteMedia).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/encode", api_mediaRequestEncode).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/resolution/add", api_mediaAddResolution).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/resolution/remove", api_mediaRemoveResolution).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/subtitles/set", api_addMediaSubtitles).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/subtitles/remove", api_removeMediaSubtitles).Methods("POST")

	router.HandleFunc("/api/media/{mid:[0-9]+}/audios/set", api_addMediaAudioTrack).Methods("POST")
	router.HandleFunc("/api/media/{mid:[0-9]+}/audios/remove", api_removeMediaAudioTrack).Methods("POST")

	// Search API
	router.HandleFunc("/api/search", api_searchMedia).Methods("GET")
	router.HandleFunc("/api/random", api_randomMedia).Methods("GET")

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

	// Is Test?

	if isTest {
		return router
	}

	// Static frontend

	frontend_path := os.Getenv("FRONTEND_PATH")

	if frontend_path == "" {
		frontend_path = "../frontend/dist/"
	}

	if CORS_INSECURE_MODE_ENABLED {
		router.Use(corsHeadInsecure)
	}

	mime.AddExtensionType(".js", "text/javascript") //nolint:errcheck

	router.Use(cacheTTLAdd)
	router.HandleFunc("/favicon.ico", handleFavicon).Methods("GET")
	router.HandleFunc("/img/icons/{file}", handleImageFile).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(frontend_path)))

	// Run server

	certFile := os.Getenv("SSL_CERT")
	keyFile := os.Getenv("SSL_KEY")

	if certFile != "" && keyFile != "" {
		// SSL
		runHTTPSecureServer(port, bindAddr, certFile, keyFile, router)
	} else {
		// Regular HTTP
		runHTTPServer(port, bindAddr, router)
	}

	return nil
}

// Runs Secure HTTPS server listener
// portOption - Port to listen
// bindAddr - Bind address
// certFile - Path to certificate file
// keyFile - Path to private key file
// router - Mux router
func runHTTPSecureServer(portOption string, bindAddr string, certFile string, keyFile string, router *mux.Router) {
	bind_addr := bindAddr

	// Setup HTTPS server
	var ssl_port int
	ssl_port = 443
	customSSLPort := portOption
	if customSSLPort != "" {
		sslp, e := strconv.Atoi(customSSLPort)
		if e == nil {
			ssl_port = sslp
		}
	}

	// Check keypair
	_, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		LogError(err)
		os.Exit(6)
	}

	// Listen
	LogInfo("[SSL] Listening on " + bind_addr + ":" + strconv.Itoa(ssl_port))
	errSSL := http.ListenAndServeTLS(bind_addr+":"+strconv.Itoa(ssl_port), certFile, keyFile, handlers.CompressHandler(router))

	if errSSL != nil {
		LogError(errSSL)
		os.Exit(5)
	}
}

// Runs flat HTTP server listener
// portOption - Port to listen
// bindAddr - Bind address
// router - Mux router
func runHTTPServer(portOption string, bindAddr string, router *mux.Router) {
	bind_addr := bindAddr

	// Setup HTTP server
	var tcp_port int
	tcp_port = 80
	customTCPPort := portOption
	if customTCPPort != "" {
		tcpp, e := strconv.Atoi(customTCPPort)
		if e == nil {
			tcp_port = tcpp
		}
	}

	// Listen
	LogInfo("[HTTP] Listening on " + bind_addr + ":" + strconv.Itoa(tcp_port))
	errHTTP := http.ListenAndServe(bind_addr+":"+strconv.Itoa(tcp_port), handlers.CompressHandler(router))

	if errHTTP != nil {
		LogError(errHTTP)
		os.Exit(5)
	}
}

// Handles favicon request
// response - HTTP response handler
// request - HTTP request handler
func handleFavicon(response http.ResponseWriter, request *http.Request) {
	file := filepath.Join(GetVault().path, "favicon.ico")

	if _, err := os.Stat(file); err == nil {
		http.ServeFile(response, request, file)
	} else if errors.Is(err, os.ErrNotExist) {
		frontend_path := os.Getenv("FRONTEND_PATH")

		if frontend_path == "" {
			frontend_path = "../frontend/dist/"
		}

		http.ServeFile(response, request, filepath.Join(frontend_path, "favicon.ico"))
	} else {
		response.WriteHeader(500)
		response.Write([]byte("Internal server error. Check logs for details.")) //nolint:errcheck
		LogError(err)
	}
}

var ALLOWED_IMAGES_OVERRIDE = map[string]bool{
	"android-chrome-192x192.png":          true,
	"android-chrome-512x512.png":          true,
	"android-chrome-maskable-192x192.png": true,
	"android-chrome-maskable-512x512.png": true,
	"apple-touch-icon.png":                true,
	"apple-touch-icon-60x60.png":          true,
	"apple-touch-icon-76x76.png":          true,
	"apple-touch-icon-120x120.png":        true,
	"apple-touch-icon-152x152.png":        true,
	"apple-touch-icon-180x180.png":        true,
	"favicon.png":                         true,
	"favicon.svg":                         true,
	"favicon-16x16.png":                   true,
	"favicon-32x32.png":                   true,
	"msapplication-icon-144x144.png":      true,
	"mstile-150x150.png":                  true,
}

// Handles image file request
// response - HTTP response handler
// request - HTTP request handler
func handleImageFile(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	fileName := vars["file"]

	if !ALLOWED_IMAGES_OVERRIDE[fileName] {
		response.WriteHeader(404)
		return
	}

	file := filepath.Join(GetVault().path, "img", "icons", fileName)

	if _, err := os.Stat(file); err == nil {
		http.ServeFile(response, request, file)
	} else if errors.Is(err, os.ErrNotExist) {
		frontend_path := os.Getenv("FRONTEND_PATH")

		if frontend_path == "" {
			frontend_path = "../frontend/dist/"
		}

		http.ServeFile(response, request, filepath.Join(frontend_path, "img", "icons", fileName))
	} else {
		response.WriteHeader(500)
		response.Write([]byte("Internal server error. Check logs for details.")) //nolint:errcheck
		LogError(err)
	}
}
