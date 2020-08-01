package main_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) (string, func()) {
	tf, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	err = tf.Close()
	require.NoError(t, err)

	fileName := tf.Name()

	cmd := exec.Command("go", "build", "-o", fileName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err)

	return fileName, func() {
		err := os.Remove(fileName)
		require.NoError(t, err)
	}
}

func startBinary(t *testing.T, c string, args ...string) func() {
	cmd := exec.Command(c, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	require.NoError(t, err)
	fmt.Println("started: ", c, strings.Join(args, " "))

	return func() {
		err = cmd.Process.Kill()
		require.NoError(t, err)
		_, err = cmd.Process.Wait()
		require.NoError(t, err)
		// require.Equal(t, 0, ps.ExitCode())
	}
}

func waitForPortToOpen(t *testing.T, addr string) {
	time.Sleep(1 * time.Second)
}

type receivedRequest struct {
	verb   string
	path   string
	body   []byte
	header http.Header
}
type mockBackend struct {
	Addr             string
	receivedRequests []receivedRequest
	mu               *sync.Mutex
}

func (m mockBackend) url() string {
	return fmt.Sprintf("http://%s", m.Addr)
}

func (m mockBackend) getReceivedRequests() []receivedRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	rrs := make([]receivedRequest, len(m.receivedRequests))
	for i, rr := range m.receivedRequests {
		rrs[i] = rr
	}

	return rrs
}

func startMockBackend(t *testing.T) (*mockBackend, func()) {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	mu := new(sync.Mutex)

	mb := &mockBackend{
		Addr: l.Addr().String(),
		mu:   mu,
	}

	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bod, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("while reading body: %s", err.Error()), 500)
			}

			h := http.Header{}
			for k, v := range r.Header {
				vs := make([]string, len(v))
				for i, vv := range v {
					vs[i] = vv
				}
				h[k] = vs
			}

			rr := receivedRequest{
				verb:   r.Method,
				path:   r.URL.Path,
				body:   bod,
				header: h,
			}

			mu.Lock()
			mb.receivedRequests = append(mb.receivedRequests, rr)
			mu.Unlock()
		}),
	}

	go s.Serve(l)

	return mb, func() {
		s.Shutdown(context.Background())
	}

}

func TestProxy(t *testing.T) {
	binName, cleanupBinary := buildBinary(t)
	defer cleanupBinary()
	require.NotNil(t, binName)

	mainBackend, shutdownMainBackend := startMockBackend(t)
	defer shutdownMainBackend()

	shutdownProxy := startBinary(t, binName, "--port", "23533", mainBackend.url())
	defer shutdownProxy()

	waitForPortToOpen(t, ":23533")

	t.Run("proxy to main backend", func(t *testing.T) {
		res, err := http.Get(fmt.Sprintf("http://localhost:23533"))
		require.NoError(t, err)
		require.Equal(t, 200, res.StatusCode)
		require.Equal(t, 1, len(mainBackend.getReceivedRequests()))
	})

}
