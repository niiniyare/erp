# AuthApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**authLogin**](AuthApi.md#authLogin) | **POST** /api/v1/auth/login | login auth
[**authLogout**](AuthApi.md#authLogout) | **POST** /api/v1/auth/logout | logout auth
[**authRefresh**](AuthApi.md#authRefresh) | **POST** /api/v1/auth/refresh | refresh auth
[**authValidate**](AuthApi.md#authValidate) | **GET** /api/v1/auth/validate | validate auth



## authLogin

login auth

Authenticate user and return JWT token

### Example

```bash
 authLogin
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginRequestBody** | [**LoginRequestBody**](LoginRequestBody.md) |  |

### Return type

[**Auth**](Auth.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## authLogout

logout auth

Logout user and invalidate token

### Example

```bash
 authLogout
```

### Parameters

This endpoint does not need any parameter.

### Return type

(empty response body)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## authRefresh

refresh auth

Refresh JWT token

### Example

```bash
 authRefresh
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refreshRequestBody** | [**RefreshRequestBody**](RefreshRequestBody.md) |  |

### Return type

[**Auth**](Auth.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## authValidate

validate auth

Validate JWT token

### Example

```bash
 authValidate
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**TokenValidation**](TokenValidation.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

