package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/minond/serv/serv"
	"golang.org/x/crypto/acme/autocert"
)

type stringListFlag []string

var (
	certDomains stringListFlag
	certCache   = flag.String("certCache", "", "Path to Let's Encrypt cache file. Use this along with the cache definition.")
	listen      = flag.String("listen", "", "Host and port to listen on.")
	config      = flag.String("config", "./Servfile", "Path to Servfile file.")
)

func (l *stringListFlag) String() string {
	return strings.Join(*l, ", ")
}

func (l *stringListFlag) Set(val string) error {
	*l = append(*l, val)
	return nil
}

func init() {
	flag.Var(&certDomains, "certDomain", "Domain(s) whitelist. Use this along with the domains definition.")
	flag.Parse()
}

func main() {
	serv.Info("Reading configuration from %v", *config)
	contents, err := ioutil.ReadFile(*config)

	if err != nil {
		serv.Fatal("Error reading Servfile: %v", err)
	}

	decls, matches := serv.Parse(string(contents))
	servers, env := serv.Runtime(decls, matches)

	if cache, ok := env.GetValue("cache"); ok && *certCache == "" {
		*certCache = cache.Value()
	}

	if domains, ok := env.GetValue("domains"); ok {
		certDomains = append(certDomains, domains.Values()...)
	}

	supervisor := http.NewServeMux()
	http.DefaultServeMux = supervisor

	supervisor.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handled := false

		for i, server := range servers {
			serv.Info("Comparing request to server #%d", i+1)

			if server.Match(*r) {
				server.Mux.ServeHTTP(w, r)
				handled = true
				break
			}
		}

		if !handled {
			serv.Warn("No matches found")
		}
	})

	if *listen == "" {
		for _, domain := range certDomains {
			serv.Info("Whitelisting %s", domain)
		}

		m := &autocert.Manager{
			Cache:      autocert.DirCache(*certCache),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(certDomains...),
		}

		go func() {
			serv.Fatal("%s", http.ListenAndServe(":http", m.HTTPHandler(nil)))
		}()

		s := &http.Server{
			Addr:      ":https",
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}

		s.ListenAndServeTLS("", "")
	} else {
		serv.Info("Starting http server on %v", *listen)
		serv.Fatal("%s", http.ListenAndServe(*listen, supervisor))
	}
}
