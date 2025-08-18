# AdminFeatureflagApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**adminFeatureflagBulkDisable**](AdminFeatureflagApi.md#adminFeatureflagBulkDisable) | **POST** /api/v1/admin/feature-flags/bulk/disable | bulk_disable admin-featureflag
[**adminFeatureflagBulkEnable**](AdminFeatureflagApi.md#adminFeatureflagBulkEnable) | **POST** /api/v1/admin/feature-flags/bulk/enable | bulk_enable admin-featureflag
[**adminFeatureflagSystemHealth**](AdminFeatureflagApi.md#adminFeatureflagSystemHealth) | **GET** /api/v1/admin/feature-flags/health | system_health admin-featureflag



## adminFeatureflagBulkDisable

bulk_disable admin-featureflag

Disable multiple feature flags in bulk

### Example

```bash
 adminFeatureflagBulkDisable X-Tenant-ID:value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xTenantID** | **string** | Tenant ID | [default to null]
 **bulkEnableRequestBody** | [**BulkEnableRequestBody**](BulkEnableRequestBody.md) |  |

### Return type

[**BulkEnableResponseBody**](BulkEnableResponseBody.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## adminFeatureflagBulkEnable

bulk_enable admin-featureflag

Enable multiple feature flags in bulk

### Example

```bash
 adminFeatureflagBulkEnable X-Tenant-ID:value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xTenantID** | **string** | Tenant ID | [default to null]
 **bulkEnableRequestBody** | [**BulkEnableRequestBody**](BulkEnableRequestBody.md) |  |

### Return type

[**BulkEnableResponseBody**](BulkEnableResponseBody.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## adminFeatureflagSystemHealth

system_health admin-featureflag

Get system health status

### Example

```bash
 adminFeatureflagSystemHealth X-Tenant-ID:value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xTenantID** | **string** | Tenant ID | [default to null]

### Return type

[**SystemHealthResponseBody**](SystemHealthResponseBody.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

