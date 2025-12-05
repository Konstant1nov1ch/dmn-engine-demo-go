# DMN Engine Go — Диаграммы (Mermaid)

## 1. ER-диаграмма базы данных

```mermaid
erDiagram
    dmn_tenants ||--o{ dmn_definitions : "has many"
    dmn_tenants ||--o{ dmn_evaluation_history : "has many"
    dmn_definitions ||--o{ dmn_evaluation_history : "evaluated as"
    dmn_deployments ||--o{ dmn_deployment_definitions : "contains"
    dmn_definitions ||--o{ dmn_deployment_definitions : "deployed in"

    dmn_tenants {
        varchar(64) id PK
        varchar(255) name
        jsonb settings
        int rate_limit
        boolean is_active
        timestamptz created_at
    }

    dmn_definitions {
        uuid id PK
        varchar(255) key
        int version
        varchar(255) name
        text source
        jsonb parsed_model
        varchar(64) checksum
        varchar(64) tenant_id FK
        timestamptz created_at
        varchar(255) created_by
    }

    dmn_evaluation_history {
        uuid id PK
        uuid definition_id FK
        varchar(255) definition_key
        int definition_ver
        jsonb inputs
        jsonb outputs
        text[] matched_rules
        bigint duration_ns
        timestamptz evaluated_at
        varchar(64) tenant_id FK
        varchar(255) correlation_id
        varchar(20) status
        text error_message
    }

    dmn_deployments {
        uuid id PK
        varchar(255) name
        varchar(64) tenant_id
        timestamptz deployed_at
        varchar(255) deployed_by
        varchar(50) source_type
        varchar(500) source_ref
    }

    dmn_deployment_definitions {
        uuid id PK
        uuid deployment_id FK
        uuid definition_id FK
    }
```

## 2. Диаграмма компонентов

```mermaid
graph TB
    subgraph "External"
        Client[Client Applications]
        Admin[Admin/DevOps]
    end

    subgraph "API Layer"
        REST[REST API Controller]
        MW[Middleware<br/>Auth, Logging, Metrics]
    end

    subgraph "Core Layer"
        DefSvc[Definition Service]
        EvalEngine[Evaluation Engine]
        HistSvc[History Service]
        
        subgraph "Decision Processing"
            HitPolicy[Hit Policy Engine]
            DRG[DRG Resolver]
        end
    end

    subgraph "Domain Layer"
        Parser[DMN Parser]
        Model[DMN Model]
        Validator[DMN Validator]
    end

    subgraph "Infrastructure Layer"
        FEELClient[FEEL Client]
        Repo[Repository]
        Cache[Cache Manager]
        Metrics[Metrics/Tracing]
    end

    subgraph "External Services"
        FEEL[FEEL Engine<br/>Scala/Rust]
        PG[(PostgreSQL)]
        Redis[(Redis)]
        Prom[Prometheus]
    end

    Client --> REST
    Admin --> REST
    REST --> MW
    MW --> DefSvc
    MW --> EvalEngine
    MW --> HistSvc

    DefSvc --> Parser
    DefSvc --> Validator
    DefSvc --> Repo
    DefSvc --> Cache

    EvalEngine --> HitPolicy
    EvalEngine --> DRG
    EvalEngine --> FEELClient
    EvalEngine --> Cache

    Parser --> Model
    Validator --> Model

    HistSvc --> Repo

    Repo --> PG
    Cache --> Redis
    FEELClient --> FEEL
    Metrics --> Prom

    style REST fill:#e1f5fe
    style EvalEngine fill:#fff3e0
    style Parser fill:#f3e5f5
    style PG fill:#e8f5e9
    style Redis fill:#ffebee
```

## 3. Диаграмма классов (Domain Model)

```mermaid
classDiagram
    class Definitions {
        +string id
        +string name
        +string namespace
        +[]Decision decisions
        +[]InputData inputData
        +GetDecision(id) Decision
        +Validate() []ValidationError
    }

    class Decision {
        +string id
        +string name
        +Variable variable
        +[]InformationRequirement requirements
        +DecisionTable decisionTable
        +LiteralExpression literalExpression
        +GetDependencies() []string
    }

    class DecisionTable {
        +string id
        +HitPolicy hitPolicy
        +Aggregation aggregation
        +[]Input inputs
        +[]Output outputs
        +[]Rule rules
        +GetInputCount() int
        +GetRuleCount() int
    }

    class Rule {
        +string id
        +string description
        +[]InputEntry inputEntries
        +[]OutputEntry outputEntries
        +Matches(inputs, evaluator) bool
    }

    class Input {
        +string id
        +string label
        +InputExpression expression
    }

    class Output {
        +string id
        +string name
        +string typeRef
    }

    class HitPolicy {
        <<enumeration>>
        UNIQUE
        FIRST
        PRIORITY
        ANY
        COLLECT
        RULE_ORDER
        OUTPUT_ORDER
    }

    class EvaluationEngine {
        -FEELClient feelClient
        -Repository repository
        -Cache cache
        +Evaluate(request) EvaluateResult
        +EvaluateBatch(requests) []EvaluateResult
        -evaluateDecision(decision, vars) []Output
        -buildExecutionOrder(defs, target) []string
    }

    class FEELClient {
        <<interface>>
        +Evaluate(request) EvaluateResponse
        +EvaluateUnaryTest(request) bool
        +Health() error
    }

    class HTTPClient {
        -string baseURL
        -http.Client httpClient
    }

    Definitions "1" --> "*" Decision
    Definitions "1" --> "*" InputData
    Decision "1" --> "0..1" DecisionTable
    Decision "1" --> "0..1" LiteralExpression
    DecisionTable "1" --> "*" Input
    DecisionTable "1" --> "*" Output
    DecisionTable "1" --> "*" Rule
    DecisionTable --> HitPolicy
    EvaluationEngine --> FEELClient
    HTTPClient ..|> FEELClient
```

## 4. Диаграмма развёртывания

```mermaid
graph TB
    subgraph "Internet"
        Users[Users / API Clients]
    end

    subgraph "Cloud Provider AWS/GCP"
        subgraph "Load Balancer"
            ALB[Application Load Balancer<br/>TLS Termination]
        end

        subgraph "Kubernetes Cluster"
            subgraph "Namespace: dmn-engine"
                Ingress[Ingress Controller]
                
                subgraph "DMN Engine Pods"
                    Pod1[Pod 1<br/>dmn-engine:8080]
                    Pod2[Pod 2<br/>dmn-engine:8080]
                    Pod3[Pod 3<br/>dmn-engine:8080]
                end

                subgraph "FEEL Engine Pods"
                    Feel1[Pod 1<br/>feel-engine:8090]
                    Feel2[Pod 2<br/>feel-engine:8090]
                end

                HPA[Horizontal Pod Autoscaler<br/>Min: 3, Max: 10]
            end

            subgraph "Namespace: data"
                subgraph "PostgreSQL HA"
                    PGPrimary[(Primary<br/>:5432)]
                    PGReplica[(Replica<br/>:5432)]
                end

                subgraph "Redis Cluster"
                    Redis1[(Master 1)]
                    Redis2[(Master 2)]
                    Redis3[(Master 3)]
                end
            end

            subgraph "Namespace: monitoring"
                Prometheus[Prometheus]
                Grafana[Grafana]
                Jaeger[Jaeger]
            end
        end
    end

    Users --> ALB
    ALB --> Ingress
    Ingress --> Pod1
    Ingress --> Pod2
    Ingress --> Pod3

    Pod1 --> Feel1
    Pod2 --> Feel1
    Pod3 --> Feel2

    Pod1 --> PGPrimary
    Pod2 --> PGPrimary
    Pod3 --> PGPrimary

    PGPrimary -.->|Replication| PGReplica

    Pod1 --> Redis1
    Pod2 --> Redis2
    Pod3 --> Redis3

    HPA -.->|Manages| Pod1
    HPA -.->|Manages| Pod2
    HPA -.->|Manages| Pod3

    Pod1 -.->|Metrics| Prometheus
    Pod2 -.->|Metrics| Prometheus
    Pod3 -.->|Metrics| Prometheus

    style ALB fill:#ff9800
    style Pod1 fill:#4caf50
    style Pod2 fill:#4caf50
    style Pod3 fill:#4caf50
    style Feel1 fill:#2196f3
    style Feel2 fill:#2196f3
    style PGPrimary fill:#9c27b0
    style Redis1 fill:#f44336
```

## 5. Sequence Diagram: Evaluate Decision

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant API as API Handler
    participant E as Engine
    participant Cache as Redis Cache
    participant DB as PostgreSQL
    participant FEEL as FEEL Engine

    C->>API: POST /api/v1/evaluate
    API->>E: Evaluate(request)
    
    E->>Cache: Get cached definition
    alt Cache Hit
        Cache-->>E: Definition (cached)
    else Cache Miss
        E->>DB: GetByKey(key, tenant)
        DB-->>E: Definition
        E->>Cache: Set definition (TTL: 1h)
    end

    E->>E: Build execution order (TopSort)
    
    loop For each decision in order
        loop For each rule
            E->>FEEL: Evaluate input expression
            FEEL-->>E: Input value
            
            E->>FEEL: Evaluate unary test
            FEEL-->>E: Match: true/false
            
            alt Rule matched
                E->>FEEL: Evaluate output expression
                FEEL-->>E: Output value
            end
        end
        E->>E: Apply hit policy
    end

    E-->>API: EvaluateResult
    API-->>C: 200 OK {outputs}
```

## 6. Sequence Diagram: Deploy Definition

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant API as API Handler
    participant P as Parser
    participant V as Validator
    participant R as Repository
    participant Cache as Redis Cache

    C->>API: POST /api/v1/definitions
    Note right of C: DMN XML body

    API->>P: Parse(xmlContent)
    P->>P: XML → Go structs
    P-->>API: *Definitions

    API->>V: Validate(definitions)
    V->>V: Check IDs, rules, cycles
    V-->>API: []ValidationError

    alt Validation failed
        API-->>C: 400 Bad Request {errors}
    else Validation passed
        API->>R: GetNextVersion(key, tenant)
        R-->>API: version: N

        API->>R: Deploy(definition)
        R->>R: INSERT with version N+1
        R-->>API: *Definition

        API->>Cache: Set(key, definition, TTL)
        Cache-->>API: OK

        API-->>C: 201 Created {definition}
    end
```

## 7. State Diagram: Definition Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Draft: Create
    Draft --> Validating: Submit
    Validating --> Invalid: Validation Failed
    Validating --> Deployed: Validation Passed
    Invalid --> Draft: Edit
    Deployed --> Active: First Evaluation
    Active --> Active: Evaluate
    Active --> Deprecated: New Version Deployed
    Deprecated --> Archived: TTL Expired
    Archived --> [*]: Delete

    note right of Deployed
        Version N deployed
        Ready for evaluation
    end note

    note right of Deprecated
        Version N+1 exists
        Still can be used explicitly
    end note
```

## 8. Activity Diagram: Decision Evaluation

```mermaid
flowchart TD
    A[Start] --> B[Receive Evaluate Request]
    B --> C{Definition in Cache?}
    C -->|Yes| D[Load from Cache]
    C -->|No| E[Load from Database]
    E --> F[Cache Definition]
    F --> D
    
    D --> G[Build DRG Execution Order]
    G --> H{Has Dependencies?}
    H -->|Yes| I[Evaluate Dependencies First]
    I --> J[Store Results in Context]
    J --> H
    H -->|No| K[Get Decision Table]
    
    K --> L[Initialize Matched Rules List]
    L --> M{More Rules?}
    M -->|Yes| N[Get Next Rule]
    N --> O[Evaluate Input Conditions]
    O --> P{All Conditions Match?}
    P -->|Yes| Q[Evaluate Output Expressions]
    Q --> R[Add to Matched Rules]
    R --> S{Stop on First Match?}
    S -->|Yes| T[Apply Hit Policy]
    S -->|No| M
    P -->|No| M
    M -->|No| T
    
    T --> U[Build Result]
    U --> V[Record in History]
    V --> W[Return Result]
    W --> X[End]

    style A fill:#4caf50
    style X fill:#f44336
    style T fill:#ff9800
```

## 9. C4 Context Diagram

```mermaid
C4Context
    title System Context Diagram for DMN Engine Go

    Person(analyst, "Business Analyst", "Creates and manages DMN decision models")
    Person(developer, "Developer", "Integrates decision evaluation into applications")
    Person(ops, "DevOps", "Monitors and maintains the system")

    System(dmnEngine, "DMN Engine Go", "High-performance decision engine implementing DMN 1.3")

    System_Ext(backendServices, "Backend Services", "Business applications consuming decisions")
    System_Ext(feelEngine, "FEEL Engine", "Evaluates FEEL expressions")
    System_Ext(monitoring, "Monitoring Stack", "Prometheus, Grafana, Jaeger")

    Rel(analyst, dmnEngine, "Deploys DMN models", "REST API")
    Rel(developer, dmnEngine, "Evaluates decisions", "REST API")
    Rel(ops, dmnEngine, "Monitors", "Metrics/Logs")
    
    Rel(backendServices, dmnEngine, "Evaluates decisions", "REST API")
    Rel(dmnEngine, feelEngine, "Evaluates FEEL", "HTTP")
    Rel(dmnEngine, monitoring, "Sends metrics", "Prometheus")
```

## 10. C4 Container Diagram

```mermaid
C4Container
    title Container Diagram for DMN Engine Go

    Person(client, "API Client", "External system or user")

    Container_Boundary(dmnSystem, "DMN Engine Go") {
        Container(api, "API Application", "Go, Fiber", "Handles HTTP requests, orchestrates evaluation")
        ContainerDb(cache, "Redis Cache", "Redis 7", "Caches definitions and results")
        ContainerDb(db, "PostgreSQL", "PostgreSQL 16", "Stores definitions, history")
    }

    Container_Ext(feel, "FEEL Engine", "Scala/Rust", "Evaluates FEEL expressions")

    Rel(client, api, "Uses", "HTTPS/JSON")
    Rel(api, cache, "Reads/Writes", "Redis Protocol")
    Rel(api, db, "Reads/Writes", "PostgreSQL Protocol")
    Rel(api, feel, "Evaluates expressions", "HTTP/JSON")
```

---

## Использование диаграмм

### GitHub/GitLab
Диаграммы Mermaid автоматически рендерятся в GitHub и GitLab при просмотре Markdown файлов.

### VS Code
Установите расширение "Markdown Preview Mermaid Support" для предпросмотра.

### Экспорт в PNG/SVG
Используйте [Mermaid Live Editor](https://mermaid.live/) для экспорта диаграмм в изображения.

### PlantUML альтернатива
Если нужен PlantUML формат, диаграммы можно конвертировать или пересоздать в PlantUML синтаксисе.

