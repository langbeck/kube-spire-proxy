package kubeproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func HandlerFromKubeconfig(filename string) (http.Handler, error) {
	transport, serverURL, err := getKubernetesTransportAndServerURL(filename)
	if err != nil {
		return nil, err
	}

	return HandlerWithRawConfig(serverURL, transport), nil
}

func HandlerWithRawConfig(target *url.URL, transport http.RoundTripper) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.FlushInterval = -1

	return &authHandler{next: proxy}
}
