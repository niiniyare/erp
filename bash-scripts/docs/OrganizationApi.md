# OrganizationApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**organizationArchive**](OrganizationApi.md#organizationArchive) | **PATCH** /api/v1/organizations/{id}/archive | archive organization
[**organizationCreate**](OrganizationApi.md#organizationCreate) | **POST** /api/v1/organizations | create organization
[**organizationGet**](OrganizationApi.md#organizationGet) | **GET** /api/v1/organizations/{id} | get organization
[**organizationHierarchy**](OrganizationApi.md#organizationHierarchy) | **GET** /api/v1/organizations/{id}/hierarchy | hierarchy organization
[**organizationList**](OrganizationApi.md#organizationList) | **GET** /api/v1/organizations | list organization
[**organizationUpdate**](OrganizationApi.md#organizationUpdate) | **PUT** /api/v1/organizations/{id} | update organization



## organizationArchive

archive organization

Archive an organization (soft delete)

### Example

```bash
 organizationArchive id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Organization ID | [default to null]

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## organizationCreate

create organization

Create a new organization

### Example

```bash
 organizationCreate
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createOrganizationPayload** | [**CreateOrganizationPayload**](CreateOrganizationPayload.md) |  |

### Return type

[**Organization**](Organization.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## organizationGet

get organization

Get organization by ID

### Example

```bash
 organizationGet id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Organization ID | [default to null]

### Return type

[**Organization**](Organization.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## organizationHierarchy

hierarchy organization

Get organization hierarchy (parent/child relationships)

### Example

```bash
 organizationHierarchy id=value  depth=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Organization ID | [default to null]
 **depth** | **integer** | Hierarchy depth to retrieve | [optional] [default to 3]

### Return type

[**OrganizationHierarchy**](OrganizationHierarchy.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## organizationList

list organization

List organizations with pagination and filtering

### Example

```bash
 organizationList  page=value  page_size=value  sort_by=value  sort_order=value  name_filter=value  type_filter=value  status_filter=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **integer** | Page number (1-based) | [optional] [default to 1]
 **pageSize** | **integer** | Number of items per page | [optional] [default to 20]
 **sortBy** | **string** | Field to sort by | [optional] [default to null]
 **sortOrder** | **string** | Sort order | [optional] [default to desc]
 **nameFilter** | **string** | Filter by organization name | [optional] [default to null]
 **typeFilter** | **string** | Filter by organization type | [optional] [default to null]
 **statusFilter** | **string** | Filter by status | [optional] [default to null]

### Return type

[**ListResponseBody2**](ListResponseBody2.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## organizationUpdate

update organization

Update an existing organization

### Example

```bash
 organizationUpdate id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Organization ID | [default to null]
 **updateOrganizationPayload2** | [**UpdateOrganizationPayload2**](UpdateOrganizationPayload2.md) |  |

### Return type

[**Organization**](Organization.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

