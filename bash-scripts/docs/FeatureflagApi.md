# FeatureflagApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**featureflagCreate**](FeatureflagApi.md#featureflagCreate) | **POST** /api/v1/feature-flags | create featureflag
[**featureflagDelete**](FeatureflagApi.md#featureflagDelete) | **DELETE** /api/v1/feature-flags/{id} | delete featureflag
[**featureflagEvaluate**](FeatureflagApi.md#featureflagEvaluate) | **POST** /api/v1/feature-flags/evaluate/{name} | evaluate featureflag
[**featureflagEvaluateMultiple**](FeatureflagApi.md#featureflagEvaluateMultiple) | **POST** /api/v1/feature-flags/evaluate | evaluateMultiple featureflag
[**featureflagGet**](FeatureflagApi.md#featureflagGet) | **GET** /api/v1/feature-flags/{name} | get featureflag
[**featureflagGetById**](FeatureflagApi.md#featureflagGetById) | **GET** /api/v1/feature-flags/by-id/{id} | getById featureflag
[**featureflagGetByType**](FeatureflagApi.md#featureflagGetByType) | **GET** /api/v1/feature-flags/type/{flag_type} | getByType featureflag
[**featureflagGetStats**](FeatureflagApi.md#featureflagGetStats) | **GET** /api/v1/feature-flags/stats | getStats featureflag
[**featureflagHealth**](FeatureflagApi.md#featureflagHealth) | **GET** /api/v1/feature-flags/health | health featureflag
[**featureflagList**](FeatureflagApi.md#featureflagList) | **GET** /api/v1/feature-flags | list featureflag
[**featureflagSearch**](FeatureflagApi.md#featureflagSearch) | **GET** /api/v1/feature-flags/search | search featureflag
[**featureflagUpdate**](FeatureflagApi.md#featureflagUpdate) | **PUT** /api/v1/feature-flags/{id} | update featureflag



## featureflagCreate

create featureflag

Create a new feature flag

### Example

```bash
 featureflagCreate
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createFeatureFlagPayload** | [**CreateFeatureFlagPayload**](CreateFeatureFlagPayload.md) |  |

### Return type

[**Featureflag**](Featureflag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagDelete

delete featureflag

Delete a feature flag (soft delete)

### Example

```bash
 featureflagDelete id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Feature flag ID | [default to null]

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagEvaluate

evaluate featureflag

Evaluate a feature flag for a specific context

### Example

```bash
 featureflagEvaluate name=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Feature flag name | [default to null]
 **evaluateFeatureFlagPayload2** | [**EvaluateFeatureFlagPayload2**](EvaluateFeatureFlagPayload2.md) |  |

### Return type

[**EvaluationResult**](EvaluationResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagEvaluateMultiple

evaluateMultiple featureflag

Evaluate multiple feature flags for a specific context

### Example

```bash
 featureflagEvaluateMultiple
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **bulkEvaluatePayload** | [**BulkEvaluatePayload**](BulkEvaluatePayload.md) |  |

### Return type

[**EvaluateMultipleResponseBody**](EvaluateMultipleResponseBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagGet

get featureflag

Get feature flag by name

### Example

```bash
 featureflagGet name=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Feature flag name | [default to null]

### Return type

[**Featureflag**](Featureflag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagGetById

getById featureflag

Get feature flag by ID

### Example

```bash
 featureflagGetById id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Feature flag ID | [default to null]

### Return type

[**Featureflag**](Featureflag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagGetByType

getByType featureflag

Get all feature flags of a specific type

### Example

```bash
 featureflagGetByType flag_type=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **flagType** | **string** | Flag type | [default to null]

### Return type

[**GetByTypeResponseBody**](GetByTypeResponseBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagGetStats

getStats featureflag

Get feature flag usage statistics

### Example

```bash
 featureflagGetStats
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**FeatureFlagStats**](FeatureFlagStats.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagHealth

health featureflag

Health check for feature flag service

### Example

```bash
 featureflagHealth
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HealthResponseBody**](HealthResponseBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagList

list featureflag

List feature flags with pagination and filtering

### Example

```bash
 featureflagList  page=value  page_size=value  sort_by=value  sort_order=value  flag_type=value  name_filter=value  enabled_only=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **integer** | Page number (1-based) | [optional] [default to 1]
 **pageSize** | **integer** | Number of items per page | [optional] [default to 20]
 **sortBy** | **string** | Field to sort by | [optional] [default to null]
 **sortOrder** | **string** | Sort order | [optional] [default to desc]
 **flagType** | **string** | Filter by flag type | [optional] [default to null]
 **nameFilter** | **string** | Filter by flag name | [optional] [default to null]
 **enabledOnly** | **boolean** | Show only enabled flags | [optional] [default to null]

### Return type

[**ListResponseBody**](ListResponseBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagSearch

search featureflag

Search feature flags by name or description

### Example

```bash
 featureflagSearch  query=value  limit=value  offset=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **query** | **string** | Search query | [default to null]
 **limit** | **integer** | Maximum number of results | [optional] [default to 10]
 **offset** | **integer** | Number of results to skip | [optional] [default to 0]

### Return type

[**SearchResponseBody**](SearchResponseBody.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## featureflagUpdate

update featureflag

Update an existing feature flag

### Example

```bash
 featureflagUpdate id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Feature flag ID | [default to null]
 **updateFeatureFlagPayload2** | [**UpdateFeatureFlagPayload2**](UpdateFeatureFlagPayload2.md) |  |

### Return type

[**Featureflag**](Featureflag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

