package tlsinfo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"regexp"
	"sync/atomic"

	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
)

type contextKey string

const connContextKey = contextKey("connInfo")

var nextConnID uint32

var (
	ErrNoUser     = errors.New("no user information")
	ErrNoPeerCert = errors.New("no peer certificates")
)

var spiffeIDPathPattern = regexp.MustCompile("^/k8s-user/([^/]+)/([^/]+)$")

func CreateContext(ctx context.Context, c net.Conn) context.Context {
	tlsConn, ok := c.(*tls.Conn)
	if ok {
		return context.WithValue(ctx, connContextKey, resolveUser(tlsConn))
	}

	return nil
}

func GetUserInfo(ctx context.Context) (*UserInfo, error) {
	info, ok := ctx.Value(connContextKey).(*UserInfo)
	if !ok {
		return nil, ErrNoUser
	}

	if info.err != nil {
		return nil, info.err
	}

	atomic.AddUint32(&info.used, 1)
	return info, nil
}

type UserInfo struct {
	// Debug information
	ID   uint32
	used uint32

	// User information
	User   string
	Groups []string

	// Persisted error
	err error
}

func (userInfo *UserInfo) Used() uint32 {
	return atomic.LoadUint32(&userInfo.used)
}

func resolveUser(tlsConn *tls.Conn) (info *UserInfo) {
	info = new(UserInfo)

	err := tlsConn.Handshake()
	if err != nil {
		info.err = err
		return
	}

	peerCertificates := tlsConn.ConnectionState().PeerCertificates
	if len(peerCertificates) == 0 {
		info.err = ErrNoPeerCert
		return
	}

	spiffeID, err := x509svid.IDFromCert(peerCertificates[0])
	if err != nil {
		info.err = fmt.Errorf("invalid peer certificate: %w", err)
		return
	}

	matchGroups := spiffeIDPathPattern.FindStringSubmatch(spiffeID.Path())
	if len(matchGroups) == 0 {
		info.err = fmt.Errorf("could not extract user info from spiffeID: %s", spiffeID)
		return info
	}

	info.User = matchGroups[1]
	info.Groups = []string{matchGroups[2]}
	info.ID = atomic.AddUint32(&nextConnID, 1)
	return info
}
