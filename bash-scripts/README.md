# Enterprise AWO ERP System API Bash client

## Overview

This is a Bash client script for accessing Enterprise AWO ERP System API service.

The script uses cURL underneath for making all REST calls.

## Usage

```shell
# Make sure the script has executable rights
$ chmod u+x 

# Print the list of operations available on the service
$ ./ -h

# Print the service description
$ ./ --about

# Print detailed information about specific operation
$ ./ <operationId> -h

# Make GET request
./ --host http://<hostname>:<port> --accept xml <operationId> <queryParam1>=<value1> <header_key1>:<header_value2>

# Make GET request using arbitrary curl options (must be passed before <operationId>) to an SSL service using username:password
 -k -sS --tlsv1.2 --host https://<hostname> -u <user>:<password> --accept xml <operationId> <queryParam1>=<value1> <header_key1>:<header_value2>

# Make POST request
$ echo '<body_content>' |  --host <hostname> --content-type json <operationId> -

# Make POST request with simple JSON content, e.g.:
# {
#   "key1": "value1",
#   "key2": "value2",
#   "key3": 23
# }
$ echo '<body_content>' |  --host <hostname> --content-type json <operationId> key1==value1 key2=value2 key3:=23 -

# Make POST request with form data
$  --host <hostname> <operationId> key1:=value1 key2:=value2 key3:=23

# Preview the cURL command without actually executing it
$  --host http://<hostname>:<port> --dry-run <operationid>

```

## Docker image

You can easily create a Docker image containing a preconfigured environment
for using the REST Bash client including working autocompletion and short
welcome message with basic instructions, using the generated Dockerfile:

```shell
docker build -t my-rest-client .
docker run -it my-rest-client
```

By default you will be logged into a Zsh environment which has much more
advanced auto completion, but you can switch to Bash, where basic autocompletion
is also available.

## Shell completion

### Bash

The generated bash-completion script can be either directly loaded to the current Bash session using:

```shell
source .bash-completion
```

Alternatively, the script can be copied to the `/etc/bash-completion.d` (or on OSX with Homebrew to `/usr/local/etc/bash-completion.d`):

```shell
sudo cp .bash-completion /etc/bash-completion.d/
```

#### OS X

On OSX you might need to install bash-completion using Homebrew:

```shell
brew install bash-completion
```

and add the following to the `~/.bashrc`:

```shell
if [ -f $(brew --prefix)/etc/bash_completion ]; then
  . $(brew --prefix)/etc/bash_completion
fi
```

### Zsh

In Zsh, the generated `_` Zsh completion file must be copied to one of the folders under `$FPATH` variable.

## Documentation for API Endpoints

All URIs are relative to **

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AbacApi* | [**abacAuditDecisions**](docs/AbacApi.md#abacauditdecisions) | **GET** /abac/audit | audit_decisions abac
*AbacApi* | [**abacAuthorize**](docs/AbacApi.md#abacauthorize) | **POST** /abac/authorize | authorize abac
*AbacApi* | [**abacCollectAttributes**](docs/AbacApi.md#abaccollectattributes) | **POST** /abac/collect-attributes | collect_attributes abac
*AbacApi* | [**abacDiscoverPolicies**](docs/AbacApi.md#abacdiscoverpolicies) | **POST** /abac/discover-policies | discover_policies abac
*AbacApi* | [**abacEvaluate**](docs/AbacApi.md#abacevaluate) | **POST** /abac/evaluate | evaluate abac
*AbacApi* | [**abacEvaluateBulk**](docs/AbacApi.md#abacevaluatebulk) | **POST** /abac/evaluate-bulk | evaluate_bulk abac
*AbacApi* | [**abacExplain**](docs/AbacApi.md#abacexplain) | **POST** /abac/explain | explain abac
*AbacApi* | [**abacHealth**](docs/AbacApi.md#abachealth) | **GET** /abac/health | health abac
*AbacApi* | [**abacInvalidateCache**](docs/AbacApi.md#abacinvalidatecache) | **POST** /abac/invalidate-cache | invalidate_cache abac
*AbacApi* | [**abacMetrics**](docs/AbacApi.md#abacmetrics) | **GET** /abac/metrics | metrics abac
*AccessRequestApi* | [**accessRequestCreate**](docs/AccessRequestApi.md#accessrequestcreate) | **POST** /api/v1/access-requests | create access_request
*AccessRequestApi* | [**accessRequestCreateConditionalRule**](docs/AccessRequestApi.md#accessrequestcreateconditionalrule) | **POST** /api/v1/conditional-access/rules | create_conditional_rule access_request
*AccessRequestApi* | [**accessRequestDetectAnomalies**](docs/AccessRequestApi.md#accessrequestdetectanomalies) | **POST** /api/v1/analytics/users/{user_id}/detect-anomalies | detect_anomalies access_request
*AccessRequestApi* | [**accessRequestEvaluateConditionalAccess**](docs/AccessRequestApi.md#accessrequestevaluateconditionalaccess) | **POST** /api/v1/conditional-access/evaluate | evaluate_conditional_access access_request
*AccessRequestApi* | [**accessRequestGet**](docs/AccessRequestApi.md#accessrequestget) | **GET** /api/v1/access-requests/{id} | get access_request
*AccessRequestApi* | [**accessRequestList**](docs/AccessRequestApi.md#accessrequestlist) | **GET** /api/v1/access-requests | list access_request
*AccessRequestApi* | [**accessRequestProcess**](docs/AccessRequestApi.md#accessrequestprocess) | **POST** /api/v1/access-requests/{id}/process | process access_request
*AccessRequestApi* | [**accessRequestRevoke**](docs/AccessRequestApi.md#accessrequestrevoke) | **DELETE** /api/v1/access-requests/{id} | revoke access_request
*AccessRequestApi* | [**accessRequestStats**](docs/AccessRequestApi.md#accessrequeststats) | **GET** /api/v1/access-requests/stats | stats access_request
*AccessRequestApi* | [**accessRequestUserBehaviorAnalytics**](docs/AccessRequestApi.md#accessrequestuserbehavioranalytics) | **GET** /api/v1/analytics/users/{user_id}/behavior | user_behavior_analytics access_request
*AccessRequestApi* | [**accessRequestUserInsights**](docs/AccessRequestApi.md#accessrequestuserinsights) | **GET** /api/v1/analytics/users/{user_id}/insights | user_insights access_request
*AccessRequestApi* | [**accessRequestUserRiskAssessment**](docs/AccessRequestApi.md#accessrequestuserriskassessment) | **GET** /api/v1/analytics/users/{user_id}/risk | user_risk_assessment access_request
*AdminFeatureflagApi* | [**adminFeatureflagBulkDisable**](docs/AdminFeatureflagApi.md#adminfeatureflagbulkdisable) | **POST** /api/v1/admin/feature-flags/bulk/disable | bulk_disable admin-featureflag
*AdminFeatureflagApi* | [**adminFeatureflagBulkEnable**](docs/AdminFeatureflagApi.md#adminfeatureflagbulkenable) | **POST** /api/v1/admin/feature-flags/bulk/enable | bulk_enable admin-featureflag
*AdminFeatureflagApi* | [**adminFeatureflagSystemHealth**](docs/AdminFeatureflagApi.md#adminfeatureflagsystemhealth) | **GET** /api/v1/admin/feature-flags/health | system_health admin-featureflag
*AuthApi* | [**authLogin**](docs/AuthApi.md#authlogin) | **POST** /api/v1/auth/login | login auth
*AuthApi* | [**authLogout**](docs/AuthApi.md#authlogout) | **POST** /api/v1/auth/logout | logout auth
*AuthApi* | [**authRefresh**](docs/AuthApi.md#authrefresh) | **POST** /api/v1/auth/refresh | refresh auth
*AuthApi* | [**authValidate**](docs/AuthApi.md#authvalidate) | **GET** /api/v1/auth/validate | validate auth
*FeatureflagApi* | [**featureflagCreate**](docs/FeatureflagApi.md#featureflagcreate) | **POST** /api/v1/feature-flags | create featureflag
*FeatureflagApi* | [**featureflagDelete**](docs/FeatureflagApi.md#featureflagdelete) | **DELETE** /api/v1/feature-flags/{id} | delete featureflag
*FeatureflagApi* | [**featureflagEvaluate**](docs/FeatureflagApi.md#featureflagevaluate) | **POST** /api/v1/feature-flags/evaluate/{name} | evaluate featureflag
*FeatureflagApi* | [**featureflagEvaluateMultiple**](docs/FeatureflagApi.md#featureflagevaluatemultiple) | **POST** /api/v1/feature-flags/evaluate | evaluateMultiple featureflag
*FeatureflagApi* | [**featureflagGet**](docs/FeatureflagApi.md#featureflagget) | **GET** /api/v1/feature-flags/{name} | get featureflag
*FeatureflagApi* | [**featureflagGetById**](docs/FeatureflagApi.md#featureflaggetbyid) | **GET** /api/v1/feature-flags/by-id/{id} | getById featureflag
*FeatureflagApi* | [**featureflagGetByType**](docs/FeatureflagApi.md#featureflaggetbytype) | **GET** /api/v1/feature-flags/type/{flag_type} | getByType featureflag
*FeatureflagApi* | [**featureflagGetStats**](docs/FeatureflagApi.md#featureflaggetstats) | **GET** /api/v1/feature-flags/stats | getStats featureflag
*FeatureflagApi* | [**featureflagHealth**](docs/FeatureflagApi.md#featureflaghealth) | **GET** /api/v1/feature-flags/health | health featureflag
*FeatureflagApi* | [**featureflagList**](docs/FeatureflagApi.md#featureflaglist) | **GET** /api/v1/feature-flags | list featureflag
*FeatureflagApi* | [**featureflagSearch**](docs/FeatureflagApi.md#featureflagsearch) | **GET** /api/v1/feature-flags/search | search featureflag
*FeatureflagApi* | [**featureflagUpdate**](docs/FeatureflagApi.md#featureflagupdate) | **PUT** /api/v1/feature-flags/{id} | update featureflag
*HealthApi* | [**healthHealth**](docs/HealthApi.md#healthhealth) | **GET** /health | health health
*HealthApi* | [**healthReady**](docs/HealthApi.md#healthready) | **GET** /ready | ready health
*OpenapiApi* | [**openapiSpec**](docs/OpenapiApi.md#openapispec) | **GET** /openapi.json | spec openapi
*OpenapiApi* | [**openapiUi**](docs/OpenapiApi.md#openapiui) | **GET** /swagger-ui | ui openapi
*OrganizationApi* | [**organizationArchive**](docs/OrganizationApi.md#organizationarchive) | **PATCH** /api/v1/organizations/{id}/archive | archive organization
*OrganizationApi* | [**organizationCreate**](docs/OrganizationApi.md#organizationcreate) | **POST** /api/v1/organizations | create organization
*OrganizationApi* | [**organizationGet**](docs/OrganizationApi.md#organizationget) | **GET** /api/v1/organizations/{id} | get organization
*OrganizationApi* | [**organizationHierarchy**](docs/OrganizationApi.md#organizationhierarchy) | **GET** /api/v1/organizations/{id}/hierarchy | hierarchy organization
*OrganizationApi* | [**organizationList**](docs/OrganizationApi.md#organizationlist) | **GET** /api/v1/organizations | list organization
*OrganizationApi* | [**organizationUpdate**](docs/OrganizationApi.md#organizationupdate) | **PUT** /api/v1/organizations/{id} | update organization
*TenantApi* | [**tenantCreate**](docs/TenantApi.md#tenantcreate) | **POST** /api/v1/tenants | create tenant
*TenantApi* | [**tenantDelete**](docs/TenantApi.md#tenantdelete) | **DELETE** /api/v1/tenants/{id} | delete tenant
*TenantApi* | [**tenantGet**](docs/TenantApi.md#tenantget) | **GET** /api/v1/tenants/{id} | get tenant
*TenantApi* | [**tenantHealth**](docs/TenantApi.md#tenanthealth) | **GET** /api/v1/tenants/health | health tenant
*TenantApi* | [**tenantList**](docs/TenantApi.md#tenantlist) | **GET** /api/v1/tenants | list tenant
*TenantApi* | [**tenantUpdate**](docs/TenantApi.md#tenantupdate) | **PUT** /api/v1/tenants/{id} | update tenant
*UserApi* | [**userAssignRole**](docs/UserApi.md#userassignrole) | **POST** /api/v1/users/{user_id}/roles | assign_role user
*UserApi* | [**userAuthorizeAction**](docs/UserApi.md#userauthorizeaction) | **POST** /api/v1/users/{user_id}/authorize | authorize_action user
*UserApi* | [**userBulkUpdateAttributes**](docs/UserApi.md#userbulkupdateattributes) | **POST** /api/v1/users/bulk-attributes | bulk_update_attributes user
*UserApi* | [**userCheckPermission**](docs/UserApi.md#usercheckpermission) | **POST** /api/v1/users/{user_id}/check-permission | check_permission user
*UserApi* | [**userCreate**](docs/UserApi.md#usercreate) | **POST** /api/v1/users | create user
*UserApi* | [**userDeactivate**](docs/UserApi.md#userdeactivate) | **PATCH** /api/v1/users/{id}/deactivate | deactivate user
*UserApi* | [**userGet**](docs/UserApi.md#userget) | **GET** /api/v1/users/{id} | get user
*UserApi* | [**userGetAttributes**](docs/UserApi.md#usergetattributes) | **GET** /api/v1/users/{id}/attributes | get_attributes user
*UserApi* | [**userGetSessionAttributes**](docs/UserApi.md#usergetsessionattributes) | **GET** /api/v1/users/{user_id}/sessions/{session_id}/attributes | get_session_attributes user
*UserApi* | [**userGetUserContext**](docs/UserApi.md#usergetusercontext) | **GET** /api/v1/users/{id}/context | get_user_context user
*UserApi* | [**userList**](docs/UserApi.md#userlist) | **GET** /api/v1/users | list user
*UserApi* | [**userPermissions**](docs/UserApi.md#userpermissions) | **GET** /api/v1/users/{id}/permissions | permissions user
*UserApi* | [**userRefreshAttributes**](docs/UserApi.md#userrefreshattributes) | **POST** /api/v1/users/{id}/refresh-attributes | refresh_attributes user
*UserApi* | [**userRemoveRole**](docs/UserApi.md#userremoverole) | **DELETE** /api/v1/users/{user_id}/roles/{role_id} | remove_role user
*UserApi* | [**userSetAttributes**](docs/UserApi.md#usersetattributes) | **PUT** /api/v1/users/{id}/attributes | set_attributes user
*UserApi* | [**userSetSessionContext**](docs/UserApi.md#usersetsessioncontext) | **PUT** /api/v1/users/{user_id}/sessions/{session_id}/context | set_session_context user
*UserApi* | [**userUpdate**](docs/UserApi.md#userupdate) | **PUT** /api/v1/users/{id} | update user
*UserApi* | [**userValidateAttributes**](docs/UserApi.md#uservalidateattributes) | **POST** /api/v1/users/{id}/validate-attributes | validate_attributes user


## Documentation For Models

 - [ABACComponentHealth](docs/ABACComponentHealth.md)
 - [AccessRequestListResult](docs/AccessRequestListResult.md)
 - [AccessRequestResult](docs/AccessRequestResult.md)
 - [AccessRequestStatsResult](docs/AccessRequestStatsResult.md)
 - [ActivityStats](docs/ActivityStats.md)
 - [ActivitySummary](docs/ActivitySummary.md)
 - [AddressInfo](docs/AddressInfo.md)
 - [AnomalyDetectionResult](docs/AnomalyDetectionResult.md)
 - [AnomalyResult](docs/AnomalyResult.md)
 - [AssignRolePayload](docs/AssignRolePayload.md)
 - [AssignRolePayload2](docs/AssignRolePayload2.md)
 - [AttributeCollectionRequest](docs/AttributeCollectionRequest.md)
 - [AttributeCollectionResponse](docs/AttributeCollectionResponse.md)
 - [AttributeMetadata](docs/AttributeMetadata.md)
 - [AttributeMetrics](docs/AttributeMetrics.md)
 - [AttributeValidationDetail](docs/AttributeValidationDetail.md)
 - [AttributeValidationResult](docs/AttributeValidationResult.md)
 - [AttributeValue](docs/AttributeValue.md)
 - [AttributesSummary](docs/AttributesSummary.md)
 - [AuditDecisionsPayload](docs/AuditDecisionsPayload.md)
 - [Auth](docs/Auth.md)
 - [AuthorizationRequest](docs/AuthorizationRequest.md)
 - [AuthorizationResponse](docs/AuthorizationResponse.md)
 - [AuthorizationResult](docs/AuthorizationResult.md)
 - [AuthorizeActionPayload](docs/AuthorizeActionPayload.md)
 - [AuthorizeActionPayload2](docs/AuthorizeActionPayload2.md)
 - [BehavioralAnalysis](docs/BehavioralAnalysis.md)
 - [BulkEnableRequestBody](docs/BulkEnableRequestBody.md)
 - [BulkEnableResponseBody](docs/BulkEnableResponseBody.md)
 - [BulkEvaluatePayload](docs/BulkEvaluatePayload.md)
 - [BulkPerformanceMetrics](docs/BulkPerformanceMetrics.md)
 - [BulkPolicyEvaluationRequest](docs/BulkPolicyEvaluationRequest.md)
 - [BulkPolicyEvaluationResponse](docs/BulkPolicyEvaluationResponse.md)
 - [BulkUpdateOptions](docs/BulkUpdateOptions.md)
 - [BulkUpdateResult](docs/BulkUpdateResult.md)
 - [BulkUpdateSummary](docs/BulkUpdateSummary.md)
 - [BulkUpdateUserAttributesPayload](docs/BulkUpdateUserAttributesPayload.md)
 - [BulkUpdateUserAttributesResult](docs/BulkUpdateUserAttributesResult.md)
 - [BusinessHours](docs/BusinessHours.md)
 - [CacheEvent](docs/CacheEvent.md)
 - [CacheImpact](docs/CacheImpact.md)
 - [CacheInfo](docs/CacheInfo.md)
 - [CacheMetrics](docs/CacheMetrics.md)
 - [CheckPermissionPayload](docs/CheckPermissionPayload.md)
 - [CheckPermissionPayload2](docs/CheckPermissionPayload2.md)
 - [ClientInfo](docs/ClientInfo.md)
 - [ComplianceExemption](docs/ComplianceExemption.md)
 - [ComplianceStatus](docs/ComplianceStatus.md)
 - [ComplianceValidation](docs/ComplianceValidation.md)
 - [ComplianceViolation](docs/ComplianceViolation.md)
 - [ConditionalAccessContext](docs/ConditionalAccessContext.md)
 - [ConditionalAccessPayload](docs/ConditionalAccessPayload.md)
 - [ConditionalAccessResult](docs/ConditionalAccessResult.md)
 - [ConditionalRuleResult](docs/ConditionalRuleResult.md)
 - [ConflictResolutionSummary](docs/ConflictResolutionSummary.md)
 - [ContactInfo](docs/ContactInfo.md)
 - [CreateAccessRequestPayload](docs/CreateAccessRequestPayload.md)
 - [CreateConditionalRulePayload](docs/CreateConditionalRulePayload.md)
 - [CreateFeatureFlagPayload](docs/CreateFeatureFlagPayload.md)
 - [CreateOrganizationPayload](docs/CreateOrganizationPayload.md)
 - [CreateTenantPayload](docs/CreateTenantPayload.md)
 - [CreateUserPayload](docs/CreateUserPayload.md)
 - [DayHours](docs/DayHours.md)
 - [DecisionAuditEntry](docs/DecisionAuditEntry.md)
 - [DecisionAuditResponse](docs/DecisionAuditResponse.md)
 - [DecisionAuditTrail](docs/DecisionAuditTrail.md)
 - [DownstreamImpact](docs/DownstreamImpact.md)
 - [EnvironmentContext](docs/EnvironmentContext.md)
 - [ErpError](docs/ErpError.md)
 - [ErpError2](docs/ErpError2.md)
 - [Error](docs/Error.md)
 - [EvaluateFeatureFlagPayload](docs/EvaluateFeatureFlagPayload.md)
 - [EvaluateFeatureFlagPayload2](docs/EvaluateFeatureFlagPayload2.md)
 - [EvaluateMultipleResponseBody](docs/EvaluateMultipleResponseBody.md)
 - [EvaluationContext](docs/EvaluationContext.md)
 - [EvaluationMetrics](docs/EvaluationMetrics.md)
 - [EvaluationResult](docs/EvaluationResult.md)
 - [EvaluationStep](docs/EvaluationStep.md)
 - [EvaluationTiming](docs/EvaluationTiming.md)
 - [FeatureFlagActivitySummary](docs/FeatureFlagActivitySummary.md)
 - [FeatureFlagStats](docs/FeatureFlagStats.md)
 - [Featureflag](docs/Featureflag.md)
 - [FreshnessStatus](docs/FreshnessStatus.md)
 - [GetByTypeResponseBody](docs/GetByTypeResponseBody.md)
 - [HealthChecks](docs/HealthChecks.md)
 - [HealthResponseBody](docs/HealthResponseBody.md)
 - [HealthResponseBody2](docs/HealthResponseBody2.md)
 - [HealthResult](docs/HealthResult.md)
 - [HealthStatus](docs/HealthStatus.md)
 - [InvalidateCachePayload](docs/InvalidateCachePayload.md)
 - [InvalidateCacheResult](docs/InvalidateCacheResult.md)
 - [ListAccessRequestsPayload](docs/ListAccessRequestsPayload.md)
 - [ListResponseBody](docs/ListResponseBody.md)
 - [ListResponseBody2](docs/ListResponseBody2.md)
 - [ListResponseBody3](docs/ListResponseBody3.md)
 - [ListResponseBody4](docs/ListResponseBody4.md)
 - [LocationContext](docs/LocationContext.md)
 - [LoginRequestBody](docs/LoginRequestBody.md)
 - [MetricsResult](docs/MetricsResult.md)
 - [NotificationPreferences](docs/NotificationPreferences.md)
 - [Organization](docs/Organization.md)
 - [OrganizationHierarchy](docs/OrganizationHierarchy.md)
 - [OrganizationNode](docs/OrganizationNode.md)
 - [OrganizationSettings](docs/OrganizationSettings.md)
 - [Pagination](docs/Pagination.md)
 - [PaginationMeta](docs/PaginationMeta.md)
 - [PerformanceInfo](docs/PerformanceInfo.md)
 - [PermissionCheckResult](docs/PermissionCheckResult.md)
 - [PermissionInfo](docs/PermissionInfo.md)
 - [PolicyAdvice](docs/PolicyAdvice.md)
 - [PolicyDiscoveryRequest](docs/PolicyDiscoveryRequest.md)
 - [PolicyDiscoveryResponse](docs/PolicyDiscoveryResponse.md)
 - [PolicyEvaluationRequest](docs/PolicyEvaluationRequest.md)
 - [PolicyEvaluationResponse](docs/PolicyEvaluationResponse.md)
 - [PolicyEvaluationSummary](docs/PolicyEvaluationSummary.md)
 - [PolicyExplanation](docs/PolicyExplanation.md)
 - [PolicyExplanationRequest](docs/PolicyExplanationRequest.md)
 - [PolicyExplanationResponse](docs/PolicyExplanationResponse.md)
 - [PolicyObligation](docs/PolicyObligation.md)
 - [PolicySummary](docs/PolicySummary.md)
 - [ProcessAccessRequestPayload](docs/ProcessAccessRequestPayload.md)
 - [ProcessAccessRequestPayload2](docs/ProcessAccessRequestPayload2.md)
 - [PropagationInfo](docs/PropagationInfo.md)
 - [ReadinessStatus](docs/ReadinessStatus.md)
 - [RefreshAttributesRequestBody](docs/RefreshAttributesRequestBody.md)
 - [RefreshAttributesResult](docs/RefreshAttributesResult.md)
 - [RefreshError](docs/RefreshError.md)
 - [RefreshRequestBody](docs/RefreshRequestBody.md)
 - [ResultMetadata](docs/ResultMetadata.md)
 - [RiskAssessment](docs/RiskAssessment.md)
 - [RoleInfo](docs/RoleInfo.md)
 - [SearchResponseBody](docs/SearchResponseBody.md)
 - [SecurityAssessment](docs/SecurityAssessment.md)
 - [SecurityContext](docs/SecurityContext.md)
 - [SessionAttributesResult](docs/SessionAttributesResult.md)
 - [SessionContext](docs/SessionContext.md)
 - [SessionPerformanceMetrics](docs/SessionPerformanceMetrics.md)
 - [SessionRiskAssessment](docs/SessionRiskAssessment.md)
 - [SetAttributeMetadata](docs/SetAttributeMetadata.md)
 - [SetAttributeOptions](docs/SetAttributeOptions.md)
 - [SetSessionContextPayload](docs/SetSessionContextPayload.md)
 - [SetSessionContextPayload2](docs/SetSessionContextPayload2.md)
 - [SetSessionContextResult](docs/SetSessionContextResult.md)
 - [SetUserAttributesPayload](docs/SetUserAttributesPayload.md)
 - [SetUserAttributesPayload2](docs/SetUserAttributesPayload2.md)
 - [SetUserAttributesResult](docs/SetUserAttributesResult.md)
 - [SourceRefreshResult](docs/SourceRefreshResult.md)
 - [SpecResponseBody](docs/SpecResponseBody.md)
 - [SubscriptionInfo](docs/SubscriptionInfo.md)
 - [SystemContext](docs/SystemContext.md)
 - [SystemHealthResponseBody](docs/SystemHealthResponseBody.md)
 - [Tenant](docs/Tenant.md)
 - [TenantInfo](docs/TenantInfo.md)
 - [TenantLimits](docs/TenantLimits.md)
 - [TenantSettings](docs/TenantSettings.md)
 - [TimeContext](docs/TimeContext.md)
 - [TokenValidation](docs/TokenValidation.md)
 - [UpdateFeatureFlagPayload](docs/UpdateFeatureFlagPayload.md)
 - [UpdateFeatureFlagPayload2](docs/UpdateFeatureFlagPayload2.md)
 - [UpdateOrganizationPayload](docs/UpdateOrganizationPayload.md)
 - [UpdateOrganizationPayload2](docs/UpdateOrganizationPayload2.md)
 - [UpdateTenantPayload](docs/UpdateTenantPayload.md)
 - [UpdateTenantPayload2](docs/UpdateTenantPayload2.md)
 - [UpdateUserPayload](docs/UpdateUserPayload.md)
 - [UpdateUserPayload2](docs/UpdateUserPayload2.md)
 - [User](docs/User.md)
 - [UserAccessPatterns](docs/UserAccessPatterns.md)
 - [UserAttributeUpdate](docs/UserAttributeUpdate.md)
 - [UserAttributesResult](docs/UserAttributesResult.md)
 - [UserBehaviorResult](docs/UserBehaviorResult.md)
 - [UserContextResult](docs/UserContextResult.md)
 - [UserInfo](docs/UserInfo.md)
 - [UserInsightsResult](docs/UserInsightsResult.md)
 - [UserMonitoringInfo](docs/UserMonitoringInfo.md)
 - [UserPermissions](docs/UserPermissions.md)
 - [UserPolicyAdvice](docs/UserPolicyAdvice.md)
 - [UserPolicyExplanation](docs/UserPolicyExplanation.md)
 - [UserPolicyObligation](docs/UserPolicyObligation.md)
 - [UserPreferences](docs/UserPreferences.md)
 - [UserProfile](docs/UserProfile.md)
 - [UserRiskFactor](docs/UserRiskFactor.md)
 - [UserRiskProfile](docs/UserRiskProfile.md)
 - [UserRiskResult](docs/UserRiskResult.md)
 - [UserValidationResults](docs/UserValidationResults.md)
 - [ValidateAttributesRequestBody](docs/ValidateAttributesRequestBody.md)
 - [ValidationMessage](docs/ValidationMessage.md)


## Documentation For Authorization


## jwt_header_Authorization


- **Type**: HTTP Bearer Token authentication

