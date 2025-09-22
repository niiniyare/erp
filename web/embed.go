package web

import (
	"embed"
)

// AdminUI contains all the admin interface files
//go:embed admin/*.html admin/js admin/config admin/pages admin/lib
var AdminUI embed.FS