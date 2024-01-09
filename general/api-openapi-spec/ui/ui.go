// Package ui is just a wrapper to serve swagger ui from the API.
package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

type relativeFs struct {
	content embed.FS
}

func (c relativeFs) Open(name string) (fs.File, error) {
	return c.content.Open(path.Join("submodules/api-openapi-spec/ui/", name))
}

// FS simple wrapper what will allow you to serve ui as a mounted file system.
func FS(efs embed.FS) http.FileSystem {
	return http.FS(relativeFs{efs})
}

// File allows you to send embedded spec file through the API.
func File(dta []byte) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		gcx.Data(http.StatusOK, "text/yaml", dta)
	}
}
