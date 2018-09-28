package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

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
	ch := make(chan bool)

	setupHandler()
	go setupListener()
	go watch(*config, ch)

	Info("Watching %v for changes", *config)

	for {
		<-ch
		Info("Reacting to changes in %v", *config)
		setupHandler()
		Info("Applied updates to %v", *config)
	}
}

func watch(fileName string, ch chan bool) {
	curr, err := os.Stat(fileName)

	if err != nil {
		Warn("Error getting stats for %v: %v", *config, err)
		return
	}

	for {
		next, err := os.Stat(fileName)

		if err != nil {
			Warn("Error getting stats for %v: %v", *config, err)
		} else if next.ModTime() != curr.ModTime() {
			curr = next
			ch <- true
			time.Sleep(time.Second * 15)
		} else {
			time.Sleep(time.Second * 60)
		}
	}
}

func setupHandler() {
	Info("Reading configuration from %v", *config)
	contents, err := ioutil.ReadFile(*config)

	if err != nil {
		Warn("Error reading Servfile: %v", err)
		return
	}

	decls, matches := Parse(string(contents))
	servers, _ := Runtime(decls, matches)

	supervisor := http.NewServeMux()
	http.DefaultServeMux = supervisor

	supervisor.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handled := false

		for i, server := range servers {
			Info("Comparing request to server #%d", i+1)

			if server.Match(*r) {
				server.Mux.ServeHTTP(w, r)
				handled = true
				break
			}
		}

		if !handled {
			Warn("No matches found")
		}
	})
}

func setupListener() {
	Info("Reading configuration from %v", *config)
	contents, err := ioutil.ReadFile(*config)

	if err != nil {
		Fatal("Error reading Servfile: %v", err)
	}

	decls, matches := Parse(string(contents))
	_, env := Runtime(decls, matches)

	if cache, ok := env.GetValue("cache"); ok && *certCache == "" {
		*certCache = cache.Value()
	}

	if domains, ok := env.GetValue("domains"); ok {
		certDomains = append(certDomains, domains.Values()...)
	}

	if *listen == "" {
		for _, domain := range certDomains {
			Info("Whitelisting %s", domain)
		}

		m := &autocert.Manager{
			Cache:      autocert.DirCache(*certCache),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(certDomains...),
		}

		go func() {
			Fatal("%s", http.ListenAndServe(":http", m.HTTPHandler(nil)))
		}()

		s := &http.Server{
			Addr:      ":https",
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}

		s.ListenAndServeTLS("", "")
	} else {
		Info("Starting http server on %v", *listen)
		Fatal("%s", http.ListenAndServe(*listen, nil))
	}
}
