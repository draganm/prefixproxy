package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func main() {

	app := cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Value:   5555,
				EnvVars: []string{"PORT"},
			},
		},

		Action: func(c *cli.Context) error {

			mux := http.NewServeMux()

			for _, proxy := range c.Args().Slice() {
				parts := strings.SplitN(proxy, "=", 2)

				if len(parts) != 2 {
					return errors.Errorf("malformed proxy %q", proxy)
				}

				path := parts[0]
				urlString := parts[1]
				url, err := url.Parse(urlString)

				if err != nil {
					return errors.Wrapf(err, "while parsing URL %q", urlString)
				}

				proxy := httputil.NewSingleHostReverseProxy(url)

				log.Printf("proxy: %s => %s", path, urlString)

				mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
					log.Printf("[%s]: %s %s", urlString, r.Method, r.URL.Path)
					proxy.ServeHTTP(w, r)
				})

			}

			addr := fmt.Sprintf(":%d", c.Int("port"))
			l, err := net.Listen("tcp", addr)
			if err != nil {
				return errors.Wrapf(err, "while listening on %s", addr)
			}

			s := &http.Server{
				Handler: mux,
			}

			log.Println("listening on", addr)

			return s.Serve(l)

		},
	}

	app.RunAndExitOnError()

}
