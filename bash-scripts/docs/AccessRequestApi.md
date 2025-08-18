# AccessRequestApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**accessRequestCreate**](AccessRequestApi.md#accessRequestCreate) | **POST** /api/v1/access-requests | create access_request
[**accessRequestCreateConditionalRule**](AccessRequestApi.md#accessRequestCreateConditionalRule) | **POST** /api/v1/conditional-access/rules | create_conditional_rule access_request
[**accessRequestDetectAnomalies**](AccessRequestApi.md#accessRequestDetectAnomalies) | **POST** /api/v1/analytics/users/{user_id}/detect-anomalies | detect_anomalies access_request
[**accessRequestEvaluateConditionalAccess**](AccessRequestApi.md#accessRequestEvaluateConditionalAccess) | **POST** /api/v1/conditional-access/evaluate | evaluate_conditional_access access_request
[**accessRequestGet**](AccessRequestApi.md#accessRequestGet) | **GET** /api/v1/access-requests/{id} | get access_request
[**accessRequestList**](AccessRequestApi.md#accessRequestList) | **GET** /api/v1/access-requests | list access_request
[**accessRequestProcess**](AccessRequestApi.md#accessRequestProcess) | **POST** /api/v1/access-requests/{id}/process | process access_request
[**accessRequestRevoke**](AccessRequestApi.md#accessRequestRevoke) | **DELETE** /api/v1/access-requests/{id} | revoke access_request
[**accessRequestStats**](AccessRequestApi.md#accessRequestStats) | **GET** /api/v1/access-requests/stats | stats access_request
[**accessRequestUserBehaviorAnalytics**](AccessRequestApi.md#accessRequestUserBehaviorAnalytics) | **GET** /api/v1/analytics/users/{user_id}/behavior | user_behavior_analytics access_request
[**accessRequestUserInsights**](AccessRequestApi.md#accessRequestUserInsights) | **GET** /api/v1/analytics/users/{user_id}/insights | user_insights access_request
[**accessRequestUserRiskAssessment**](AccessRequestApi.md#accessRequestUserRiskAssessment) | **GET** /api/v1/analytics/users/{user_id}/risk | user_risk_assessment access_request



## accessRequestCreate

create access_request

Create a new access request

### Example

```bash
 accessRequestCreate
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createAccessRequestPayload** | [**CreateAccessRequestPayload**](CreateAccessRequestPayload.md) |  |

### Return type

[**AccessRequestResult**](AccessRequestResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestCreateConditionalRule

create_conditional_rule access_request

Create a conditional access rule

### Example

```bash
 accessRequestCreateConditionalRule
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createConditionalRulePayload** | [**CreateConditionalRulePayload**](CreateConditionalRulePayload.md) |  |

### Return type

[**ConditionalRuleResult**](ConditionalRuleResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestDetectAnomalies

detect_anomalies access_request

Detect user behavior anomalies

### Example

```bash
 accessRequestDetectAnomalies user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]

### Return type

[**AnomalyDetectionResult**](AnomalyDetectionResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestEvaluateConditionalAccess

evaluate_conditional_access access_request

Evaluate conditional access rules

### Example

```bash
 accessRequestEvaluateConditionalAccess
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **conditionalAccessPayload** | [**ConditionalAccessPayload**](ConditionalAccessPayload.md) |  |

### Return type

[**ConditionalAccessResult**](ConditionalAccessResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestGet

get access_request

Get access request by ID

### Example

```bash
 accessRequestGet id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Access request ID | [default to null]

### Return type

[**AccessRequestResult**](AccessRequestResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestList

list access_request

List access requests with filtering

### Example

```bash
 accessRequestList  status=value  requester_id=value  entity_id=value  limit=value  offset=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **status** | **string** | Filter by status | [optional] [default to null]
 **requesterId** | **string** | Filter by requester ID | [optional] [default to null]
 **entityId** | **string** | Filter by entity ID | [optional] [default to null]
 **limit** | **integer** | Number of requests to return | [optional] [default to 50]
 **offset** | **integer** | Number of requests to skip | [optional] [default to 0]

### Return type

[**AccessRequestListResult**](AccessRequestListResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestProcess

process access_request

Process an access request (approve/deny)

### Example

```bash
 accessRequestProcess id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Access request ID | [default to null]
 **processAccessRequestPayload2** | [**ProcessAccessRequestPayload2**](ProcessAccessRequestPayload2.md) |  |

### Return type

[**AccessRequestResult**](AccessRequestResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestRevoke

revoke access_request

Revoke an access request

### Example

```bash
 accessRequestRevoke id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string** | Access request ID | [default to null]

### Return type

(empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestStats

stats access_request

Get access request statistics

### Example

```bash
 accessRequestStats
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**AccessRequestStatsResult**](AccessRequestStatsResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestUserBehaviorAnalytics

user_behavior_analytics access_request

Get user behavior analytics

### Example

```bash
 accessRequestUserBehaviorAnalytics user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]

### Return type

[**UserBehaviorResult**](UserBehaviorResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestUserInsights

user_insights access_request

Get personalized user insights

### Example

```bash
 accessRequestUserInsights user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]

### Return type

[**UserInsightsResult**](UserInsightsResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## accessRequestUserRiskAssessment

user_risk_assessment access_request

Get user risk assessment

### Example

```bash
 accessRequestUserRiskAssessment user_id=value
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **string** | User ID | [default to null]

### Return type

[**UserRiskResult**](UserRiskResult.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

