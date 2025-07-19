package design

import (
	. "goa.design/goa/v3/dsl"
)

// OpenAPIService provides access to the generated OpenAPI specification
var _ = Service("openapi", func() {
	Description("OpenAPI specification service")

	HTTP(func() {
		Path("/")
	})

	// Serve OpenAPI JSON specification
	Method("spec", func() {
		Description("Get OpenAPI specification in JSON format")

		Result(func() {
			Attribute("spec", Any, "OpenAPI specification", func() {
				Example(map[string]interface{}{
					"openapi": "3.0.3",
					"info": map[string]interface{}{
						"title":   "Enterprise AWO ERP System API",
						"version": "1.0.0",
					},
				})
			})
			Required("spec")
		})

		Error("internal_error")

		HTTP(func() {
			GET("/openapi.json")
			Response(StatusOK)
			Response("internal_error", StatusInternalServerError)
		})

		// No security required for OpenAPI spec
		NoSecurity()
	})

	// Serve Swagger UI (optional)
	Method("ui", func() {
		Description("Redirect to Swagger UI")

		Result(func() {
			Attribute("location", String, "Redirect location", func() {
				Example("/swagger-ui/")
			})
			Required("location")
		})

		HTTP(func() {
			GET("/swagger-ui")
			Response(StatusMovedPermanently, func() {
				Header("location:Location", String, "Redirect location", func() {
					Example("/swagger-ui/")
				})
			})
		})

		NoSecurity()
	})
})
