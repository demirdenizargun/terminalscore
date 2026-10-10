package espn

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// Android'de (Termux) /etc/resolv.conf yok; Go'nun kendi çözücüsü o zaman
// localhost:53'e sorup başarısız oluyor. Bu durumda Termux'un resolv.conf'unu,
// o da yoksa genel bir DNS sunucusunu kullanıyoruz.
func transport() http.RoundTripper {
	if runtime.GOOS != "linux" && runtime.GOOS != "android" {
		return http.DefaultTransport
	}
	if _, err := os.Stat("/etc/resolv.conf"); err == nil {
		return http.DefaultTransport
	}

	server := fallbackDNS()
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, server)
		},
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, Resolver: resolver}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = dialer.DialContext
	return t
}

func fallbackDNS() string {
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		if f, err := os.Open(prefix + "/etc/resolv.conf"); err == nil {
			defer f.Close()
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				fields := strings.Fields(sc.Text())
				if len(fields) == 2 && fields[0] == "nameserver" {
					return net.JoinHostPort(fields[1], "53")
				}
			}
		}
	}
	return "1.1.1.1:53"
}
