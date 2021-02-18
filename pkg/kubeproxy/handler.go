package kubeproxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/langbeck/kube-spire-proxy/pkg/tlsinfo"
)

type authHandler struct {
	next http.Handler
}

func (handler *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userInfo, err := tlsinfo.GetUserInfo(r.Context())
	if err != nil {
		w.Header().Set("Cache-Control", "no-cache, private")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, err)
		return
	}

	logRequest(r, userInfo)

	// Add impersonation headers
	r.Header["X-Remote-User"] = []string{userInfo.User}
	r.Header["X-Remote-Group"] = userInfo.Groups

	// Forward to the next handler
	handler.next.ServeHTTP(w, r)
}

func logRequest(r *http.Request, userInfo *tlsinfo.UserInfo) {
	requestID := fmt.Sprintf("%3d.%d", userInfo.ID, userInfo.Used())
	method := r.Method
	if method == "GET" {
		if r.URL.Query().Get("watch") == "true" {
			method = "WATCH"
		}
	}

	log.Printf("%-8s %-20s %-5s %-90s", requestID, userInfo.User, method, r.URL.Path)
}
