package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"strings"
)

func listenAndServe() error {
	l := fmt.Sprintf("%s:%d", listenHost, listenPortHTTP)
	srv := &http.Server{Addr: l, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectToCanonicalHost(w, r)
		return
	})}
	return srv.ListenAndServe()
}

func listenAndServeTLS() error {
	l := fmt.Sprintf("%s:%d", listenHost, listenPortHTTPS)
	srv := &http.Server{Addr: l, Handler: newServeMux()}
	if tlsKeyFile == "" && tlsCertFile == "" {
		hosts := append([]string{listenHost}, fmt.Sprintf("www.%s", listenHost))
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(hosts...),
			Cache:      autocert.DirCache("certs"),
		}
		srv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}
	}
	return srv.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
}

func newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/new",
		ensureCanonicalHost(
			acceptMethod("POST",
				requireToken(newHandler))))

	mux.Handle("/", ensureCanonicalHost(indexHandler))
	return mux
}

// ensureCanonicalHost decorates a http.Handler, ensuring that the
// request being served is on the canonical hostname.
func ensureCanonicalHost(h http.Handler) http.Handler {
	listenHostPort := fmt.Sprintf("%s:%d", listenHost, listenPortHTTPS)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != listenHost && r.Host != listenHostPort {
			log.Println(r.Host, "doesnt equal", listenHost, "or", listenPortHTTPS)
			redirectToCanonicalHost(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// acceptMethod decorates a http.Handler, ensuring that the given HTTP method
// is used in the request.
func acceptMethod(m string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func verifyToken(s string) bool {
	sv := fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
	return subtle.ConstantTimeCompare([]byte(sv), []byte(tokenSHA256)) == 1
}

// requireToken decorates a http.Handler, ensuring that the request has a valid
// token identifier.
func requireToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !verifyToken(r.PostFormValue("token")) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// redirectToCanonicalHost responds with a redirect to the canonical host.
func redirectToCanonicalHost(w http.ResponseWriter, r *http.Request) {
	t := canonicalHost + r.URL.Path
	http.Redirect(w, r, t, http.StatusMovedPermanently)
	if _, err := fmt.Fprintf(w, `<a href="%s">Redirecting...</a>`, t); err != nil {
		log.Println("error writing response:", err.Error())
	}
}

// newHandler is an http.Handler that creates a new item in the urlStore store.
var newHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	link := r.PostFormValue("url")
	if len(link) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	k, err := urlStore.newItemID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	urlStore.mu.Lock()
	urlStore.entries[k] = newItem(link)
	urlStore.mu.Unlock()
	b, err := json.Marshal(map[string]string{"short-url": canonicalHost + k.String()})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = fmt.Fprintf(w, `{"error": "%s"}\n`, strings.Replace(err.Error(), `"`, `'`, -1))
		if err != nil {
			log.Println("error writing response:", err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err = fmt.Fprintln(w, string(b)); err != nil {
		log.Println("error writing response:", err.Error())
	}
	return
})

// indexHandler is a catch-all http.Handler that attempts to lookup items in
// the store based on request path.
var indexHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimLeft(r.URL.Path, "/")
	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	k := itemID(key)
	urlStore.mu.RLock()
	v, ok := urlStore.entries[k]
	urlStore.mu.RUnlock()
	if ok {
		if urlStore.isExpired(v) {
			// rely on the reaper function to actually delete items.
			w.WriteHeader(http.StatusNotFound)
			return
		}
		u := v.String()
		http.Redirect(w, r, u, http.StatusMovedPermanently)
		if _, err := fmt.Fprintf(w, `<a href="%s">Redirecting...</a>`, u); err != nil {
			log.Println("error writing response:", err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
	return
})