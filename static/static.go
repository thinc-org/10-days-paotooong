package static

import (
	"embed"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

//go:embed *
var static embed.FS

func GetDocHandler() runtime.HandlerFunc {
	fs := http.FileServer(http.FS(static))
	return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/static")
		fs.ServeHTTP(w, r)
	}
}
