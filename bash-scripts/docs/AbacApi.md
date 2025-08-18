# AbacApi

All URIs are relative to **

Method | HTTP request | Description
------------- | ------------- | -------------
[**abacAuditDecisions**](AbacApi.md#abacAuditDecisions) | **GET** /abac/audit | audit_decisions abac
[**abacAuthorize**](AbacApi.md#abacAuthorize) | **POST** /abac/authorize | authorize abac
[**abacCollectAttributes**](AbacApi.md#abacCollectAttributes) | **POST** /abac/collect-attributes | collect_attributes abac
[**abacDiscoverPolicies**](AbacApi.md#abacDiscoverPolicies) | **POST** /abac/discover-policies | discover_policies abac
[**abacEvaluate**](AbacApi.md#abacEvaluate) | **POST** /abac/evaluate | evaluate abac
[**abacEvaluateBulk**](AbacApi.md#abacEvaluateBulk) | **POST** /abac/evaluate-bulk | evaluate_bulk abac
[**abacExplain**](AbacApi.md#abacExplain) | **POST** /abac/explain | explain abac
[**abacHealth**](AbacApi.md#abacHealth) | **GET** /abac/health | health abac
[**abacInvalidateCache**](AbacApi.md#abacInvalidateCache) | **POST** /abac/invalidate-cache | invalidate_cache abac
[**abacMetrics**](AbacApi.md#abacMetrics) | **GET** /abac/metrics | metrics abac



## abacAuditDecisions

audit_decisions abac

Get a history of policy decisions.

### Example

```bash
 abacAuditDecisions
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **auditDecisionsPayload** | [**AuditDecisionsPayload**](AuditDecisionsPayload.md) |  |

### Return type

[**DecisionAuditResponse**](DecisionAuditResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacAuthorize

authorize abac

Simple authorization check.

### Example

```bash
 abacAuthorize
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorizationRequest** | [**AuthorizationRequest**](AuthorizationRequest.md) |  |

### Return type

[**AuthorizationResponse**](AuthorizationResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacCollectAttributes

collect_attributes abac

Collect attributes for a given context.

### Example

```bash
 abacCollectAttributes
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **attributeCollectionRequest** | [**AttributeCollectionRequest**](AttributeCollectionRequest.md) |  |

### Return type

[**AttributeCollectionResponse**](AttributeCollectionResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacDiscoverPolicies

discover_policies abac

Discover applicable policies.

### Example

```bash
 abacDiscoverPolicies
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **policyDiscoveryRequest** | [**PolicyDiscoveryRequest**](PolicyDiscoveryRequest.md) |  |

### Return type

[**PolicyDiscoveryResponse**](PolicyDiscoveryResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacEvaluate

evaluate abac

Evaluate a policy decision request.

### Example

```bash
 abacEvaluate
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **policyEvaluationRequest** | [**PolicyEvaluationRequest**](PolicyEvaluationRequest.md) |  |

### Return type

[**PolicyEvaluationResponse**](PolicyEvaluationResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacEvaluateBulk

evaluate_bulk abac

Evaluate a bulk policy decision request.

### Example

```bash
 abacEvaluateBulk
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **bulkPolicyEvaluationRequest** | [**BulkPolicyEvaluationRequest**](BulkPolicyEvaluationRequest.md) |  |

### Return type

[**BulkPolicyEvaluationResponse**](BulkPolicyEvaluationResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacExplain

explain abac

Explain a policy decision.

### Example

```bash
 abacExplain
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **policyExplanationRequest** | [**PolicyExplanationRequest**](PolicyExplanationRequest.md) |  |

### Return type

[**PolicyExplanationResponse**](PolicyExplanationResponse.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacHealth

health abac

Health check for the ABAC service.

### Example

```bash
 abacHealth
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HealthResult**](HealthResult.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacInvalidateCache

invalidate_cache abac

Invalidate the ABAC cache.

### Example

```bash
 abacInvalidateCache
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **invalidateCachePayload** | [**InvalidateCachePayload**](InvalidateCachePayload.md) |  |

### Return type

[**InvalidateCacheResult**](InvalidateCacheResult.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## abacMetrics

metrics abac

Get performance metrics for the ABAC service.

### Example

```bash
 abacMetrics
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**MetricsResult**](MetricsResult.md)

### Authorization

[jwt_header_Authorization](../README.md#jwt_header_Authorization)

### HTTP request headers

- **Content-Type**: Not Applicable
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

