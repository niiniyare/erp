# TenantApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**tenantCreate**](TenantApi.md#tenantCreate) | **POST** /api/v1/tenants | create tenant
[**tenantDelete**](TenantApi.md#tenantDelete) | **DELETE** /api/v1/tenants/{id} | delete tenant
[**tenantGet**](TenantApi.md#tenantGet) | **GET** /api/v1/tenants/{id} | get tenant
[**tenantHealth**](TenantApi.md#tenantHealth) | **GET** /api/v1/tenants/health | health tenant
[**tenantList**](TenantApi.md#tenantList) | **GET** /api/v1/tenants | list tenant
[**tenantUpdate**](TenantApi.md#tenantUpdate) | **PUT** /api/v1/tenants/{id} | update tenant



## tenantCreate

create tenant

Create a new tenant

### Example

```bash
 tenantCreate
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createTenantPayload** | [**CreateTenantPayload**](CreateTenantPayload.md) |  |

### Return type

[**Tenant**](Tenant.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## tenantDelete

delete tenant

Delete a tenant (soft delete)

### Example

```bash
 tenantDelete id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Tenant ID | [default to null]

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## tenantGet

get tenant

Get tenant by ID

### Example

```bash
 tenantGet id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Tenant ID | [default to null]

### Return type

[**Tenant**](Tenant.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## tenantHealth

health tenant

Health check for tenant service

### Example

```bash
 tenantHealth
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HealthResponseBody2**](HealthResponseBody2.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## tenantList

list tenant

List tenants with pagination and filtering

### Example

```bash
 tenantList  page=value  page_size=value  sort_by=value  sort_order=value  name_filter=value  status_filter=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **integer** | Page number (1-based) | [optional] [default to 1]
 **pageSize** | **integer** | Number of items per page | [optional] [default to 20]
 **sortBy** | **string** | Field to sort by | [optional] [default to null]
 **sortOrder** | **string** | Sort order | [optional] [default to desc]
 **nameFilter** | **string** | Filter by tenant name | [optional] [default to null]
 **statusFilter** | **string** | Filter by status | [optional] [default to null]

### Return type

[**ListResponseBody3**](ListResponseBody3.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## tenantUpdate

update tenant

Update an existing tenant

### Example

```bash
 tenantUpdate id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Tenant ID | [default to null]
 **updateTenantPayload2** | [**UpdateTenantPayload2**](UpdateTenantPayload2.md) |  |

### Return type

[**Tenant**](Tenant.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

