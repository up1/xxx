package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"time"

	// "encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"github.com/urfave/negroni"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	AuthMethod  int
}

type Routes []Route

var routes = Routes{
	{
		"Favorites Shop",
		"POST",
		"/favorite_shop",
		FavoriteShop,
		1,
	},
}

func handleAuthenMethod(handler http.Handler, r Route) http.Handler {
	// session/authorize logic goes here
	switch r.AuthMethod {
	case 0:
		// allow all
	case 1:
		// check session
		handler = VerifyEmptySession(handler)
	case 2:
		// check session and verify kplus customer
		handler = VerifyKPlusCustomerSession(handler)
	}

	return handler
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// v1
	for _, route := range routes {
		log.Infof("AuthMethod [%d]: %s\t%s\t%s", route.AuthMethod, route.Name, route.Method, route.Pattern)

		handler := handleAuthenMethod(route.HandlerFunc, route)
		router.
			PathPrefix("/v1").
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func requestMiddleware() negroni.HandlerFunc {
	return negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		start := time.Now()
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Errorln("[request middleware] unable to dump request:", err)
			JSONResponse(w, http.StatusOK, EM.General)
			return
		}

		ts := uuid.NewV4().String()
		log.Infof("[request middleware][%s] request: %s", ts, dump)

		c := httptest.NewRecorder()
		for k, v := range w.Header() {
			c.Header()[k] = v
		}
		next(c, r)

		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)

		// skip write log if response is image or terms
		if !skipLog(r.RequestURI) {
			log.Infof("[request middleware][%s] response: %s", ts, c.Body.String())
		}
		c.Body.WriteTo(w)

		res := w.(negroni.ResponseWriter)
		log.Infof(
			"[request middleware][%s] Completed %s %s %v %s in %v",
			ts,
			r.Method,
			r.URL.Path,
			res.Status(),
			http.StatusText(res.Status()),
			time.Since(start),
		)
	})
}

func skipLog(uri string) bool {
	check := []string{"/image", "/terms", ".jpeg", ".jpg", ".png"}
	for _, c := range check {
		if strings.Contains(uri, c) {
			return true
		}
	}
	return false
}

func VerifyEmptySession(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := GetSessionDataRedis(r, "lfsession")
		if sess["MobileNo"] == nil {
			JSONResponse(w, http.StatusBadRequest, EM.Mismatch.RedisData)
			return
		}
		log.Debugln("[verify session] verify session success:", sess)
		inner.ServeHTTP(w, r)
	})
}

func VerifyKPlusCustomerSession(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := GetSessionDataRedis(r, "lfsession")
		if sess["MobileNo"] == nil {
			JSONResponse(w, http.StatusBadRequest, EM.Mismatch.RedisData)
			return
		}
		if sess["MobileNo"].(string) == "" || strings.EqualFold(sess["MobileNo"].(string), "unknown") {
			JSONResponse(w, http.StatusBadRequest, EM.Timeout.Session)
			return
		}
		log.Debugln("[VerifyKPlusCustomerSession session] verify session success:", sess)
		inner.ServeHTTP(w, r)
	})
}
