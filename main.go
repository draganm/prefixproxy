package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

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
			if c.NArg() != 1 {
				return errors.New("Main proxy URL must be provided")
			}

			mainProxyURLString := c.Args().First()

			mainProxyURL, err := url.Parse(mainProxyURLString)
			if err != nil {
				return errors.Wrapf(err, "while parsing URL %q", mainProxyURLString)
			}

			mainProxy := httputil.NewSingleHostReverseProxy(mainProxyURL)

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				log.Println("forwarding request to main")
				mainProxy.ServeHTTP(w, r)
			})

			addr := fmt.Sprintf(":%d", c.Int("port"))
			l, err := net.Listen("tcp", addr)
			if err != nil {
				return errors.Wrapf(err, "while listening on %s", addr)
			}

			s := &http.Server{
				Handler: mux,
			}

			return s.Serve(l)

		},
	}

	app.RunAndExitOnError()

}
