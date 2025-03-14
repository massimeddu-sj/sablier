package traefik

import (
	"context"
	"io"
	"net/http"
)

type SablierMiddleware struct {
	client  *http.Client
	request *http.Request
	next    http.Handler
}

// New function creates the configuration
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	req, err := config.BuildRequest(name)

	if err != nil {
		return nil, err
	}

	return &SablierMiddleware{
		request: req,
		client:  &http.Client{},
		next:    next,
	}, nil
}

func (sm *SablierMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	sablierRequest := sm.request.Clone(context.TODO())

	resp, err := sm.client.Do(sablierRequest)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Sablier-Session-Status") == "ready" {
		sm.next.ServeHTTP(rw, req)
	} else {
		forward(resp, rw)
	}
}

func forward(resp *http.Response, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(rw, resp.Body)
}
