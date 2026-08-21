# Graph Report - Caelus-cloud  (2026-08-21)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1272 nodes · 3471 edges · 75 communities (67 shown, 8 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 21 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d8e847a5`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- net/http.ResponseWriter
- ServerUsecase
- context.Context
- User
- dependencies
- TestStorageAndBackupHTTP_Endpoints
- mockOrgRepo
- Credential
- CreateBackupPolicyModal.tsx
- Adapter
- monitoring/page.tsx
- github.com/google/uuid.UUID
- AlertRule
- compilerOptions
- TestAutomationHTTP_Endpoints
- LinuxCollector
- [id]/page.tsx
- storage.service.ts
- Server
- automation.service.ts
- Engine
- AutomationRule
- organizations
- setupTestAuthUsecase
- vps/page.tsx
- testing.T
- main
- automation.go
- TaskPayload
- time.Duration
- MonitoringUsecase
- Authenticate
- Hub
- index.ts
- NewJWTManager
- net/http.Client
- github.com/jackc/pgx/v5/pgxpool.Pool
- 000001_init_schema.up.sql
- runCollectionCycle
- NewHTTPClient
- ServerMetric
- Provider
- RedisQueueEngine
- usecase
- UnixSocketInspector
- CentralEventDispatcher
- AuditLogInterceptor
- NewAlertEvaluator
- useAuthStore
- createTestServerHelper
- Handler
- Scheduler
- HeartbeatWatchdog
- DistributedScheduler
- app/layout.tsx
- TelemetryReportPayload
- NewMockQueueEngine
- setupTelemetryHTTPTest
- eslint.config.mjs
- next.config.ts
- postcss.config.mjs
- schema_migrations
- schema_migrations
- github.com/havilz/caelus-cloud/agent
- github.com/havilz/caelus-cloud/backend

## God Nodes (most connected - your core abstractions)
1. `main()` - 50 edges
2. `Success()` - 39 edges
3. `Credential` - 37 edges
4. `GetOrganizationIDFromContext()` - 34 edges
5. `AutomationRule` - 25 edges
6. `Hub` - 22 edges
7. `Server` - 21 edges
8. `mockBackupRepo` - 20 edges
9. `BackupPolicy` - 19 edges
10. `User` - 19 edges

## Surprising Connections (you probably didn't know these)
- `recordAuditEntry()` --calls--> `GetOrganizationIDFromContext()`  [INFERRED]
  backend/internal/delivery/http/middleware/audit.go → backend/internal/delivery/http/middleware/auth.go
- `resolveOrganizationID()` --calls--> `GetOrganizationIDFromContext()`  [INFERRED]
  backend/internal/delivery/http/middleware/rbac.go → backend/internal/delivery/http/middleware/auth.go
- `recordAuditEntry()` --calls--> `GetUserIDFromContext()`  [INFERRED]
  backend/internal/delivery/http/middleware/audit.go → backend/internal/delivery/http/middleware/auth.go
- `RequireOrganizationRole()` --calls--> `GetUserIDFromContext()`  [INFERRED]
  backend/internal/delivery/http/middleware/rbac.go → backend/internal/delivery/http/middleware/auth.go
- `setupServerHTTPTest()` --calls--> `newMockAuditRepo()`  [INFERRED]
  backend/tests/server_test.go → backend/tests/audit_test.go

## Import Cycles
- None detected.

## Communities (75 total, 8 thin omitted)

### Community 0 - "net/http.ResponseWriter"
Cohesion: 0.07
Nodes (43): SetAuditMetadata(), GetOrganizationIDFromContext(), GetUserIDFromContext(), Error(), JSON(), Paginated(), Success(), NewRouter() (+35 more)

### Community 1 - "ServerUsecase"
Cohesion: 0.07
Nodes (42): CORS(), RequestLogger(), CredentialRepository, ProviderDriver, ProviderRepository, ServerRepository, NewCustomDriver(), NewDriverFactory() (+34 more)

### Community 2 - "context.Context"
Cohesion: 0.10
Nodes (11): BackupPolicy, BackupRecord, BackupStatus, CreateBackupPolicyInput, newMockBackupRepo(), TestBackupUsecase_PolicyAndSnapshotLifecycle(), BackupUsecase, context.Context (+3 more)

### Community 3 - "User"
Cohesion: 0.08
Nodes (18): authUsecase, LoginInput, LoginOutput, RegisterInput, RegisterOutput, User, UserRepository, NewUserRepository() (+10 more)

### Community 4 - "dependencies"
Cohesion: 0.04
Nodes (45): axios, class-variance-authority, clsx, eslint, eslint-config-next, dependencies, axios, class-variance-authority (+37 more)

### Community 5 - "TestStorageAndBackupHTTP_Endpoints"
Cohesion: 0.08
Nodes (27): NewStorageHandler(), BucketRepository, StorageFactory, StorageProviderType, NewBucketRepository(), NewStorageFactory(), NewAdapter(), NewMockStorageAdapter() (+19 more)

### Community 6 - "mockOrgRepo"
Cohesion: 0.09
Nodes (13): GetMemberFromContext(), getRoleRank(), isRoleAuthorized(), RequireOrganizationRole(), resolveOrganizationID(), Organization, OrganizationMember, OrganizationRepository (+5 more)

### Community 7 - "Credential"
Cohesion: 0.08
Nodes (12): CreateServerRequest, Credential, ProviderServer, ResizeServerRequest, CredentialUsecase, CustomDriver, MockDriver, CredentialRepository (+4 more)

### Community 8 - "CreateBackupPolicyModal.tsx"
Cohesion: 0.22
Nodes (12): CreateBackupPolicyModal(), CreateBackupPolicyModalProps, backupService, serverService, APIResponse, PaginatedMeta, PaginatedResponse, BackupPolicy (+4 more)

### Community 9 - "Adapter"
Cohesion: 0.09
Nodes (12): ObjectContent, ObjectItem, SignedURLOperation, github.com/aws/aws-sdk-go-v2/service/s3.Client, github.com/aws/aws-sdk-go-v2/service/s3.PresignClient, io.ReadCloser, sync.RWMutex, mockBucket (+4 more)

### Community 10 - "monitoring/page.tsx"
Cohesion: 0.20
Nodes (15): ServerDetailPage(), MonitoringPage(), AlertDrawerProps, CreateAlertRuleModal(), CreateAlertRuleModalProps, MetricTimeSeriesChartProps, useRealtimeTelemetry(), UseRealtimeTelemetryProps (+7 more)

### Community 11 - "github.com/google/uuid.UUID"
Cohesion: 0.14
Nodes (8): NewBackupHandler(), Bucket, StorageUsecase, github.com/google/uuid.UUID, bucketRepository, mockBucketRepo, CreatePolicyRequest, TriggerBackupRequest

### Community 12 - "AlertRule"
Cohesion: 0.12
Nodes (8): Alert, AlertRule, AlertStatus, NewAlertRepository(), AlertSeverity, AlertRepository, Server, mockAlertRepo

### Community 13 - "compilerOptions"
Cohesion: 0.07
Nodes (28): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+20 more)

### Community 14 - "TestAutomationHTTP_Endpoints"
Cohesion: 0.14
Nodes (16): NewUnifiedDispatcher(), BuildAlertHTMLTemplate(), Client, EmailMessage, NewClient(), Client, WebhookPayload, NewClient() (+8 more)

### Community 15 - "LinuxCollector"
Cohesion: 0.13
Nodes (9): Collector, NewCollector(), NewCollector(), ContainerMetrics, HostMetrics, Collector, FallbackCollector, LinuxCollector (+1 more)

### Community 16 - "[id]/page.tsx"
Cohesion: 0.22
Nodes (20): LogViewer(), LogViewerProps, MetricCard(), MetricCardProps, MetricTimeSeriesChart(), ServerStatusBadge(), Button(), ButtonProps (+12 more)

### Community 17 - "storage.service.ts"
Cohesion: 0.17
Nodes (16): StorageExplorerPageProps, CreateBucketModal(), CreateBucketModalProps, GenerateSignedUrlModal(), GenerateSignedUrlModalProps, UploadObjectModal(), UploadObjectModalProps, storageService (+8 more)

### Community 18 - "Server"
Cohesion: 0.14
Nodes (6): Server, ServerStatus, NewServerRepository(), ServerRepository, Provider, mockServerRepo

### Community 19 - "automation.service.ts"
Cohesion: 0.20
Nodes (14): CreateRuleModal(), CreateRuleModalProps, automationService, ActionResultItem, ActionType, AutomationRule, ConditionOperator, CreateRulePayload (+6 more)

### Community 20 - "Engine"
Cohesion: 0.19
Nodes (13): BackupExecutor, Engine, RuleEngine, ServerExecutor, main(), registerWorkerHandlers(), compareValues(), mapEventTypeToTriggerType() (+5 more)

### Community 21 - "AutomationRule"
Cohesion: 0.16
Nodes (5): AutomationRule, RuleTriggerType, NewAutomationRepository(), AutomationRepository, mockAutomationRepo

### Community 22 - "organizations"
Cohesion: 0.19
Nodes (17): alert_rules, alerts, server_metrics, update_alert_rules_updated_at, backup_policies, backup_records, buckets, update_backup_policies_modtime (+9 more)

### Community 23 - "setupTestAuthUsecase"
Cohesion: 0.21
Nodes (16): Compare(), DefaultArgon2Params(), Hash(), newMockOrgRepo(), newMockUserRepo(), setupTestAuthUsecase(), TestLogin_InactiveUser(), TestLogin_RefreshToken_Success() (+8 more)

### Community 24 - "vps/page.tsx"
Cohesion: 0.17
Nodes (19): VPSManagementPage(), ConnectAgentModal(), ConnectAgentModalProps, CreateServerModal(), CreateServerModalProps, OnboardingTab, ResizeServerModal(), ResizeServerModalProps (+11 more)

### Community 25 - "testing.T"
Cohesion: 0.20
Nodes (14): getEnvAsBoolOrDefault(), getEnvAsIntOrDefault(), getEnvOrDefault(), Config, LoadConfig(), TestCollector_CollectSuccess(), TestCollector_ConsecutiveDeltaCalculation(), TestLoadConfig_CustomValues() (+6 more)

### Community 26 - "main"
Cohesion: 0.21
Nodes (11): main(), main(), NewClient(), Debug(), Error(), Get(), Info(), Init() (+3 more)

### Community 27 - "automation.go"
Cohesion: 0.26
Nodes (10): ActionResultItem, ConditionOperator, CreateRuleInput, ExecutionStatus, RuleAction, RuleCondition, RuleExecutionLog, UpdateRuleInput (+2 more)

### Community 28 - "TaskPayload"
Cohesion: 0.22
Nodes (6): TaskHandler, TaskPayload, TaskType, sync.Mutex, sync.WaitGroup, MockQueueEngine

### Community 29 - "time.Duration"
Cohesion: 0.24
Nodes (14): getEnv(), getEnvAsBool(), getEnvAsDuration(), getEnvAsInt(), getEnvAsSlice(), Config, DatabaseConfig, JWTConfig (+6 more)

### Community 30 - "MonitoringUsecase"
Cohesion: 0.19
Nodes (7): AlertEvaluator, NewMonitoringUsecase(), TestMonitoringUsecase_AlertLifecycle(), TestMonitoringUsecase_GetMetricsHistory(), TestMonitoringUsecase_IngestTelemetry(), newMockServerRepo(), MonitoringUsecase

### Community 31 - "Authenticate"
Cohesion: 0.28
Nodes (13): Authenticate(), extractBearerToken(), GetClaimsFromContext(), Manager, setupMiddlewareTest(), TestAuthenticateMiddleware_InvalidFormat(), TestAuthenticateMiddleware_InvalidToken(), TestAuthenticateMiddleware_MissingHeader() (+5 more)

### Community 32 - "Hub"
Cohesion: 0.24
Nodes (6): Hub, NewClient(), NewHub(), TestWSHub_RegisterAndBroadcast(), Client, EventMessage

### Community 33 - "index.ts"
Cohesion: 0.17
Nodes (11): ServerStatusBadgeProps, Badge(), BadgeProps, cn(), cn(), Input, InputProps, AppColors (+3 more)

### Community 34 - "NewJWTManager"
Cohesion: 0.25
Nodes (9): AuditLog, NewJWTManager(), newMockAuditRepo(), TestAuditInterceptor_MutatingMethod_Captured(), TestAuditInterceptor_NonMutatingMethod_Skipped(), TestAuditInterceptor_Unauthenticated_Skipped(), TestAuditRepository_CreateAndList(), TestJWTGenerationAndValidation() (+1 more)

### Community 35 - "net/http.Client"
Cohesion: 0.16
Nodes (8): LogQueryAdapter, LokiLogEntry, MetricsQueryAdapter, NewClient(), NewClient(), net/http.Client, Client, Client

### Community 36 - "github.com/jackc/pgx/v5/pgxpool.Pool"
Cohesion: 0.19
Nodes (7): NewAuditRepository(), NewBackupRepository(), NewCredentialRepository(), NewMigrator(), github.com/jackc/pgx/v5/pgxpool.Pool, AuditRepository, Migrator

### Community 37 - "000001_init_schema.up.sql"
Cohesion: 0.30
Nodes (13): audit_logs, credentials, organization_members, organizations, providers, servers, update_credentials_updated_at, update_org_members_updated_at (+5 more)

### Community 38 - "runCollectionCycle"
Cohesion: 0.23
Nodes (8): main(), runCollectionCycle(), Inspector, NewInspector(), NewLogger(), TestDockerInspector_InvalidJSONResponse(), TestDockerInspector_MockUnixSocket(), TestDockerInspector_UnavailableSocket()

### Community 39 - "NewHTTPClient"
Cohesion: 0.24
Nodes (7): Client, NewHTTPClient(), TestTransport_SendReportRetrySuccess(), TestTransport_SendReportSuccess(), TestTransport_SendReportUnauthorized(), net/http.Response, HTTPClient

### Community 40 - "ServerMetric"
Cohesion: 0.26
Nodes (4): ServerMetric, NewMetricRepository(), MetricRepository, mockMetricRepo

### Community 41 - "Provider"
Cohesion: 0.26
Nodes (4): Provider, NewProviderRepository(), ProviderRepository, mockProviderRepo

### Community 42 - "RedisQueueEngine"
Cohesion: 0.30
Nodes (4): NewRedisQueueEngine(), redis.Client, Config, RedisQueueEngine

### Community 43 - "usecase"
Cohesion: 0.22
Nodes (4): AutomationUsecase, usecase, EventDispatcher, NewAutomationUsecase()

### Community 45 - "CentralEventDispatcher"
Cohesion: 0.38
Nodes (4): CentralEventDispatcher, EventSubscriber, NewCentralEventDispatcher(), SystemEvent

### Community 46 - "AuditLogInterceptor"
Cohesion: 0.31
Nodes (8): AuditLogInterceptor(), extractClientIP(), isMutatingMethod(), recordAuditEntry(), AuditLogRepository, auditResourceKeyType, AuditResourceMetadata, statusRecorder

### Community 47 - "NewAlertEvaluator"
Cohesion: 0.31
Nodes (5): AlertRepository, NewAlertEvaluator(), TestAlertEvaluator_MemoryAndDiskRules(), TestAlertEvaluator_TriggerOnThresholdExceeded(), AlertEvaluator

### Community 48 - "useAuthStore"
Cohesion: 0.12
Nodes (21): LoginPage(), RegisterPage(), DashboardLayout(), DashboardLayoutProps, OverviewPage(), HomePage(), Breadcrumbs(), Header() (+13 more)

### Community 49 - "createTestServerHelper"
Cohesion: 0.47
Nodes (7): NewMockDriver(), createTestServerHelper(), TestMockDriver_CreateAndGetServer(), TestMockDriver_DeleteServer(), TestMockDriver_ListServers(), TestMockDriver_PowerControls(), TestMockDriver_ResizeServer()

### Community 50 - "Handler"
Cohesion: 0.43
Nodes (4): Handler, NewHandler(), Client, github.com/gorilla/websocket.Conn

### Community 51 - "Scheduler"
Cohesion: 0.43
Nodes (3): BackupRepository, NewScheduler(), Scheduler

### Community 52 - "HeartbeatWatchdog"
Cohesion: 0.43
Nodes (3): MetricRepository, NewHeartbeatWatchdog(), HeartbeatWatchdog

### Community 53 - "DistributedScheduler"
Cohesion: 0.43
Nodes (3): NewDistributedScheduler(), DistributedScheduler, ScheduledJob

### Community 54 - "app/layout.tsx"
Cohesion: 0.33
Nodes (4): geistMono, geistSans, metadata, RootLayoutProps

### Community 55 - "TelemetryReportPayload"
Cohesion: 0.60
Nodes (3): TelemetryReportPayload, ContainerMetricPayload, HostMetricsPayload

### Community 56 - "NewMockQueueEngine"
Cohesion: 0.50
Nodes (3): NewMockQueueEngine(), TestDistributedScheduler_RegisterAndTrigger(), TestMockQueueEngine_Lifecycle()

### Community 57 - "setupTelemetryHTTPTest"
Cohesion: 0.60
Nodes (4): setupTelemetryHTTPTest(), TestAlertHTTP_ListAndAcknowledge(), TestTelemetryHTTP_IngestAndQuery(), mockServerRepo

## Knowledge Gaps
- **90 isolated node(s):** `DashboardLayoutProps`, `AlertDrawerProps`, `CreateAlertRuleModalProps`, `UseRealtimeTelemetryProps`, `ContainerMetric` (+85 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `main` to `net/http.ResponseWriter`, `ServerUsecase`, `User`, `TestStorageAndBackupHTTP_Endpoints`, `mockOrgRepo`, `github.com/google/uuid.UUID`, `AlertRule`, `TestAutomationHTTP_Endpoints`, `Server`, `Engine`, `AutomationRule`, `time.Duration`, `MonitoringUsecase`, `Hub`, `NewJWTManager`, `net/http.Client`, `github.com/jackc/pgx/v5/pgxpool.Pool`, `ServerMetric`, `Provider`, `usecase`, `CentralEventDispatcher`, `NewAlertEvaluator`, `Handler`, `Scheduler`, `HeartbeatWatchdog`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `GetOrganizationIDFromContext()` connect `net/http.ResponseWriter` to `context.Context`, `mockOrgRepo`, `github.com/google/uuid.UUID`, `AuditLogInterceptor`, `Authenticate`?**
  _High betweenness centrality (0.058) - this node is a cross-community bridge._
- **Why does `NewClient()` connect `main` to `context.Context`, `Engine`, `time.Duration`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **What connects `DashboardLayoutProps`, `AlertDrawerProps`, `CreateAlertRuleModalProps` to the rest of the system?**
  _90 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `net/http.ResponseWriter` be split into smaller, more focused modules?**
  _Cohesion score 0.06898096304591265 - nodes in this community are weakly interconnected._
- **Should `ServerUsecase` be split into smaller, more focused modules?**
  _Cohesion score 0.07213114754098361 - nodes in this community are weakly interconnected._
- **Should `context.Context` be split into smaller, more focused modules?**
  _Cohesion score 0.10431372549019607 - nodes in this community are weakly interconnected._