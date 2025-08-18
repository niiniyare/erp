# HealthApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**healthHealth**](HealthApi.md#healthHealth) | **GET** /health | health health
[**healthReady**](HealthApi.md#healthReady) | **GET** /ready | ready health



## healthHealth

health health

Check if the service is running

### Example

```bash
 healthHealth
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HealthStatus**](HealthStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## healthReady

ready health

Check if the service is ready to handle requests

### Example

```bash
 healthReady
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**ReadinessStatus**](ReadinessStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

