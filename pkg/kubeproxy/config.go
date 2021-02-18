package kubeproxy

import (
	"net/http"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesTransportAndServerURL(configPath string) (transport http.RoundTripper, server *url.URL, err error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return
	}

	server, err = server.Parse(config.Host)
	if err != nil {
		return
	}

	transport, err = rest.TransportFor(config)
	if err != nil {
		return
	}

	return
}
