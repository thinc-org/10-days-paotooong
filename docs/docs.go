package docs

import (
	_ "embed"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

//go:embed docs.html
var static []byte

func GetDocHandler() runtime.HandlerFunc {
	fs := func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
		w.Write(static)
	}
	return fs
}
