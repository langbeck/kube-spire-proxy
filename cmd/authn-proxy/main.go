package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/langbeck/kube-spire-proxy/pkg/kubeproxy"
	"github.com/langbeck/kube-spire-proxy/pkg/tlsinfo"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

const (
	defaultBindAddress = "0.0.0.0:5443"
	defaultKubeConfig  = "/etc/kubernetes/admin.conf"
)

func run(ctx context.Context) error {
	bindAddress := flag.String("bindAddress", defaultBindAddress, "Address and port to listen at")
	kubeConfig := flag.String("kubeconfig", defaultKubeConfig, "Configuration file with client credentials and server information")

	handler, err := kubeproxy.HandlerFromKubeconfig(*kubeConfig)
	if err != nil {
		return err
	}

	return listenAndServe(ctx, *bindAddress, handler)
}

func listenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return err
	}

	defer source.Close()

	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeAny())
	server := http.Server{
		Addr:        addr,
		TLSConfig:   tlsConfig,
		ConnContext: tlsinfo.CreateContext,
		Handler:     handler,
	}

	return server.ListenAndServeTLS("", "")
}

func main() {
	err := run(context.Background())
	if err != nil {
		log.Printf("[ERROR] Server failure: %v", err)
		return
	}
}
