package admin

import (
	. "goa.design/goa/v3/dsl"
)

// SystemHealth is the top-level health check response.
var SystemHealth = Type("SystemHealth", func() {
	Attribute("overall_status", String, func() { Enum("healthy", "warning", "critical", "maintenance") })
	Attribute("health_score", Float64)
	Attribute("last_check", String, func() { Format(FormatDateTime) })
	Attribute("system_info", SystemInfo)
	Attribute("core_components", MapOf(String, ComponentHealth))
	Attribute("infrastructure", InfrastructureHealth)
	Attribute("external_dependencies", MapOf(String, DependencyHealth))
	Attribute("performance_metrics", PerformanceMetrics)
	Attribute("capacity_analysis", CapacityAnalysis)
	Attribute("alerts", ArrayOf(SystemAlert))
	Attribute("recommendations", ArrayOf(String))
	Required("overall_status", "health_score", "last_check", "core_components")
})

var SystemInfo = Type("SystemInfo", func() {
	Attribute("version", String)
	Attribute("build", String)
	Attribute("deployment", String)
	Attribute("uptime_seconds", Int64)
	Attribute("uptime_percentage_30d", Float64)
	Attribute("instance_id", String)
})

var ComponentHealth = Type("ComponentHealth", func() {
	Attribute("status", String, func() { Enum("healthy", "warning", "critical", "offline") })
	Attribute("health_score", Float64)
	Attribute("response_time_ms", Float64)
	Attribute("last_check", String, func() { Format(FormatDateTime) })
	Attribute("error_rate", Float64)
	Attribute("throughput_per_second", Float64)
	Attribute("active_instances", Int)
	Attribute("load_balanced", Boolean)
})

var InfrastructureHealth = Type("InfrastructureHealth", func() {
	Attribute("database", DatabaseHealth)
	Attribute("message_queue", MessageQueueHealth)
	Attribute("storage", StorageHealth)
})

var DatabaseHealth = Type("DatabaseHealth", func() {
	Attribute("primary", DatabaseInstanceHealth)
	Attribute("replica", DatabaseInstanceHealth)
})

var DatabaseInstanceHealth = Type("DatabaseInstanceHealth", func() {
	Attribute("status", String)
	Attribute("connection_pool", ConnectionPoolHealth)
	Attribute("query_performance", QueryPerformanceHealth)
	Attribute("replication", ReplicationHealth)
})

var ConnectionPoolHealth = Type("ConnectionPoolHealth", func() {
	Attribute("active", Int)
	Attribute("idle", Int)
	Attribute("max", Int)
	Attribute("utilization", Float64)
})

var QueryPerformanceHealth = Type("QueryPerformanceHealth", func() {
	Attribute("avg_duration_ms", Float64)
	Attribute("slow_queries_count", Int)
	Attribute("deadlocks_count", Int)
})

var ReplicationHealth = Type("ReplicationHealth", func() {
	Attribute("status", String)
	Attribute("lag_ms", Float64)
	Attribute("replicas_count", Int)
})

var MessageQueueHealth = Type("MessageQueueHealth", func() {
	Attribute("status", String)
	Attribute("queue_depth", Int)
	Attribute("processing_rate", Float64)
	Attribute("dead_letter_count", Int)
	Attribute("consumer_lag_ms", Float64)
})

var StorageHealth = Type("StorageHealth", func() {
	Attribute("audit_storage", StorageDetail)
	Attribute("cache_storage", StorageDetail)
})

var StorageDetail = Type("StorageDetail", func() {
	Attribute("utilization", Float64)
	Attribute("iops_current", Float64)
	Attribute("iops_limit", Float64)
	Attribute("throughput_mbps", Float64)
})

var DependencyHealth = Type("DependencyHealth", func() {
	Attribute("status", String)
	Attribute("response_time_ms", Float64)
	Attribute("availability", Float64)
	Attribute("last_sync", String, func() { Format(FormatDateTime) })
})

var PerformanceMetrics = Type("PerformanceMetrics", func() {
	Attribute("current_load", CurrentLoad)
	Attribute("resource_utilization", ResourceUtilization)
	Attribute("sla_compliance", SLACompliance)
})

var CurrentLoad = Type("CurrentLoad", func() {
	Attribute("evaluations_per_second", Float64)
	Attribute("active_sessions", Int)
	Attribute("concurrent_requests", Int)
	Attribute("queue_depth", Int)
})

var ResourceUtilization = Type("ResourceUtilization", func() {
	Attribute("cpu_percent", Float64)
	Attribute("memory_percent", Float64)
	Attribute("disk_io_percent", Float64)
	Attribute("network_io_mbps", Float64)
})

var SLACompliance = Type("SLACompliance", func() {
	Attribute("response_time_sla", SLADetail)
	Attribute("availability_sla", SLADetail)
	Attribute("throughput_sla", SLADetail)
})

var SLADetail = Type("SLADetail", func() {
	Attribute("target_ms", Float64)
	Attribute("current_p99_ms", Float64)
	Attribute("compliance_percentage", Float64)
})

var CapacityAnalysis = Type("CapacityAnalysis", func() {
	Attribute("current_capacity", CapacityDetail)
	Attribute("growth_trends", GrowthTrend)
})

var CapacityDetail = Type("CapacityDetail", func() {
	Attribute("utilized_percentage", Float64)
	Attribute("headroom_percentage", Float64)
	Attribute("scale_out_threshold", Float64)
})

var GrowthTrend = Type("GrowthTrend", func() {
	Attribute("daily_growth_percentage", Float64)
	Attribute("projected_capacity_exhaustion", String, func() { Format(FormatDateTime) })
	Attribute("recommended_scaling_date", String, func() { Format(FormatDateTime) })
})

var SystemAlert = Type("SystemAlert", func() {
	Attribute("level", String)
	Attribute("component", String)
	Attribute("message", String)
	Attribute("threshold", Float64)
	Attribute("current_value", Float64)
	Attribute("recommended_action", String)
	Attribute("alert_time", String, func() { Format(FormatDateTime) })
})

// PerformanceMetricsResponse is the response for the GetPerformanceMetrics endpoint.
var PerformanceMetricsResponse = Type("PerformanceMetricsResponse", func() {
	Attribute("metrics_period", String)
	Attribute("granularity", String)
	Attribute("generated_at", String, func() { Format(FormatDateTime) })
	Attribute("data_points", Int)
	Attribute("overview", PerformanceOverview)
	Attribute("evaluation_metrics", EvaluationMetricsBreakdown)
	Attribute("cache_metrics", CacheMetricsBreakdown)
	Attribute("policy_metrics", PolicyMetricsBreakdown)
	Attribute("attribute_metrics", AttributeMetricsBreakdown)
	Attribute("time_series_data", ArrayOf(TimeSeriesDataPoint))
	Attribute("trends", PerformanceTrends)
})

var PerformanceOverview = Type("PerformanceOverview", func() {
	Attribute("total_evaluations", Int64)
	Attribute("avg_evaluations_per_second", Float64)
	Attribute("peak_evaluations_per_second", Float64)
	Attribute("min_evaluations_per_second", Float64)
	Attribute("evaluation_success_rate", Float64)
	Attribute("avg_evaluation_time_ms", Float64)
	Attribute("p50_evaluation_time_ms", Float64)
	Attribute("p95_evaluation_time_ms", Float64)
	Attribute("p99_evaluation_time_ms", Float64)
	Attribute("error_rate", Float64)
})

var EvaluationMetricsBreakdown = Type("EvaluationMetricsBreakdown", func() {
	Attribute("by_decision", MapOf(String, Int64))
	Attribute("by_complexity", MapOf(String, ComplexityMetrics))
	Attribute("performance_distribution", ArrayOf(PerformanceDistributionRange))
})

var ComplexityMetrics = Type("ComplexityMetrics", func() {
	Attribute("count", Int64)
	Attribute("avg_time_ms", Float64)
	Attribute("percentage", Float64)
})

var PerformanceDistributionRange = Type("PerformanceDistributionRange", func() {
	Attribute("range_ms", String)
	Attribute("count", Int64)
	Attribute("percentage", Float64)
})

var CacheMetricsBreakdown = Type("CacheMetricsBreakdown", func() {
	Attribute("policy_cache", CacheDetailMetrics)
	Attribute("attribute_cache", CacheDetailMetrics)
	Attribute("evaluation_cache", CacheDetailMetrics)
})

var CacheDetailMetrics = Type("CacheDetailMetrics", func() {
	Attribute("hit_rate", Float64)
	Attribute("miss_rate", Float64)
	Attribute("eviction_rate", Float64)
	Attribute("total_requests", Int64)
	Attribute("hits", Int64)
	Attribute("misses", Int64)
	Attribute("evictions", Int64)
	Attribute("avg_retrieval_time_ms", Float64)
	Attribute("memory_usage_mb", Float64)
	Attribute("entries_count", Int64)
})

var PolicyMetricsBreakdown = Type("PolicyMetricsBreakdown", func() {
	Attribute("total_policies_active", Int)
	Attribute("policies_evaluated", Int64)
	Attribute("unique_policies_triggered", Int)
	Attribute("most_frequent_policies", ArrayOf(FrequentPolicy))
	Attribute("policy_conflicts_resolved", Int)
	Attribute("policy_compilation_time_ms", Float64)
})

var FrequentPolicy = Type("FrequentPolicy", func() {
	Attribute("policy_id", String)
	Attribute("name", String)
	Attribute("trigger_count", Int64)
	Attribute("trigger_percentage", Float64)
	Attribute("avg_evaluation_time_ms", Float64)
	Attribute("success_rate", Float64)
})

var AttributeMetricsBreakdown = Type("AttributeMetricsBreakdown", func() {
	Attribute("total_attribute_retrievals", Int64)
	Attribute("avg_retrieval_time_ms", Float64)
	Attribute("p95_retrieval_time_ms", Float64)
	Attribute("stale_attributes_percentage", Float64)
	Attribute("failed_retrievals", Int64)
	Attribute("failed_retrievals_percentage", Float64)
	Attribute("attribute_sources_health", MapOf(String, SourceHealth))
})

var SourceHealth = Type("SourceHealth", func() {
	Attribute("success_rate", Float64)
	Attribute("avg_time_ms", Float64)
})

var TimeSeriesDataPoint = Type("TimeSeriesDataPoint", func() {
	Attribute("timestamp", String, func() { Format(FormatDateTime) })
	Attribute("evaluations_per_minute", Int64)
	Attribute("avg_response_time_ms", Float64)
	Attribute("cache_hit_rate", Float64)
	Attribute("error_count", Int)
	Attribute("active_sessions", Int)
})

var PerformanceTrends = Type("PerformanceTrends", func() {
	Attribute("evaluation_volume", TrendDetail)
	Attribute("response_time", TrendDetail)
	Attribute("error_rate", TrendDetail)
})

var TrendDetail = Type("TrendDetail", func() {
	Attribute("trend", String)
	Attribute("change_percentage", Float64)
	Attribute("projected_next_hour", Float64)
})

// CacheStatusResponse is the response for the GetCacheStatus endpoint.
var CacheStatusResponse = Type("CacheStatusResponse", func() {
	Attribute("cache_overview", CacheOverview)
	Attribute("cache_layers", MapOf(String, CacheLayerStatus))
	Attribute("cache_health", CacheHealth)
	Attribute("performance_impact", CachePerformanceImpact)
	Attribute("recommendations", ArrayOf(String))
})

var CacheOverview = Type("CacheOverview", func() {
	Attribute("total_memory_allocated_mb", Float64)
	Attribute("total_memory_used_mb", Float64)
	Attribute("total_memory_utilization", Float64)
	Attribute("total_entries", Int64)
	Attribute("global_hit_rate", Float64)
	Attribute("last_updated", String, func() { Format(FormatDateTime) })
})

var CacheLayerStatus = Type("CacheLayerStatus", func() {
	Attribute("status", String)
	Attribute("enabled", Boolean)
	Attribute("memory_allocated_mb", Float64)
	Attribute("memory_used_mb", Float64)
	Attribute("memory_utilization", Float64)
	Attribute("entries_count", Int64)
	Attribute("max_entries", Int64)
	Attribute("hit_rate", Float64)
	Attribute("miss_rate", Float64)
	Attribute("eviction_rate", Float64)
	Attribute("ttl_seconds", Int)
	Attribute("eviction_policy", String)
	Attribute("last_eviction", String, func() { Format(FormatDateTime) })
	Attribute("performance", CacheLayerPerformance)
})

var CacheLayerPerformance = Type("CacheLayerPerformance", func() {
	Attribute("avg_get_time_ms", Float64)
	Attribute("avg_set_time_ms", Float64)
	Attribute("operations_per_second", Float64)
})

var CacheHealth = Type("CacheHealth", func() {
	Attribute("overall_health", String)
	Attribute("performance_score", Float64)
	Attribute("efficiency_score", Float64)
	Attribute("memory_pressure", String)
	Attribute("eviction_pressure", String)
	Attribute("fragmentation_ratio", Float64)
})

var CachePerformanceImpact = Type("CachePerformanceImpact", func() {
	Attribute("cache_enabled_avg_time_ms", Float64)
	Attribute("estimated_nocache_time_ms", Float64)
	Attribute("performance_improvement_percentage", Float64)
	Attribute("cache_value_score", Float64)
})

// ManageCacheRequest is the payload for the ManageCache endpoint.
var ManageCacheRequest = Type("ManageCacheRequest", func() {
	Attribute("operation", String)
	Attribute("cache_scope", CacheScope)
	Attribute("options", CacheOperationOptions)
	Attribute("metadata", CacheOperationMetadata)
})

var CacheScope = Type("CacheScope", func() {
	Attribute("cache_types", ArrayOf(String))
	Attribute("selective_filters", SelectiveCacheFilters)
})

var SelectiveCacheFilters = Type("SelectiveCacheFilters", func() {
	Attribute("user_ids", ArrayOf(String))
	Attribute("policy_ids", ArrayOf(String))
	Attribute("resource_patterns", ArrayOf(String))
	Attribute("attribute_patterns", ArrayOf(String))
	Attribute("age_older_than_minutes", Int)
})

var CacheOperationOptions = Type("CacheOperationOptions", func() {
	Attribute("clear_all", Boolean)
	Attribute("warm_cache", Boolean)
	Attribute("async_operation", Boolean)
	Attribute("notify_on_completion", Boolean)
})

var CacheOperationMetadata = Type("CacheOperationMetadata", func() {
	Attribute("reason", String)
	Attribute("initiated_by", String)
	Attribute("change_request_id", String)
})

// ManageCacheResponse is the response for the ManageCache endpoint.
var ManageCacheResponse = Type("ManageCacheResponse", func() {
	Attribute("operation_id", String)
	Attribute("operation", String)
	Attribute("status", String)
	Attribute("started_at", String, func() { Format(FormatDateTime) })
	Attribute("estimated_completion", String, func() { Format(FormatDateTime) })
	Attribute("progress", OperationProgress)
	Attribute("results", CacheOperationResults)
	Attribute("selective_clear_results", SelectiveClearResults)
	Attribute("cache_warming", CacheWarmingStatus)
	Attribute("performance_impact", OperationPerformanceImpact)
})

var OperationProgress = Type("OperationProgress", func() {
	Attribute("total_steps", Int)
	Attribute("completed_steps", Int)
	Attribute("current_step", String)
	Attribute("percentage_complete", Float64)
})

var CacheOperationResults = Type("CacheOperationResults", func() {
	Attribute("evaluation_cache", CacheOperationResultDetail)
	Attribute("attribute_cache", CacheOperationResultDetail)
	Attribute("policy_cache", CacheOperationResultDetail)
})

var CacheOperationResultDetail = Type("CacheOperationResultDetail", func() {
	Attribute("status", String)
	Attribute("entries_cleared", Int64)
	Attribute("cache_size_before_mb", Float64)
	Attribute("cache_size_after_mb", Float64)
	Attribute("operation_time_ms", Float64)
})

var SelectiveClearResults = Type("SelectiveClearResults", func() {
	Attribute("users_affected", Int)
	Attribute("policies_affected", Int)
	Attribute("pattern_matches", MapOf(String, Int64))
	Attribute("age_based_clears", Int64)
})

var CacheWarmingStatus = Type("CacheWarmingStatus", func() {
	Attribute("enabled", Boolean)
	Attribute("status", String)
	Attribute("priority_items", Int)
	Attribute("estimated_warm_time_minutes", Int)
})

var OperationPerformanceImpact = Type("OperationPerformanceImpact", func() {
	Attribute("expected_cache_miss_rate_increase", Float64)
	Attribute("temporary_latency_increase_ms", Float64)
	Attribute("recovery_time_estimate_minutes", Int)
	Attribute("impact_severity", String)
})

// SystemConfigurationResponse is the response for the GetSystemConfiguration endpoint.
var SystemConfigurationResponse = Type("SystemConfigurationResponse", func() {
	Attribute("configuration_version", String)
	Attribute("last_updated", String, func() { Format(FormatDateTime) })
	Attribute("tenant_config", TenantConfiguration)
	Attribute("system_defaults", SystemDefaults)
	Attribute("feature_flags", MapOf(String, Boolean))
	Attribute("integration_settings", IntegrationSettings)
})

var TenantConfiguration = Type("TenantConfiguration", func() {
	Attribute("tenant_id", String)
	Attribute("tenant_name", String)
	Attribute("policy_evaluation", PolicyEvaluationConfig)
	Attribute("security_settings", SecuritySettingsConfig)
	Attribute("audit_settings", AuditSettingsConfig)
	Attribute("performance_limits", PerformanceLimitsConfig)
	Attribute("attribute_management", AttributeManagementConfig)
})

var PolicyEvaluationConfig = Type("PolicyEvaluationConfig", func() {
	Attribute("max_evaluation_time_ms", Int)
	Attribute("default_decision", String)
	Attribute("cache_enabled", Boolean)
	Attribute("cache_ttl_seconds", Int)
	Attribute("parallel_evaluation", Boolean)
	Attribute("max_policy_depth", Int)
	Attribute("policy_combining_algorithm", String)
	Attribute("obligation_execution_timeout_ms", Int)
})

var SecuritySettingsConfig = Type("SecuritySettingsConfig", func() {
	Attribute("require_mfa_for_sensitive", Boolean)
	Attribute("max_failed_evaluations", Int)
	Attribute("lockout_duration_minutes", Int)
	Attribute("enable_anomaly_detection", Boolean)
	Attribute("risk_threshold", Int)
	Attribute("session_timeout_minutes", Int)
	Attribute("enable_geolocation_checks", Boolean)
	Attribute("enable_device_fingerprinting", Boolean)
})

var AuditSettingsConfig = Type("AuditSettingsConfig", func() {
	Attribute("log_all_evaluations", Boolean)
	Attribute("log_attribute_access", Boolean)
	Attribute("log_policy_changes", Boolean)
	Attribute("retention_days", Int)
	Attribute("compliance_standards", ArrayOf(String))
	Attribute("real_time_monitoring", Boolean)
	Attribute("enable_behavioral_analytics", Boolean)
	Attribute("data_export_encryption", Boolean)
})

var PerformanceLimitsConfig = Type("PerformanceLimitsConfig", func() {
	Attribute("max_requests_per_second", Int)
	Attribute("max_concurrent_evaluations", Int)
	Attribute("rate_limiting_enabled", Boolean)
	Attribute("circuit_breaker_enabled", Boolean)
	Attribute("circuit_breaker_threshold", Float64)
	Attribute("bulk_operation_limit", Int)
	Attribute("query_timeout_seconds", Int)
})

var AttributeManagementConfig = Type("AttributeManagementConfig", func() {
	Attribute("attribute_validation", String)
	Attribute("auto_derive_attributes", Boolean)
	Attribute("attribute_encryption", Boolean)
	Attribute("attribute_versioning", Boolean)
	Attribute("stale_attribute_threshold_minutes", Int)
	Attribute("attribute_sources_timeout_ms", Int)
})

var SystemDefaults = Type("SystemDefaults", func() {
	Attribute("default_cache_ttl_seconds", Int)
	Attribute("default_session_timeout_minutes", Int)
	Attribute("default_rate_limit", Int)
	Attribute("default_batch_size", Int)
	Attribute("default_retry_attempts", Int)
	Attribute("default_circuit_breaker_timeout_ms", Int)
})

var IntegrationSettings = Type("IntegrationSettings", func() {
	Attribute("external_attribute_providers", ArrayOf(ExternalProvider))
	Attribute("notification_endpoints", ArrayOf(NotificationEndpoint))
})

var ExternalProvider = Type("ExternalProvider", func() {
	Attribute("name", String)
	Attribute("endpoint", String)
	Attribute("timeout_ms", Int)
	Attribute("retry_attempts", Int)
	Attribute("health_check_interval_minutes", Int)
})

var NotificationEndpoint = Type("NotificationEndpoint", func() {
	Attribute("name", String)
	Attribute("type", String)
	Attribute("endpoint", String)
	Attribute("authentication", String)
})

// UpdateSystemConfigurationRequest is the payload for the UpdateSystemConfiguration endpoint.
var UpdateSystemConfigurationRequest = Type("UpdateSystemConfigurationRequest", func() {
	Attribute("configuration_changes", Any)
	Attribute("change_metadata", ChangeMetadata)
	Attribute("validation_options", ValidationOptions)
})

var ChangeMetadata = Type("ChangeMetadata", func() {
	Attribute("change_reason", String)
	Attribute("change_request_id", String)
	Attribute("approved_by", String)
	Attribute("scheduled_deployment", String)
	Attribute("rollback_plan", String)
})

var ValidationOptions = Type("ValidationOptions", func() {
	Attribute("validate_before_apply", Boolean)
	Attribute("dry_run", Boolean)
	Attribute("backup_current_config", Boolean)
	Attribute("test_configuration", Boolean)
})

// UpdateSystemConfigurationResponse is the response for the UpdateSystemConfiguration endpoint.
var UpdateSystemConfigurationResponse = Type("UpdateSystemConfigurationResponse", func() {
	Attribute("change_id", String)
	Attribute("status", String)
	Attribute("applied_at", String, func() { Format(FormatDateTime) })
	Attribute("validation_results", ValidationResults)
	Attribute("changes_applied", ArrayOf(String))
	Attribute("system_impact", SystemImpact)
	Attribute("rollback_info", RollbackInfo)
	Attribute("monitoring", MonitoringInfo)
})

var ValidationResults = Type("ValidationResults", func() {
	Attribute("config_valid", Boolean)
	Attribute("compatibility_check", String)
	Attribute("performance_impact", String)
	Attribute("security_impact", String)
	Attribute("warnings", ArrayOf(String))
	Attribute("errors", ArrayOf(String))
})

var SystemImpact = Type("SystemImpact", func() {
	Attribute("restart_required", Boolean)
	Attribute("cache_invalidation_triggered", Boolean)
	Attribute("service_disruption", String)
	Attribute("performance_test_results", Any)
})

var RollbackInfo = Type("RollbackInfo", func() {
	Attribute("rollback_available", Boolean)
	Attribute("config_backup_id", String)
	Attribute("rollback_command", String)
})

var MonitoringInfo = Type("MonitoringInfo", func() {
	Attribute("enhanced_monitoring_enabled", Boolean)
	Attribute("monitoring_duration_minutes", Int)
	Attribute("alert_thresholds_adjusted", Boolean)
	Attribute("automatic_rollback_conditions", ArrayOf(String))
})

// AnalyzePerformanceRequest is the payload for the AnalyzePerformance endpoint.
var AnalyzePerformanceRequest = Type("AnalyzePerformanceRequest", func() {
	Attribute("analysis_scope", AnalysisScope)
	Attribute("optimization_targets", ArrayOf(String))
	Attribute("baseline_comparison", BaselineComparison)
})

var AnalysisScope = Type("AnalysisScope", func() {
	Attribute("time_period", String)
	Attribute("include_components", ArrayOf(String))
	Attribute("analysis_depth", String)
	Attribute("include_predictions", Boolean)
})

var BaselineComparison = Type("BaselineComparison", func() {
	Attribute("compare_to_period", String)
	Attribute("include_trend_analysis", Boolean)
	Attribute("highlight_regressions", Boolean)
})

// AnalyzePerformanceResponse is the response for the AnalyzePerformance endpoint.
var AnalyzePerformanceResponse = Type("AnalyzePerformanceResponse", func() {
	Attribute("analysis_id", String)
	Attribute("analysis_completed_at", String, func() { Format(FormatDateTime) })
	Attribute("analysis_period", String)
	Attribute("executive_summary", ExecutiveSummary)
	Attribute("current_performance", CurrentPerformance)
	Attribute("baseline_comparison", BaselineComparisonResult)
	Attribute("component_analysis", MapOf(String, ComponentAnalysis))
	Attribute("optimization_recommendations", ArrayOf(OptimizationRecommendation))
	Attribute("predictive_analysis", PredictiveAnalysis)
	Attribute("cost_optimization", CostOptimization)
})

var ExecutiveSummary = Type("ExecutiveSummary", func() {
	Attribute("overall_performance_score", Float64)
	Attribute("performance_trend", String)
	Attribute("critical_issues", Int)
	Attribute("optimization_opportunities", Int)
	Attribute("estimated_improvement_potential", String)
})

var CurrentPerformance = Type("CurrentPerformance", func() {
	Attribute("avg_evaluation_time_ms", Float64)
	Attribute("p95_evaluation_time_ms", Float64)
	Attribute("p99_evaluation_time_ms", Float64)
	Attribute("cache_hit_rate", Float64)
	Attribute("throughput_peak_rps", Float64)
	Attribute("error_rate", Float64)
	Attribute("availability_percentage", Float64)
})

var BaselineComparisonResult = Type("BaselineComparisonResult", func() {
	Attribute("previous_period", String)
	Attribute("improvements", ArrayOf(MetricChange))
	Attribute("regressions", ArrayOf(MetricChange))
})

var MetricChange = Type("MetricChange", func() {
	Attribute("metric", String)
	Attribute("previous", Float64)
	Attribute("current", Float64)
	Attribute("improvement_percentage", Float64)
	Attribute("regression_percentage", Float64)
	Attribute("severity", String)
})

var ComponentAnalysis = Type("ComponentAnalysis", func() {
	Attribute("performance_score", Float64)
	Attribute("bottlenecks", ArrayOf(Bottleneck))
	Attribute("optimization_opportunities", ArrayOf(OptimizationOpportunity))
})

var Bottleneck = Type("Bottleneck", func() {
	Attribute("issue", String)
	Attribute("impact", String)
	Attribute("frequency", String)
	Attribute("avg_overhead_ms", Float64)
})

var OptimizationOpportunity = Type("OptimizationOpportunity", func() {
	Attribute("opportunity", String)
	Attribute("estimated_improvement", String)
	Attribute("implementation_effort", String)
})

var OptimizationRecommendation = Type("OptimizationRecommendation", func() {
	Attribute("category", String)
	Attribute("recommendation", String)
	Attribute("estimated_improvement", String)
	Attribute("implementation_complexity", String)
	Attribute("resource_impact", String)
	Attribute("implementation_steps", ArrayOf(String))
	Attribute("prerequisites", ArrayOf(String))
	Attribute("cost_analysis", CostAnalysis)
})

var CostAnalysis = Type("CostAnalysis", func() {
	Attribute("additional_cost_monthly", Float64)
	Attribute("performance_value", String)
	Attribute("roi_months", Int)
})

var PredictiveAnalysis = Type("PredictiveAnalysis", func() {
	Attribute("capacity_forecasting", CapacityForecasting)
	Attribute("performance_trends", MapOf(String, String))
	Attribute("risk_factors", ArrayOf(RiskFactor))
})

var CapacityForecasting = Type("CapacityForecasting", func() {
	Attribute("current_utilization", Float64)
	Attribute("growth_rate_daily", Float64)
	Attribute("projected_capacity_exhaustion", String, func() { Format(FormatDateTime) })
	Attribute("recommended_scaling_timeline", String, func() { Format(FormatDateTime) })
})

var RiskFactor = Type("RiskFactor", func() {
	Attribute("factor", String)
	Attribute("probability", Float64)
	Attribute("impact", String)
	Attribute("mitigation", String)
})

var CostOptimization = Type("CostOptimization", func() {
	Attribute("current_monthly_cost", Float64)
	Attribute("optimization_potential", Float64)
	Attribute("savings_opportunities", ArrayOf(SavingsOpportunity))
})

var SavingsOpportunity = Type("SavingsOpportunity", func() {
	Attribute("area", String)
	Attribute("potential_savings", Float64)
	Attribute("implementation_effort", String)
})

// ApplyOptimizationsRequest is the payload for the ApplyOptimizations endpoint.
var ApplyOptimizationsRequest = Type("ApplyOptimizationsRequest", func() {
	Attribute("analysis_id", String)
	Attribute("optimizations_to_apply", ArrayOf(OptimizationToApply))
	Attribute("deployment_options", DeploymentOptions)
	Attribute("safety_measures", SafetyMeasures)
})

var OptimizationToApply = Type("OptimizationToApply", func() {
	Attribute("optimization_id", String)
	Attribute("parameters", Any)
})

var DeploymentOptions = Type("DeploymentOptions", func() {
	Attribute("apply_immediately", Boolean)
	Attribute("scheduled_time", String, func() { Format(FormatDateTime) })
	Attribute("enable_rollback", Boolean)
	Attribute("testing_required", Boolean)
	Attribute("notification_recipients", ArrayOf(String))
})

var SafetyMeasures = Type("SafetyMeasures", func() {
	Attribute("canary_deployment", Boolean)
	Attribute("canary_percentage", Int)
	Attribute("success_criteria", SuccessCriteria)
	Attribute("automatic_rollback", Boolean)
})

var SuccessCriteria = Type("SuccessCriteria", func() {
	Attribute("max_error_rate", Float64)
	Attribute("max_avg_response_time_ms", Int)
	Attribute("min_availability", Float64)
})

// ApplyOptimizationsResponse is the response for the ApplyOptimizations endpoint.
var ApplyOptimizationsResponse = Type("ApplyOptimizationsResponse", func() {
	Attribute("optimization_deployment_id", String)
	Attribute("status", String)
	Attribute("scheduled_for", String, func() { Format(FormatDateTime) })
	Attribute("optimizations_included", Int)
	Attribute("deployment_plan", DeploymentPlan)
	Attribute("expected_improvements", ExpectedImprovements)
	Attribute("monitoring_plan", MonitoringPlan)
	Attribute("rollback_preparation", RollbackPreparation)
})

var DeploymentPlan = Type("DeploymentPlan", func() {
	Attribute("phases", ArrayOf(DeploymentPhase))
})

var DeploymentPhase = Type("DeploymentPhase", func() {
	Attribute("phase", Int)
	Attribute("name", String)
	Attribute("start_time", String, func() { Format(FormatDateTime) })
	Attribute("duration_minutes", Int)
	Attribute("activities", ArrayOf(String))
})

var ExpectedImprovements = Type("ExpectedImprovements", func() {
	Attribute("evaluation_time_reduction", String)
	Attribute("cache_efficiency_improvement", String)
	Attribute("overall_performance_gain", String)
})

var MonitoringPlan = Type("MonitoringPlan", func() {
	Attribute("enhanced_monitoring_duration_hours", Int)
	Attribute("key_metrics_tracked", ArrayOf(String))
	Attribute("alert_thresholds_adjusted", Boolean)
	Attribute("automatic_rollback_conditions", ArrayOf(String))
})

var RollbackPreparation = Type("RollbackPreparation", func() {
	Attribute("rollback_available", Boolean)
	Attribute("rollback_time_estimate_minutes", Int)
	Attribute("rollback_triggers", String)
	Attribute("configuration_backup_id", String)
})

// SystemStatisticsResponse is the response for the GetSystemStatistics endpoint.
var SystemStatisticsResponse = Type("SystemStatisticsResponse", func() {
	Attribute("statistics_period", String)
	Attribute("generated_at", String, func() { Format(FormatDateTime) })
	Attribute("report_version", String)
	Attribute("usage_statistics", UsageStatistics)
	Attribute("security_statistics", SecurityStatistics)
	Attribute("performance_statistics", PerformanceStatistics)
	Attribute("business_insights", BusinessInsights)
	Attribute("trends_analysis", TrendsAnalysis)
	Attribute("forecasts", Forecasts)
	Attribute("recommendations", ArrayOf(String))
})

var UsageStatistics = Type("UsageStatistics", func() {
	Attribute("evaluation_volume", EvaluationVolume)
	Attribute("decision_analysis", DecisionAnalysis)
	Attribute("user_behavior", UserBehavior)
})

var EvaluationVolume = Type("EvaluationVolume", func() {
	Attribute("total_evaluations", Int64)
	Attribute("unique_users", Int)
	Attribute("unique_resources", Int)
	Attribute("unique_policies_triggered", Int)
	Attribute("peak_concurrent_users", Int)
	Attribute("avg_evaluations_per_user", Float64)
	Attribute("busiest_hour", BusiestHour)
	Attribute("evaluation_distribution", MapOf(String, Float64))
})

var BusiestHour = Type("BusiestHour", func() {
	Attribute("hour", String)
	Attribute("evaluations", Int64)
	Attribute("users", Int)
})

var DecisionAnalysis = Type("DecisionAnalysis", func() {
	Attribute("allow_decisions", Int64)
	Attribute("allow_percentage", Float64)
	Attribute("deny_decisions", Int64)
	Attribute("deny_percentage", Float64)
	Attribute("pending_approvals", Int64)
	Attribute("pending_percentage", Float64)
	Attribute("errors", Int64)
	Attribute("error_percentage", Float64)
})

var UserBehavior = Type("UserBehavior", func() {
	Attribute("most_active_users", ArrayOf(ActiveUser))
	Attribute("department_usage", ArrayOf(DepartmentUsage))
})

var ActiveUser = Type("ActiveUser", func() {
	Attribute("user_id", String)
	Attribute("evaluations", Int64)
	Attribute("unique_resources", Int)
	Attribute("risk_score", Float64)
})

var DepartmentUsage = Type("DepartmentUsage", func() {
	Attribute("department", String)
	Attribute("evaluations", Int64)
	Attribute("users", Int)
	Attribute("avg_risk_score", Float64)
})

var SecurityStatistics = Type("SecurityStatistics", func() {
	Attribute("threat_detection", ThreatDetection)
	Attribute("access_patterns", AccessPatterns)
	Attribute("compliance_events", ComplianceEvents)
})

var ThreatDetection = Type("ThreatDetection", func() {
	Attribute("total_security_events", Int)
	Attribute("critical_events", Int)
	Attribute("high_events", Int)
	Attribute("medium_events", Int)
	Attribute("low_events", Int)
	Attribute("false_positives", Int)
	Attribute("true_positives", Int)
	Attribute("detection_accuracy", Float64)
})

var AccessPatterns = Type("AccessPatterns", func() {
	Attribute("anomalous_access_attempts", Int)
	Attribute("blocked_suspicious_activities", Int)
	Attribute("geographic_anomalies", Int)
	Attribute("temporal_anomalies", Int)
	Attribute("behavioral_anomalies", Int)
})

var ComplianceEvents = Type("ComplianceEvents", func() {
	Attribute("gdpr_related_events", Int)
	Attribute("sox_related_events", Int)
	Attribute("hipaa_related_events", Int)
	Attribute("pci_related_events", Int)
	Attribute("data_subject_requests", Int)
	Attribute("privacy_violations", Int)
})

var PerformanceStatistics = Type("PerformanceStatistics", func() {
	Attribute("evaluation_performance", EvaluationPerformance)
	Attribute("system_performance", SystemPerformance)
	Attribute("cache_performance", CachePerformance)
})

var EvaluationPerformance = Type("EvaluationPerformance", func() {
	Attribute("avg_evaluation_time_ms", Float64)
	Attribute("median_evaluation_time_ms", Float64)
	Attribute("p95_evaluation_time_ms", Float64)
	Attribute("p99_evaluation_time_ms", Float64)
	Attribute("fastest_evaluation_ms", Float64)
	Attribute("slowest_evaluation_ms", Float64)
	Attribute("timeout_count", Int)
	Attribute("sla_compliance_percentage", Float64)
})

var SystemPerformance = Type("SystemPerformance", func() {
	Attribute("avg_cpu_utilization", Float64)
	Attribute("peak_cpu_utilization", Float64)
	Attribute("avg_memory_utilization", Float64)
	Attribute("peak_memory_utilization", Float64)
	Attribute("disk_io_avg_iops", Float64)
	Attribute("network_throughput_mbps", Float64)
	Attribute("uptime_percentage", Float64)
	Attribute("restart_count", Int)
})

var CachePerformance = Type("CachePerformance", func() {
	Attribute("overall_hit_rate", Float64)
	Attribute("policy_cache_hit_rate", Float64)
	Attribute("attribute_cache_hit_rate", Float64)
	Attribute("evaluation_cache_hit_rate", Float64)
	Attribute("cache_memory_efficiency", Float64)
	Attribute("eviction_events", Int)
})

var BusinessInsights = Type("BusinessInsights", func() {
	Attribute("productivity_metrics", ProductivityMetrics)
	Attribute("resource_insights", ResourceInsights)
})

var ProductivityMetrics = Type("ProductivityMetrics", func() {
	Attribute("avg_decision_time_reduction", String)
	Attribute("automation_rate", Float64)
	Attribute("manual_intervention_rate", Float64)
	Attribute("compliance_automation_rate", Float64)
	Attribute("cost_per_evaluation", Float64)
})

var ResourceInsights = Type("ResourceInsights", func() {
	Attribute("most_accessed_resources", ArrayOf(AccessedResource))
	Attribute("resource_risk_distribution", MapOf(String, Float64))
})

var AccessedResource = Type("AccessedResource", func() {
	Attribute("resource_type", String)
	Attribute("access_count", Int64)
	Attribute("unique_users", Int)
	Attribute("avg_risk_score", Float64)
})

var TrendsAnalysis = Type("TrendsAnalysis", func() {
	Attribute("volume_trends", VolumeTrends)
	Attribute("performance_trends", MapOf(String, String))
	Attribute("security_trends", MapOf(String, String))
})

var VolumeTrends = Type("VolumeTrends", func() {
	Attribute("daily_growth_rate", Float64)
	Attribute("weekly_growth_rate", Float64)
	Attribute("monthly_growth_rate", Float64)
	Attribute("seasonal_patterns", ArrayOf(String))
})

var Forecasts = Type("Forecasts", func() {
	Attribute("capacity_forecast", CapacityForecast)
	Attribute("performance_forecast", PerformanceForecast)
})

var CapacityForecast = Type("CapacityForecast", func() {
	Attribute("next_30_days", ForecastDetail)
	Attribute("next_90_days", ForecastDetail)
})

var ForecastDetail = Type("ForecastDetail", func() {
	Attribute("predicted_peak_load", Float64)
	Attribute("capacity_utilization", Float64)
	Attribute("scaling_recommendation", String)
	Attribute("predicted_growth", Float64)
	Attribute("infrastructure_changes_needed", Boolean)
	Attribute("cost_impact", Float64)
})

var PerformanceForecast = Type("PerformanceForecast", func() {
	Attribute("expected_response_time_trend", String)
	Attribute("predicted_bottlenecks", ArrayOf(String))
	Attribute("optimization_opportunities", Int)
})
