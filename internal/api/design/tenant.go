package design

import (
    . "goa.design/goa/v3/dsl"
)

// TenantService defines the tenant management service
var _ = Service("tenant", func() {
    Description("Tenant management service")
    
    // Create tenant endpoint
    Method("create", func() {
        Description("Create a new tenant")
        
        Payload(func() {
            Attribute("name", String, "Tenant name", func() {
                Meta("rpc:tag", "1")
            })
            Attribute("subdomain", String, "Unique subdomain", func() {
                Meta("rpc:tag", "2")
            })
            Attribute("plan_type", String, "Subscription plan", func() {
                Meta("rpc:tag", "3")
            })
            
            Required("name", "subdomain")
        })
        
        Result(TenantResult)
        
        Error("bad_request", String, "Invalid request")
        Error("conflict", String, "Subdomain already exists")
        Error("internal_error", String, "Internal server error")
        
        HTTP(func() {
            POST("/tenants")
            Response(StatusCreated)
            Response("bad_request", StatusBadRequest)
            Response("conflict", StatusConflict)
            Response("internal_error", StatusInternalServerError)
        })
        
        GRPC(func() {
            Response(CodeOK)
            Response("bad_request", CodeInvalidArgument)
            Response("conflict", CodeAlreadyExists)
            Response("internal_error", CodeInternal)
        })
    })
    
    // Get tenant endpoint
    Method("get", func() {
        Description("Get tenant by ID")
        
        Payload(func() {
            Attribute("id", String, "Tenant ID", func() {
                Meta("rpc:tag", "1")
            })
            Token("token", String, "JWT token")
            Required("id")
        })
        
        Result(TenantResult)
        
        Error("not_found", String, "Tenant not found")
        Error("internal_error", String, "Internal server error")
        
        Security(JWTAuth)
        
        HTTP(func() {
            GET("/tenants/{id}")
            Response(StatusOK)
            Response("not_found", StatusNotFound)
            Response("internal_error", StatusInternalServerError)
        })
        
        GRPC(func() {
            Response(CodeOK)
            Response("not_found", CodeNotFound)
            Response("internal_error", CodeInternal)
        })
    })
    
    // List tenants endpoint
    Method("list", func() {
        Description("List tenants with pagination")
        
        Payload(func() {
            Attribute("offset", Int, "Pagination offset", func() {
                Meta("rpc:tag", "1")
            })
            Attribute("limit", Int, "Pagination limit", func() {
                Meta("rpc:tag", "2")
            })
            Token("token", String, "JWT token")
        })
        
        Result(CollectionOf(TenantResult, func() {
            View("minimal")
        }))
        
        Security(JWTAuth, func() {
            Scope("admin")
        })
        
        HTTP(func() {
            GET("/tenants")
            Param("offset")
            Param("limit")
            Response(StatusOK)
        })
        
        GRPC(func() {
            Response(CodeOK)
        })
    })
})
