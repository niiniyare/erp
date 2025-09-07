package admin

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("administrative-performance", func() {
	Description("Administrative and Performance Management Service")

	HTTP(func() {
		Path("/api/v1/admin")
		Header("X-Tenant-ID", String, "Tenant identifier", func() {
			Example("550e8400-e29b-41d4-a716-446655440000")
		})
	})

	Security("jwt")

	Error("unauthorized", String, "Unauthorized access")
	Error("forbidden", String, "Forbidden access")
	Error("not_found", String, "Resource not found")
	Error("bad_request", String, "Bad request")

	Method("get_system_health", func() {
		Description(" Health Check")
		Payload(func() {
			Attribute("include_dependencies", Boolean, "Include external dependency status", func() {
				Default(true)
			})
			Attribute("include_performance", Boolean, "Include performance metrics", func() {
				Default(true)
			})
			Attribute("include_capacity", Boolean, "Include capacity utilization", func() {
				Default(false)
			})
			Attribute("include_predictions", Boolean, "Include predictive analytics", func() {
				Default(false)
			})
			Attribute("detail_level", String, "Detail level", func() {
				Enum("basic", "standard", "comprehensive")
				Default("standard")
			})
		})
		Result(SystemHealth)
		HTTP(func() {
			GET("/health")
			Response(StatusOK)
		})
	})

	Method("get_performance_metrics", func() {
		Description("Real-time Performance Metrics")
		Payload(func() {
			Attribute("period", String, "Time period", func() {
				Enum("1m", "5m", "1h", "6h", "24h", "7d")
				Default("1h")
			})
			Attribute("granularity", String, "Granularity", func() {
				Enum("1s", "1m", "5m", "1h")
				Default("1m")
			})
			Attribute("include_breakdown", Boolean, "Include component-level breakdown", func() {
				Default(true)
			})
			Attribute("include_trends", Boolean, "Include trend analysis", func() {
				Default(true)
			})
			Attribute("metrics", String, "Comma-separated list of specific metrics")
		})
		Result(PerformanceMetricsResponse)
		HTTP(func() {
			GET("/metrics")
			Response(StatusOK)
		})
	})

	Method("get_cache_status", func() {
		Description("Cache Status and Management")
		Payload(func() {
			Attribute("include_statistics", Boolean, "Include statistics", func() {
				Default(true)
			})
			Attribute("include_health", Boolean, "Include health", func() {
				Default(true)
			})
		})
		Result(CacheStatusResponse)
		HTTP(func() {
			GET("/cache/status")
			Response(StatusOK)
		})
	})

	Method("manage_cache", func() {
		Description("Cache Operations")
		Payload(ManageCacheRequest)
		Result(ManageCacheResponse)
		HTTP(func() {
			POST("/cache/operations")
			Response(StatusOK)
		})
	})

	Method("get_system_configuration", func() {
		Description("System Configuration")
		Payload(func() {
			Attribute("include_defaults", Boolean, "Include defaults", func() {
				Default(false)
			})
			Attribute("include_sensitive", Boolean, "Include sensitive", func() {
				Default(false)
			})
		})
		Result(SystemConfigurationResponse)
		HTTP(func() {
			GET("/config")
			Response(StatusOK)
		})
	})

	Method("update_system_configuration", func() {
		Description("Update Configuration")
		Payload(UpdateSystemConfigurationRequest)
		Result(UpdateSystemConfigurationResponse)
		HTTP(func() {
			PUT("/config")
			Response(StatusOK)
		})
	})

	Method("analyze_performance", func() {
		Description("Performance Analysis")
		Payload(AnalyzePerformanceRequest)
		Result(AnalyzePerformanceResponse)
		HTTP(func() {
			POST("/analysis/performance")
			Response(StatusOK)
		})
	})

	Method("apply_optimizations", func() {
		Description("Apply Performance Optimizations")
		Payload(ApplyOptimizationsRequest)
		Result(ApplyOptimizationsResponse)
		HTTP(func() {
			POST("/optimization/{analysis_id}/apply")
			Response(StatusAccepted)
		})
	})

	Method("get_system_statistics", func() {
		Description(" System Statistics")
		Payload(func() {
			Attribute("period", String, "Time period", func() {
				Default("24h")
			})
			Attribute("include_trends", Boolean, "Include trends", func() {
				Default(true)
			})
			Attribute("include_forecasts", Boolean, "Include forecasts", func() {
				Default(true)
			})
		})
		Result(SystemStatisticsResponse)
		HTTP(func() {
			GET("/statistics")
			Response(StatusOK)
		})
	})
})
