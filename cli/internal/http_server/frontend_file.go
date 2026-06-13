//go:build e2e_file_frontend

package http_server

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const frontendDirEnv = "CHATTO_FRONTEND_DIR"

func frontendAssetFS() (fs.FS, error) {
	dir := os.Getenv(frontendDirEnv)
	if dir == "" {
		return nil, fmt.Errorf("%s must point at a built frontend directory", frontendDirEnv)
	}

	if _, err := os.Stat(filepath.Join(dir, "200.html")); err != nil {
		return nil, fmt.Errorf("%s must point at a built frontend directory containing 200.html: %w", frontendDirEnv, err)
	}

	return os.DirFS(dir), nil
}
