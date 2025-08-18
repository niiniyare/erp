# UserApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**userAssignRole**](UserApi.md#userAssignRole) | **POST** /api/v1/users/{user_id}/roles | assign_role user
[**userAuthorizeAction**](UserApi.md#userAuthorizeAction) | **POST** /api/v1/users/{user_id}/authorize | authorize_action user
[**userBulkUpdateAttributes**](UserApi.md#userBulkUpdateAttributes) | **POST** /api/v1/users/bulk-attributes | bulk_update_attributes user
[**userCheckPermission**](UserApi.md#userCheckPermission) | **POST** /api/v1/users/{user_id}/check-permission | check_permission user
[**userCreate**](UserApi.md#userCreate) | **POST** /api/v1/users | create user
[**userDeactivate**](UserApi.md#userDeactivate) | **PATCH** /api/v1/users/{id}/deactivate | deactivate user
[**userGet**](UserApi.md#userGet) | **GET** /api/v1/users/{id} | get user
[**userGetAttributes**](UserApi.md#userGetAttributes) | **GET** /api/v1/users/{id}/attributes | get_attributes user
[**userGetSessionAttributes**](UserApi.md#userGetSessionAttributes) | **GET** /api/v1/users/{user_id}/sessions/{session_id}/attributes | get_session_attributes user
[**userGetUserContext**](UserApi.md#userGetUserContext) | **GET** /api/v1/users/{id}/context | get_user_context user
[**userList**](UserApi.md#userList) | **GET** /api/v1/users | list user
[**userPermissions**](UserApi.md#userPermissions) | **GET** /api/v1/users/{id}/permissions | permissions user
[**userRefreshAttributes**](UserApi.md#userRefreshAttributes) | **POST** /api/v1/users/{id}/refresh-attributes | refresh_attributes user
[**userRemoveRole**](UserApi.md#userRemoveRole) | **DELETE** /api/v1/users/{user_id}/roles/{role_id} | remove_role user
[**userSetAttributes**](UserApi.md#userSetAttributes) | **PUT** /api/v1/users/{id}/attributes | set_attributes user
[**userSetSessionContext**](UserApi.md#userSetSessionContext) | **PUT** /api/v1/users/{user_id}/sessions/{session_id}/context | set_session_context user
[**userUpdate**](UserApi.md#userUpdate) | **PUT** /api/v1/users/{id} | update user
[**userValidateAttributes**](UserApi.md#userValidateAttributes) | **POST** /api/v1/users/{id}/validate-attributes | validate_attributes user



## userAssignRole

assign_role user

Assign a role to a user

### Example

```bash
 userAssignRole user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]
 **assignRolePayload2** | [**AssignRolePayload2**](AssignRolePayload2.md) |  |

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userAuthorizeAction

authorize_action user

Authorize user action with ABAC context

### Example

```bash
 userAuthorizeAction user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]
 **authorizeActionPayload2** | [**AuthorizeActionPayload2**](AuthorizeActionPayload2.md) |  |

### Return type

[**AuthorizationResult**](AuthorizationResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userBulkUpdateAttributes

bulk_update_attributes user

Bulk update user attributes for ABAC

### Example

```bash
 userBulkUpdateAttributes
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **bulkUpdateUserAttributesPayload** | [**BulkUpdateUserAttributesPayload**](BulkUpdateUserAttributesPayload.md) |  |

### Return type

[**BulkUpdateUserAttributesResult**](BulkUpdateUserAttributesResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userCheckPermission

check_permission user

Check if user has permission using ABAC context

### Example

```bash
 userCheckPermission user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]
 **checkPermissionPayload2** | [**CheckPermissionPayload2**](CheckPermissionPayload2.md) |  |

### Return type

[**PermissionCheckResult**](PermissionCheckResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userCreate

create user

Create a new user

### Example

```bash
 userCreate
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createUserPayload** | [**CreateUserPayload**](CreateUserPayload.md) |  |

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userDeactivate

deactivate user

Deactivate a user account

### Example

```bash
 userDeactivate id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userGet

get user

Get user by ID

### Example

```bash
 userGet id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userGetAttributes

get_attributes user

Get user attributes for ABAC evaluation

### Example

```bash
 userGetAttributes id=value  include_metadata=value  include_derived=value  fresh_only=value  attribute_filter=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]
 **includeMetadata** | **boolean** | Include attribute metadata | [optional] [default to false]
 **includeDerived** | **boolean** | Include computed attributes | [optional] [default to false]
 **freshOnly** | **boolean** | Only return non-expired attributes | [optional] [default to false]
 **attributeFilter** | **string** | Comma-separated attribute names | [optional] [default to null]

### Return type

[**UserAttributesResult**](UserAttributesResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userGetSessionAttributes

get_session_attributes user

Get user session attributes for ABAC

### Example

```bash
 userGetSessionAttributes user_id=value session_id=value  include_analytics=value  include_risk_assessment=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]
 **sessionId** | **string** | Session ID | [default to null]
 **includeAnalytics** | **boolean** | Include session analytics | [optional] [default to false]
 **includeRiskAssessment** | **boolean** | Include risk assessment | [optional] [default to false]

### Return type

[**SessionAttributesResult**](SessionAttributesResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userGetUserContext

get_user_context user

Get comprehensive user context for ABAC

### Example

```bash
 userGetUserContext id=value  include_derived=value  include_session=value  include_access_patterns=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]
 **includeDerived** | **boolean** | Include derived attributes | [optional] [default to true]
 **includeSession** | **boolean** | Include current session context | [optional] [default to true]
 **includeAccessPatterns** | **boolean** | Include access patterns | [optional] [default to false]

### Return type

[**UserContextResult**](UserContextResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userList

list user

List users with pagination and filtering

### Example

```bash
 userList  page=value  page_size=value  sort_by=value  sort_order=value  email_filter=value  status_filter=value  user_type_filter=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **page** | **integer** | Page number (1-based) | [optional] [default to 1]
 **pageSize** | **integer** | Number of items per page | [optional] [default to 20]
 **sortBy** | **string** | Field to sort by | [optional] [default to null]
 **sortOrder** | **string** | Sort order | [optional] [default to desc]
 **emailFilter** | **string** | Filter by email | [optional] [default to null]
 **statusFilter** | **string** | Filter by status | [optional] [default to null]
 **userTypeFilter** | **string** | Filter by user type | [optional] [default to null]

### Return type

[**ListResponseBody4**](ListResponseBody4.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userPermissions

permissions user

Get user permissions and roles

### Example

```bash
 userPermissions id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]

### Return type

[**UserPermissions**](UserPermissions.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userRefreshAttributes

refresh_attributes user

Refresh user attributes from authoritative sources

### Example

```bash
 userRefreshAttributes id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]
 **refreshAttributesRequestBody** | [**RefreshAttributesRequestBody**](RefreshAttributesRequestBody.md) |  |

### Return type

[**RefreshAttributesResult**](RefreshAttributesResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userRemoveRole

remove_role user

Remove a role from a user

### Example

```bash
 userRemoveRole user_id=value role_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]
 **roleId** | **string** | Role ID | [default to null]

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userSetAttributes

set_attributes user

Set/update user attributes for ABAC

### Example

```bash
 userSetAttributes id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]
 **setUserAttributesPayload2** | [**SetUserAttributesPayload2**](SetUserAttributesPayload2.md) |  |

### Return type

[**SetUserAttributesResult**](SetUserAttributesResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userSetSessionContext

set_session_context user

Set session context for user

### Example

```bash
 userSetSessionContext user_id=value session_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]
 **sessionId** | **string** | Session ID | [default to null]
 **setSessionContextPayload2** | [**SetSessionContextPayload2**](SetSessionContextPayload2.md) |  |

### Return type

[**SetSessionContextResult**](SetSessionContextResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userUpdate

update user

Update an existing user

### Example

```bash
 userUpdate id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]
 **updateUserPayload2** | [**UpdateUserPayload2**](UpdateUserPayload2.md) |  |

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## userValidateAttributes

validate_attributes user

Validate user attributes for ABAC compliance

### Example

```bash
 userValidateAttributes id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | User ID | [default to null]
 **validateAttributesRequestBody** | [**ValidateAttributesRequestBody**](ValidateAttributesRequestBody.md) |  |

### Return type

[**AttributeValidationResult**](AttributeValidationResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json, application/vnd.goa.error

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

