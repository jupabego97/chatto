//go:build !e2e_file_frontend

package http_server

import (
	"embed"
	"io/fs"
)

//go:embed all:.client
var embeddedWebUIFS embed.FS

func frontendAssetFS() (fs.FS, error) {
	return fs.Sub(embeddedWebUIFS, ".client")
}
