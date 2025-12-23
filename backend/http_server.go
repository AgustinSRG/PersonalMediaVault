// HTTP Server

// cSpell:ignore Subrouter, msapplication, mstile

package main

import (
	"crypto/tls"
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	CORS_INSECURE_MODE_ENABLED = false // Insecure CORS mode for development and testing
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Logging middleware to log requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mark stating time
		startTime := time.Now()

		lrw := loggingResponseWriter{w, http.StatusOK}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(&lrw, r)

		// Log request
		LogRequest(r, lrw.statusCode, time.Since(startTime))
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

	// Assets (get encrypted media files)

	router.HandleFunc("/assets/b/{mid:[0-9]+}/{asset:[0-9]+}/{filename}", api_handleAssetGet).Methods("GET", "HEAD")
	router.HandleFunc("/assets/p/{mid:[0-9]+}/{asset:[0-9]+}/{filename}", api_handleAssetVideoPreviews).Methods("GET")

	router.HandleFunc("/assets/album_thumb/{asset:[0-9]+}/{filename}", api_handleAlbumThumbnailAssetGet).Methods("GET", "HEAD")

	// API routes

	apiRouter := router.PathPrefix("/api/").Subrouter()
	apiRouter.Use(handlers.CompressHandler)

	// Auth API
	apiRouter.HandleFunc("/auth/login", api_handleAuthLogin).Methods("POST")
	apiRouter.HandleFunc("/auth/logout", api_handleAuthLogout).Methods("POST")

	// Account API (changing credentials)
	apiRouter.HandleFunc("/account", api_getAccountContext).Methods("GET")

	apiRouter.HandleFunc("/account/username", api_getAccountContext).Methods("GET")
	apiRouter.HandleFunc("/account/username", api_changeUsername).Methods("POST")

	apiRouter.HandleFunc("/account/password", api_changePassword).Methods("POST")

	apiRouter.HandleFunc("/account/security", api_getSecurityOptions).Methods("GET")
	apiRouter.HandleFunc("/account/security", api_setSecurityOptions).Methods("POST")

	apiRouter.HandleFunc("/account/security/tfa/totp", api_getParametersTwoFactorAuthTimeOtp).Methods("GET")
	apiRouter.HandleFunc("/account/security/tfa/totp", api_enableTwoFactorAuthTimeOtp).Methods("POST")

	apiRouter.HandleFunc("/account/security/tfa/disable", api_disableTwoFactorAuth).Methods("POST")

	// Invite codes API
	apiRouter.HandleFunc("/invites", api_getInviteCodeStatus).Methods("GET")
	apiRouter.HandleFunc("/invites/sessions", api_getInviteCodeSessions).Methods("GET")
	apiRouter.HandleFunc("/invites/login", api_loginWithInviteCode).Methods("POST")
	apiRouter.HandleFunc("/invites/generate", api_generateInviteCode).Methods("POST")
	apiRouter.HandleFunc("/invites/clear", api_clearInviteCode).Methods("POST")
	apiRouter.HandleFunc("/invites/sessions/{index}", api_closeInviteSession).Methods("DELETE")

	// Admin API (Manage accounts)
	apiRouter.HandleFunc("/admin/accounts", api_getAccounts).Methods("GET")
	apiRouter.HandleFunc("/admin/accounts", api_createAccount).Methods("POST")
	apiRouter.HandleFunc("/admin/accounts/delete", api_deleteAccount).Methods("POST")
	apiRouter.HandleFunc("/admin/accounts/update", api_updateAccount).Methods("POST")
	apiRouter.HandleFunc("/admin/launcher/{tag}", api_checkLauncherTag).Methods("GET")

	// Media API
	apiRouter.HandleFunc("/media/{mid:[0-9]+}", api_getMedia).Methods("GET")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/albums", api_getMediaAlbums).Methods("GET")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/size_stats", api_getMediaSizeStats).Methods("GET")

	apiRouter.HandleFunc("/upload", api_uploadMedia).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/title", api_editMediaTitle).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/extra", api_editMediaExtraParams).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/thumbnail", api_editMediaThumbnail).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/notes", api_setImageNotes).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/description", api_setDescription).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/time_slices", api_editMediaTimelineSplices).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/edit/related", api_editMediaRelated).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/delete", api_deleteMedia).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/encode", api_mediaRequestEncode).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/replace", api_replaceMedia).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/resolution/add", api_mediaAddResolution).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/resolution/remove", api_mediaRemoveResolution).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/subtitles/set", api_addMediaSubtitles).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/subtitles/remove", api_removeMediaSubtitles).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/subtitles/rename", api_renameMediaSubtitles).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/audios/set", api_addMediaAudioTrack).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/audios/remove", api_removeMediaAudioTrack).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/audios/rename", api_renameMediaAudioTrack).Methods("POST")

	apiRouter.HandleFunc("/media/{mid:[0-9]+}/attachments/add", api_addMediaAttachment).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/attachments/rename", api_updateMediaAttachment).Methods("POST")
	apiRouter.HandleFunc("/media/{mid:[0-9]+}/attachments/remove", api_removeMediaAttachment).Methods("POST")

	// Search API
	apiRouter.HandleFunc("/search", api_searchMedia).Methods("GET")
	apiRouter.HandleFunc("/search/advanced", api_advancedSearch).Methods("GET")
	apiRouter.HandleFunc("/random", api_randomMedia).Methods("GET")

	// Tags API
	apiRouter.HandleFunc("/tags", api_getTags).Methods("GET")

	apiRouter.HandleFunc("/tags/add", api_tagMedia).Methods("POST")
	apiRouter.HandleFunc("/tags/remove", api_untagMedia).Methods("POST")

	// Albums API
	apiRouter.HandleFunc("/albums", api_getAlbums).Methods("GET")
	apiRouter.HandleFunc("/albums/{id:[0-9]+}", api_getAlbum).Methods("GET")

	apiRouter.HandleFunc("/albums", api_createAlbum).Methods("POST")

	apiRouter.HandleFunc("/albums/{id:[0-9]+}/delete", api_deleteAlbum).Methods("POST")

	apiRouter.HandleFunc("/albums/{id:[0-9]+}/rename", api_renameAlbum).Methods("POST")

	apiRouter.HandleFunc("/albums/{id:[0-9]+}/set", api_setAlbumList).Methods("POST")

	apiRouter.HandleFunc("/albums/{id:[0-9]+}/add", api_albumAddMedia).Methods("POST")
	apiRouter.HandleFunc("/albums/{id:[0-9]+}/remove", api_albumRemoveMedia).Methods("POST")
	apiRouter.HandleFunc("/albums/{id:[0-9]+}/move", api_albumMoveMedia).Methods("POST")

	apiRouter.HandleFunc("/albums/{id:[0-9]+}/thumbnail", api_editAlbumThumbnail).Methods("POST")

	// Config API
	apiRouter.HandleFunc("/config", api_getConfig).Methods("GET")
	apiRouter.HandleFunc("/config", api_setConfig).Methods("POST")

	// Home page API
	apiRouter.HandleFunc("/home", api_listHomePageGroups).Methods("GET")
	apiRouter.HandleFunc("/home", api_createHomePageGroup).Methods("POST")
	apiRouter.HandleFunc("/home/{id:[0-9]+}/elements", api_getHomePageGroupElements).Methods("GET")
	apiRouter.HandleFunc("/home/{id:[0-9]+}/name", api_renameHomePageGroup).Methods("POST")
	apiRouter.HandleFunc("/home/{id:[0-9]+}/move", api_moveHomePageGroup).Methods("POST")
	apiRouter.HandleFunc("/home/{id:[0-9]+}/elements", api_setElementsHomePageGroup).Methods("POST")
	apiRouter.HandleFunc("/home/{id:[0-9]+}", api_deleteHomePageGroup).Methods("DELETE")

	// Tasks API
	apiRouter.HandleFunc("/tasks", api_getTasks).Methods("GET")
	apiRouter.HandleFunc("/tasks/{id:[0-9]+}", api_getTask).Methods("GET")

	// About API
	apiRouter.HandleFunc("/about", api_about).Methods("GET")
	apiRouter.HandleFunc("/about/disk_usage", api_getDiskUsage).Methods("GET")

	// Semantic search API
	apiRouter.HandleFunc("/search/semantic", api_searchMediaSemantic).Methods("POST")

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

	router.Path("/favicon.ico").Methods("GET").Handler(handlers.CompressHandler(cacheTTLAdd(http.HandlerFunc(handleFavicon))))
	router.Path("/img/icons/{file}").Methods("GET").Handler(handlers.CompressHandler(cacheTTLAdd(http.HandlerFunc(handleImageFile))))
	router.PathPrefix("/").Handler(handlers.CompressHandler(cacheTTLAdd(http.FileServer(http.Dir(frontend_path)))))

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
		p, e := strconv.Atoi(customSSLPort)
		if e == nil {
			ssl_port = p
		}
	}

	// Check key pair
	_, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		LogError(err)
		os.Exit(6)
	}

	// Listen
	LogInfo("[SSL] Listening on " + bind_addr + ":" + strconv.Itoa(ssl_port))
	errSSL := http.ListenAndServeTLS(bind_addr+":"+strconv.Itoa(ssl_port), certFile, keyFile, router)

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
		p, e := strconv.Atoi(customTCPPort)
		if e == nil {
			tcp_port = p
		}
	}

	// Listen
	LogInfo("[HTTP] Listening on " + bind_addr + ":" + strconv.Itoa(tcp_port))
	errHTTP := http.ListenAndServe(bind_addr+":"+strconv.Itoa(tcp_port), router)

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
