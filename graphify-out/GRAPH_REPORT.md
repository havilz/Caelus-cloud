# Graph Report - Caelus-cloud  (2026-08-21)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 1429 nodes · 3944 edges · 84 communities (74 shown, 10 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 24 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `2553594d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- net/http.ResponseWriter
- Credential
- LoadConfig
- User
- dependencies
- context.Context
- github.com/jackc/pgx/v5/pgxpool.Pool
- github.com/google/uuid.UUID
- [id]/page.tsx
- monitoring/page.tsx
- AlertRule
- compilerOptions
- vps/page.tsx
- Bucket
- AutomationRule
- organizations
- TestAutomationHTTP_Endpoints
- storage.service.ts
- Server
- security/page.tsx
- automation.service.ts
- CreateServerModal.tsx
- setupTestAuthUsecase
- index.ts
- Authenticate
- DistributedScheduler
- Engine
- HeartbeatWatchdog
- LinuxCollector
- MonitoringUsecase
- Hub
- NewJWTManager
- NewCredentialUsecase
- MockStorageAdapter
- TaskPayload
- main
- StorageUsecase
- UnixSocketInspector
- ServerUsecase
- 000001_init_schema.up.sql
- pkg/config/config.go
- Adapter
- testing.T
- SystemEvent
- ProviderDriver
- ServerMetric
- RedisQueueEngine
- CredentialUsecase
- usecase
- automation.go
- Info
- CredentialRepository
- RuleExecutionLog
- time.Duration
- NewAlertEvaluator
- SecurityFinding
- NewVulnScanner
- ScanType
- Sidebar.tsx
- useAuthStore.ts
- SecurityUsecase
- ScanTarget
- Handler
- app/layout.tsx
- TelemetryReportPayload
- setupTelemetryHTTPTest
- eslint.config.mjs
- next.config.ts
- postcss.config.mjs
- schema_migrations
- schema_migrations
- github.com/havilz/caelus-cloud/agent
- github.com/havilz/caelus-cloud/backend

## God Nodes (most connected - your core abstractions)
1. `main()` - 54 edges
2. `Success()` - 46 edges
3. `GetOrganizationIDFromContext()` - 44 edges
4. `Credential` - 37 edges
5. `AutomationRule` - 25 edges
6. `Hub` - 22 edges
7. `Server` - 21 edges
8. `mockBackupRepo` - 20 edges
9. `User` - 19 edges
10. `BackupPolicy` - 19 edges

## Surprising Connections (you probably didn't know these)
- `ResizeServerModalProps` --references--> `Server`  [EXTRACTED]
  frontend/src/components/server/ResizeServerModal.tsx → frontend/src/types/server.ts
- `recordAuditEntry()` --calls--> `GetOrganizationIDFromContext()`  [INFERRED]
  backend/internal/delivery/http/middleware/audit.go → backend/internal/delivery/http/middleware/auth.go
- `setupMiddlewareTest()` --calls--> `newMockOrgRepo()`  [INFERRED]
  backend/tests/middleware_test.go → backend/tests/auth_usecase_test.go
- `recordAuditEntry()` --calls--> `GetUserIDFromContext()`  [INFERRED]
  backend/internal/delivery/http/middleware/audit.go → backend/internal/delivery/http/middleware/auth.go
- `TestBackupUsecase_PolicyAndSnapshotLifecycle()` --calls--> `newMockServerRepo()`  [INFERRED]
  backend/tests/backup_usecase_test.go → backend/tests/server_test.go

## Import Cycles
- None detected.

## Communities (84 total, 10 thin omitted)

### Community 0 - "net/http.ResponseWriter"
Cohesion: 0.06
Nodes (45): SetAuditMetadata(), GetOrganizationIDFromContext(), resolveOrganizationID(), Error(), JSON(), Paginated(), Success(), NewRouter() (+37 more)

### Community 1 - "Credential"
Cohesion: 0.13
Nodes (8): CreateServerRequest, Credential, ProviderServer, ResizeServerRequest, CustomDriver, MockDriver, ServerStatus, mockCredRepo

### Community 2 - "LoadConfig"
Cohesion: 0.07
Nodes (30): main(), runCollectionCycle(), Collector, NewCollector(), getEnvAsBoolOrDefault(), getEnvAsIntOrDefault(), getEnvOrDefault(), Config (+22 more)

### Community 3 - "User"
Cohesion: 0.08
Nodes (18): authUsecase, LoginInput, LoginOutput, RegisterInput, RegisterOutput, User, UserRepository, NewUserRepository() (+10 more)

### Community 4 - "dependencies"
Cohesion: 0.04
Nodes (45): axios, class-variance-authority, clsx, eslint, eslint-config-next, dependencies, axios, class-variance-authority (+37 more)

### Community 5 - "context.Context"
Cohesion: 0.11
Nodes (10): BackupPolicy, BackupRecord, BackupStatus, CreateBackupPolicyInput, LokiLogEntry, BackupUsecase, context.Context, time.Time (+2 more)

### Community 6 - "github.com/jackc/pgx/v5/pgxpool.Pool"
Cohesion: 0.06
Nodes (20): GetMemberFromContext(), getRoleRank(), isRoleAuthorized(), Organization, OrganizationMember, OrganizationRepository, OrganizationRole, Provider (+12 more)

### Community 7 - "github.com/google/uuid.UUID"
Cohesion: 0.10
Nodes (11): IncidentStatus, SecurityIncident, SecurityPostureOverview, SecurityScan, NewSecurityRepository(), newMockSecurityRepo(), TestSecurityHTTP_Endpoints(), ScanStatus (+3 more)

### Community 8 - "[id]/page.tsx"
Cohesion: 0.26
Nodes (17): LogViewer(), LogViewerProps, MetricCard(), MetricCardProps, MetricTimeSeriesChart(), ServerStatusBadge(), Card(), CardContent() (+9 more)

### Community 9 - "monitoring/page.tsx"
Cohesion: 0.12
Nodes (24): LoginPage(), RegisterPage(), ServerDetailPage(), DashboardLayout(), DashboardLayoutProps, MonitoringPage(), HomePage(), Breadcrumbs() (+16 more)

### Community 10 - "AlertRule"
Cohesion: 0.11
Nodes (8): Alert, AlertRule, AlertStatus, NewAlertRepository(), AlertSeverity, AlertRepository, Server, mockAlertRepo

### Community 11 - "compilerOptions"
Cohesion: 0.07
Nodes (28): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+20 more)

### Community 12 - "vps/page.tsx"
Cohesion: 0.20
Nodes (14): VPSManagementPage(), OverviewPage(), ConnectAgentModal(), ConnectAgentModalProps, CreateServerModal(), ResizeServerModal(), ResizeServerModalProps, Button() (+6 more)

### Community 13 - "Bucket"
Cohesion: 0.17
Nodes (3): Bucket, bucketRepository, mockBucketRepo

### Community 14 - "AutomationRule"
Cohesion: 0.13
Nodes (6): AutomationRule, NewAutomationRepository(), TestCentralEventDispatcher_FanOut(), TestRuleEngine_ConditionEvaluationAndActionExecution(), AutomationRepository, mockAutomationRepo

### Community 15 - "organizations"
Cohesion: 0.15
Nodes (22): alert_rules, alerts, server_metrics, update_alert_rules_updated_at, backup_policies, backup_records, buckets, update_backup_policies_modtime (+14 more)

### Community 16 - "TestAutomationHTTP_Endpoints"
Cohesion: 0.15
Nodes (14): NewUnifiedDispatcher(), BuildAlertHTMLTemplate(), Client, EmailMessage, NewClient(), Client, WebhookPayload, NewClient() (+6 more)

### Community 17 - "storage.service.ts"
Cohesion: 0.17
Nodes (16): StorageExplorerPageProps, CreateBucketModal(), CreateBucketModalProps, GenerateSignedUrlModal(), GenerateSignedUrlModalProps, UploadObjectModal(), UploadObjectModalProps, storageService (+8 more)

### Community 18 - "Server"
Cohesion: 0.14
Nodes (6): Server, ServerStatus, NewServerRepository(), ServerRepository, Provider, mockServerRepo

### Community 19 - "security/page.tsx"
Cohesion: 0.18
Nodes (17): FindingDetailModal(), FindingDetailModalProps, SecurityScoreBadge(), SecurityScoreBadgeProps, securityService, FindingCategory, FindingSeverity, FindingStatus (+9 more)

### Community 20 - "automation.service.ts"
Cohesion: 0.20
Nodes (14): CreateRuleModal(), CreateRuleModalProps, automationService, ActionResultItem, ActionType, AutomationRule, ConditionOperator, CreateRulePayload (+6 more)

### Community 21 - "CreateServerModal.tsx"
Cohesion: 0.13
Nodes (24): CreateBackupPolicyModal(), CreateBackupPolicyModalProps, CreateServerModalProps, OnboardingTab, cn(), Input, InputProps, apiClient (+16 more)

### Community 22 - "setupTestAuthUsecase"
Cohesion: 0.21
Nodes (16): Compare(), DefaultArgon2Params(), Hash(), newMockOrgRepo(), newMockUserRepo(), setupTestAuthUsecase(), TestLogin_InactiveUser(), TestLogin_RefreshToken_Success() (+8 more)

### Community 23 - "index.ts"
Cohesion: 0.23
Nodes (8): ServerStatusBadgeProps, Badge(), BadgeProps, cn(), AppColors, AppControls, AppTheme, ServerStatus

### Community 24 - "Authenticate"
Cohesion: 0.26
Nodes (15): Authenticate(), extractBearerToken(), GetClaimsFromContext(), GetUserIDFromContext(), RequireOrganizationRole(), Manager, setupMiddlewareTest(), TestAuthenticateMiddleware_InvalidFormat() (+7 more)

### Community 25 - "DistributedScheduler"
Cohesion: 0.39
Nodes (4): QueueEngine, NewDistributedScheduler(), DistributedScheduler, ScheduledJob

### Community 26 - "Engine"
Cohesion: 0.24
Nodes (10): BackupExecutor, Engine, RuleEngine, ServerExecutor, compareValues(), mapEventTypeToTriggerType(), NewEngine(), toFloat64() (+2 more)

### Community 27 - "HeartbeatWatchdog"
Cohesion: 0.31
Nodes (5): MetricRepository, NewHeartbeatWatchdog(), sync.Mutex, sync.WaitGroup, HeartbeatWatchdog

### Community 28 - "LinuxCollector"
Cohesion: 0.25
Nodes (3): NewCollector(), Collector, LinuxCollector

### Community 29 - "MonitoringUsecase"
Cohesion: 0.15
Nodes (11): AlertEvaluator, LogQueryAdapter, MetricsQueryAdapter, NewClient(), NewClient(), NewMonitoringUsecase(), TestMonitoringUsecase_AlertLifecycle(), TestMonitoringUsecase_GetMetricsHistory() (+3 more)

### Community 30 - "Hub"
Cohesion: 0.24
Nodes (6): Hub, NewClient(), NewHub(), TestWSHub_RegisterAndBroadcast(), Client, EventMessage

### Community 31 - "NewJWTManager"
Cohesion: 0.14
Nodes (18): AuditLogInterceptor(), extractClientIP(), isMutatingMethod(), recordAuditEntry(), AuditLog, AuditLogRepository, JWTConfig, NewJWTManager() (+10 more)

### Community 32 - "NewCredentialUsecase"
Cohesion: 0.37
Nodes (11): NewCredentialUsecase(), Decrypt(), Encrypt(), newMockCredRepo(), newMockProviderRepo(), TestCredentialUsecase_Create_Success(), TestCredentialUsecase_Delete_Success(), TestCredentialUsecase_Get_ForbiddenOrg() (+3 more)

### Community 33 - "MockStorageAdapter"
Cohesion: 0.18
Nodes (5): ObjectItem, sync.RWMutex, mockBucket, mockObject, MockStorageAdapter

### Community 34 - "TaskPayload"
Cohesion: 0.21
Nodes (7): TaskHandler, TaskPayload, TaskType, NewMockQueueEngine(), TestDistributedScheduler_RegisterAndTrigger(), TestMockQueueEngine_Lifecycle(), MockQueueEngine

### Community 35 - "main"
Cohesion: 0.05
Nodes (40): NewLogger(), main(), NewBackupHandler(), NewStorageHandler(), BackupRepository, BucketRepository, StorageFactory, StorageProviderType (+32 more)

### Community 36 - "StorageUsecase"
Cohesion: 0.24
Nodes (3): ObjectContent, StorageUsecase, io.ReadCloser

### Community 37 - "UnixSocketInspector"
Cohesion: 0.28
Nodes (4): UnixSocketInspector, net/http.Client, Client, Client

### Community 38 - "ServerUsecase"
Cohesion: 0.26
Nodes (6): ProviderRepository, NewServerUsecase(), validateCreateServerInput(), CreateServerInput, ResizeServerInput, ServerUsecase

### Community 39 - "000001_init_schema.up.sql"
Cohesion: 0.30
Nodes (13): audit_logs, credentials, organization_members, organizations, providers, servers, update_credentials_updated_at, update_org_members_updated_at (+5 more)

### Community 40 - "pkg/config/config.go"
Cohesion: 0.26
Nodes (12): getEnv(), getEnvAsBool(), getEnvAsDuration(), getEnvAsInt(), getEnvAsSlice(), Config, DatabaseConfig, Load() (+4 more)

### Community 41 - "Adapter"
Cohesion: 0.25
Nodes (3): github.com/aws/aws-sdk-go-v2/service/s3.Client, github.com/aws/aws-sdk-go-v2/service/s3.PresignClient, Adapter

### Community 42 - "testing.T"
Cohesion: 0.18
Nodes (21): TestCollector_CollectSuccess(), TestCollector_ConsecutiveDeltaCalculation(), CORS(), TestHealthCheckEndpoints(), createServerViaHTTP(), executeServerRequest(), setupServerHTTPTest(), setupServerUsecaseTest() (+13 more)

### Community 43 - "SystemEvent"
Cohesion: 0.38
Nodes (4): CentralEventDispatcher, EventSubscriber, NewCentralEventDispatcher(), SystemEvent

### Community 44 - "ProviderDriver"
Cohesion: 0.24
Nodes (12): ProviderDriver, NewCustomDriver(), NewDriverFactory(), NewMockDriver(), createTestServerHelper(), TestMockDriver_CreateAndGetServer(), TestMockDriver_DeleteServer(), TestMockDriver_ListServers() (+4 more)

### Community 45 - "ServerMetric"
Cohesion: 0.26
Nodes (4): ServerMetric, NewMetricRepository(), MetricRepository, mockMetricRepo

### Community 46 - "RedisQueueEngine"
Cohesion: 0.27
Nodes (4): NewRedisQueueEngine(), redis.Client, Config, RedisQueueEngine

### Community 47 - "CredentialUsecase"
Cohesion: 0.21
Nodes (4): CredentialRepository, CredentialUsecase, CreateCredentialInput, UpdateCredentialInput

### Community 48 - "usecase"
Cohesion: 0.22
Nodes (4): AutomationUsecase, usecase, EventDispatcher, NewAutomationUsecase()

### Community 49 - "automation.go"
Cohesion: 0.40
Nodes (9): ActionResultItem, ConditionOperator, CreateRuleInput, RuleAction, RuleCondition, RuleTriggerType, UpdateRuleInput, ActionType (+1 more)

### Community 50 - "Info"
Cohesion: 0.16
Nodes (12): main(), main(), registerWorkerHandlers(), RequestLogger(), NewClient(), Debug(), Error(), Get() (+4 more)

### Community 54 - "NewAlertEvaluator"
Cohesion: 0.31
Nodes (5): AlertRepository, NewAlertEvaluator(), TestAlertEvaluator_MemoryAndDiskRules(), TestAlertEvaluator_TriggerOnThresholdExceeded(), AlertEvaluator

### Community 55 - "SecurityFinding"
Cohesion: 0.53
Nodes (5): FindingCategory, FindingSeverity, FindingStatus, NormalizedFinding, SecurityFinding

### Community 56 - "NewVulnScanner"
Cohesion: 0.33
Nodes (4): NewVulnScanner(), TestVulnScanner_EvaluatesOS(), KnownVulnRule, VulnScanner

### Community 57 - "ScanType"
Cohesion: 0.17
Nodes (8): NewSecurityHandler(), ScanType, NewTLSScanner(), TLSScanner, CreateIncidentRequest, TriggerScanRequest, UpdateFindingStatusRequest, UpdateIncidentStatusRequest

### Community 58 - "Sidebar.tsx"
Cohesion: 0.50
Nodes (4): cn(), NavItem, navItems, Sidebar()

### Community 61 - "useAuthStore.ts"
Cohesion: 0.42
Nodes (7): AuthState, AuthData, LoginInput, Organization, RegisterInput, TokenPair, User

### Community 64 - "SecurityUsecase"
Cohesion: 0.16
Nodes (12): SecurityRepository, ServerRepository, NewFindingNormalizer(), Orchestrator, NewOrchestrator(), NewRiskEngine(), SecurityUsecase, NewSecurityUsecase() (+4 more)

### Community 65 - "ScanTarget"
Cohesion: 0.10
Nodes (14): ScanTarget, NewHeadersScanner(), NewHostConfigScanner(), NewPortScanner(), TestFindingNormalizer_GeneratesFingerprint(), TestHeadersScanner_AuditsHeaders(), TestHostConfigScanner_EvaluatesTelemetry(), TestPortScanner_ScanLocalhost() (+6 more)

### Community 67 - "Handler"
Cohesion: 0.43
Nodes (4): Handler, NewHandler(), Client, github.com/gorilla/websocket.Conn

### Community 70 - "app/layout.tsx"
Cohesion: 0.33
Nodes (4): geistMono, geistSans, metadata, RootLayoutProps

### Community 71 - "TelemetryReportPayload"
Cohesion: 0.60
Nodes (3): TelemetryReportPayload, ContainerMetricPayload, HostMetricsPayload

### Community 73 - "setupTelemetryHTTPTest"
Cohesion: 0.60
Nodes (4): setupTelemetryHTTPTest(), TestAlertHTTP_ListAndAcknowledge(), TestTelemetryHTTP_IngestAndQuery(), mockServerRepo

## Knowledge Gaps
- **92 isolated node(s):** `ConnectAgentModalProps`, `CreateServerModalProps`, `OnboardingTab`, `DialogProps`, `StorageExplorerPageProps` (+87 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `main` to `net/http.ResponseWriter`, `User`, `github.com/jackc/pgx/v5/pgxpool.Pool`, `github.com/google/uuid.UUID`, `AlertRule`, `AutomationRule`, `TestAutomationHTTP_Endpoints`, `Server`, `Engine`, `HeartbeatWatchdog`, `MonitoringUsecase`, `Hub`, `NewJWTManager`, `NewCredentialUsecase`, `ServerUsecase`, `pkg/config/config.go`, `SystemEvent`, `ProviderDriver`, `ServerMetric`, `usecase`, `Info`, `CredentialRepository`, `NewAlertEvaluator`, `ScanType`, `SecurityUsecase`, `Handler`?**
  _High betweenness centrality (0.072) - this node is a cross-community bridge._
- **Why does `GetOrganizationIDFromContext()` connect `net/http.ResponseWriter` to `Authenticate`, `github.com/google/uuid.UUID`, `context.Context`, `NewJWTManager`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Why does `setupTelemetryHTTPTest()` connect `setupTelemetryHTTPTest` to `net/http.ResponseWriter`, `Handler`, `github.com/google/uuid.UUID`, `testing.T`, `AlertRule`, `ServerMetric`, `NewAlertEvaluator`, `Authenticate`, `MonitoringUsecase`, `Hub`, `NewJWTManager`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **What connects `ConnectAgentModalProps`, `CreateServerModalProps`, `OnboardingTab` to the rest of the system?**
  _92 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `net/http.ResponseWriter` be split into smaller, more focused modules?**
  _Cohesion score 0.06400987480162229 - nodes in this community are weakly interconnected._
- **Should `Credential` be split into smaller, more focused modules?**
  _Cohesion score 0.12903225806451613 - nodes in this community are weakly interconnected._
- **Should `LoadConfig` be split into smaller, more focused modules?**
  _Cohesion score 0.06871035940803383 - nodes in this community are weakly interconnected._