package main

import (
	"net/http"

	"github.com/niiniyare/erp/internal/platform/admin"
)

// MountAdminUI adds the admin UI routes to the muxer
func MountAdminUI(mux http.Handler) http.Handler {
	return admin.Mount(mux)
}