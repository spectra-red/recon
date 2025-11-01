# API Context Gathering Skill

This skill enables agents to discover, document, and analyze internal APIs, services, functions, and interfaces that are relevant to implementing a new feature or task.

## Objective

Identify and document internal APIs, service interfaces, function signatures, type definitions, and communication patterns that the new feature will interact with or need to understand.

## Input Required

- **Feature/Task Description**: What is being built
- **Technology Stack**: Programming languages and frameworks
- **Integration Scope**: What internal systems/services need to be accessed

## Research Process

### 1. Identify API Types in Codebase

**Common API patterns to search for:**
- REST API endpoints (routes, controllers, handlers)
- GraphQL schemas and resolvers
- gRPC service definitions
- Internal function libraries
- Service classes/interfaces
- Database repositories
- Message queue handlers
- Event emitters/listeners
- WebSocket handlers

### 2. Discover Relevant Internal APIs

**Search strategies:**

**By naming patterns:**
- Files containing "api", "service", "handler", "controller", "repository"
- Interface/type definitions
- Route/endpoint definitions
- Function exports

**By functionality:**
- User-related APIs (if building user features)
- Data-related APIs (if accessing data)
- Auth-related APIs (if dealing with authentication)
- Payment APIs, notification APIs, etc.

### 3. Extract API Signatures

**For each relevant API, capture:**
- Function/method signatures
- Input parameters and types
- Return types
- Error handling patterns
- Authentication/authorization requirements
- Rate limiting or quotas
- Documentation comments

### 4. Analyze Usage Patterns

**Find existing usage examples:**
- How other parts of the code call these APIs
- What parameters they pass
- How they handle responses
- Error handling approaches
- Retry logic
- Timeout configurations

### 5. Document Communication Patterns

**Understand:**
- How services communicate (HTTP, gRPC, message queue, direct calls)
- Data serialization formats (JSON, Protocol Buffers, etc.)
- Authentication mechanisms (API keys, JWT, OAuth, etc.)
- Request/response structures
- Error response formats

## Output Format

```markdown
## API Context Analysis

### Internal APIs Identified

#### [API/Service Name]

**Type**: REST API / GraphQL / gRPC / Function Library / Service Class

**Location**: `path/to/api/file.ext:line`

**Purpose**: What this API does

**Signature:**
```[language]
// Function/method signature with types
function apiName(param1: Type1, param2: Type2): ReturnType
```

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| param1 | Type1 | Yes | Description |
| param2 | Type2 | No | Description |

**Returns:**
- **Success**: Description of return value
- **Error**: Error types/codes returned

**Authentication:**
- Required: Yes/No
- Method: [JWT, API Key, Session, etc.]
- Permissions needed: [List permissions]

**Usage Example:**
```[language]
// Real example from codebase
const result = await apiName(param1, param2)
// path/to/example.ext:line
```

**Error Handling:**
```[language]
// How errors are typically handled
try {
  const result = await apiName(param1, param2)
} catch (error) {
  // Error handling pattern
}
// path/to/example.ext:line
```

**Related APIs:**
- `relatedApi1()` - `path/to/file:line` - Description
- `relatedApi2()` - `path/to/file:line` - Description

---

### REST API Endpoints

#### [Endpoint Description]

**Method & Path**: `POST /api/v1/resource`

**Location**: `path/to/route/definition.ext:line`

**Handler**: `path/to/handler.ext:line`

**Request:**
```json
{
  "field1": "string",
  "field2": 123,
  "field3": {
    "nested": "object"
  }
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "field": "value"
  }
}
```

**Status Codes:**
- `200` - Success
- `400` - Bad Request - Invalid parameters
- `401` - Unauthorized - Missing/invalid auth
- `404` - Not Found - Resource doesn't exist
- `500` - Server Error - Internal error

**Headers:**
- `Authorization: Bearer <token>` - Required
- `Content-Type: application/json` - Required

**Rate Limiting:**
- Limit: [X requests per minute]
- Header: `X-RateLimit-Remaining`

**Example Usage:**
```bash
curl -X POST https://api.example.com/api/v1/resource \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{"field1": "value"}'
```

**Codebase Usage Example:**
```[language]
// How it's called from the codebase
// path/to/caller.ext:line
```

---

### GraphQL APIs

#### [Query/Mutation Name]

**Type**: Query / Mutation / Subscription

**Location**: `path/to/schema.graphql:line`

**Resolver**: `path/to/resolver.ext:line`

**Schema:**
```graphql
type Query {
  apiName(param1: String!, param2: Int): ResponseType
}

type ResponseType {
  field1: String!
  field2: Int
  field3: [NestedType!]!
}
```

**Arguments:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| param1 | String | Yes | Description |
| param2 | Int | No | Description |

**Example Query:**
```graphql
query {
  apiName(param1: "value", param2: 123) {
    field1
    field2
    field3 {
      nestedField
    }
  }
}
```

**Usage Example:**
```[language]
// How it's called from the codebase
// path/to/caller.ext:line
```

---

### Service Interfaces

#### [Service Name]

**Location**: `path/to/service.ext:line`

**Interface/Type:**
```[language]
interface ServiceName {
  method1(param: Type): Promise<ReturnType>
  method2(param: Type): ReturnType
}
```

**Implementation**: `path/to/implementation.ext:line`

**Initialization:**
```[language]
// How the service is initialized
// path/to/initialization.ext:line
```

**Dependency Injection:**
- Dependencies required: [List]
- How it's injected: [Pattern]

**Usage Example:**
```[language]
// Real usage from codebase
// path/to/example.ext:line
```

---

### Database APIs/Repositories

#### [Repository/Model Name]

**Location**: `path/to/repository.ext:line`

**Database**: [PostgreSQL, MongoDB, MySQL, etc.]

**Schema/Table**: `table_name` - `path/to/schema.ext:line`

**Available Methods:**

##### `findById(id: string): Promise<Entity>`
- **Purpose**: Fetch entity by ID
- **Location**: `path/to/file.ext:line`
- **Example**: `path/to/usage.ext:line`

##### `create(data: CreateDTO): Promise<Entity>`
- **Purpose**: Create new entity
- **Location**: `path/to/file.ext:line`
- **Example**: `path/to/usage.ext:line`

##### `update(id: string, data: UpdateDTO): Promise<Entity>`
- **Purpose**: Update existing entity
- **Location**: `path/to/file.ext:line`
- **Example**: `path/to/usage.ext:line`

**Query Examples:**
```sql
-- Common queries used
SELECT * FROM table_name WHERE condition
```

**Transaction Patterns:**
```[language]
// How transactions are handled
// path/to/transaction-example.ext:line
```

---

### Type Definitions & Interfaces

#### [Type Name]

**Location**: `path/to/types.ext:line`

**Definition:**
```[language]
interface TypeName {
  field1: string
  field2: number
  field3?: OptionalType
  field4: ComplexType
}
```

**Used By:**
- `api1()` - `path/to/file:line`
- `api2()` - `path/to/file:line`

**Validation:**
```[language]
// Validation logic if any
// path/to/validation.ext:line
```

---

### Communication Patterns

#### [Pattern Name]

**Type**: [Synchronous/Asynchronous, HTTP/gRPC/Message Queue/etc.]

**Description**: How services communicate

**Example Flow:**
```
ServiceA ---HTTP POST---> ServiceB
                          |
                          v
                        Database
                          |
                          v
ServiceA <--Response----- ServiceB
```

**Implementation:**
```[language]
// Code showing the pattern
// path/to/implementation.ext:line
```

**Error Handling:**
- Retry logic: [Pattern]
- Timeout: [Duration]
- Circuit breaker: [Yes/No]

---

### Authentication & Authorization

**Authentication Method**: [JWT, Session, API Key, OAuth, etc.]

**Location**: `path/to/auth/middleware.ext:line`

**How to Use:**
```[language]
// How to add auth to new endpoints
// path/to/example.ext:line
```

**Token Structure:**
```json
{
  "userId": "uuid",
  "roles": ["admin", "user"],
  "permissions": ["read", "write"]
}
```

**Permission Checks:**
```[language]
// How permissions are checked
// path/to/permission-check.ext:line
```

---

### API Middleware & Interceptors

#### [Middleware Name]

**Location**: `path/to/middleware.ext:line`

**Purpose**: What it does (logging, validation, auth, etc.)

**Applied To**: Which routes/endpoints

**Configuration:**
```[language]
// Middleware configuration
// path/to/config.ext:line
```

**Usage:**
```[language]
// How it's applied
// path/to/usage.ext:line
```

---

### Event-Driven APIs

#### [Event Name]

**Type**: Publisher / Subscriber

**Location**: `path/to/event-handler.ext:line`

**Event Payload:**
```[language]
interface EventPayload {
  eventType: string
  data: DataType
  timestamp: Date
}
```

**Publishers:**
- `path/to/publisher.ext:line`

**Subscribers:**
- `path/to/subscriber.ext:line`

**Usage Example:**
```[language]
// How events are emitted and handled
// path/to/example.ext:line
```

---

### API Documentation

**Documentation Location**:
- OpenAPI/Swagger: `path/to/openapi.yaml`
- README: `path/to/API.md`
- Inline comments: [Good/Sparse/None]

**Generated Docs**:
- Tool used: [Swagger, TypeDoc, etc.]
- URL: [If hosted]

---

### Integration Recommendations

**For the new feature, you will need to:**

1. **Call these existing APIs:**
   - `api1()` from `service1` - For [purpose]
   - `api2()` from `service2` - For [purpose]

2. **Create these new APIs:**
   - `newApi1()` - Should be placed in `path/to/location`
   - `newApi2()` - Should be placed in `path/to/location`

3. **Modify these existing APIs:**
   - `existingApi()` in `path/to/file:line` - Add [functionality]

4. **Follow these patterns:**
   - Error handling: [Pattern to follow]
   - Authentication: [How to add auth]
   - Validation: [How to validate inputs]
   - Response format: [Format to use]

---

### Testing

**API Testing Patterns:**

**Location**: `path/to/api-tests.ext:line`

**Example:**
```[language]
// How APIs are tested in this codebase
// path/to/test-example.ext:line
```

**Mocking:**
```[language]
// How external APIs are mocked
// path/to/mock-example.ext:line
```

---

### Known Issues & Considerations

1. **[Issue 1]**: Description and workaround
2. **[Issue 2]**: Description and workaround

---

### Quick Reference

| API | Type | Location | Purpose |
|-----|------|----------|---------|
| api1 | REST | path/to/file:line | Description |
| api2 | Function | path/to/file:line | Description |
| api3 | GraphQL | path/to/file:line | Description |
```

## Search Strategies

### Find API Endpoints
```bash
# REST routes
Grep: "(get|post|put|delete|patch)\\(" -i
Grep: "router\\." -i
Grep: "@(Get|Post|Put|Delete)" # Decorators

# GraphQL
Glob: **/*.{graphql,gql}
Grep: "type (Query|Mutation|Subscription)"

# gRPC
Glob: **/*.proto
```

### Find Service Definitions
```bash
Glob: **/*service*.{ts,js,go,py}
Glob: **/*handler*.{ts,js,go,py}
Grep: "class.*Service"
Grep: "interface.*Service"
```

### Find Type Definitions
```bash
Glob: **/*.d.ts
Glob: **/types/**/*
Grep: "interface.*DTO"
Grep: "type.*Request"
Grep: "type.*Response"
```

### Find Usage Examples
```bash
Grep: "import.*{ServiceName}"
Grep: "apiName\\("
```

## Best Practices

1. **Trace the Flow**: Follow requests from route → handler → service → repository
2. **Find Real Examples**: Look for actual usage, not just definitions
3. **Check Tests**: Test files often show clearest usage examples
4. **Document Types**: Include full type definitions, not just method signatures
5. **Map Dependencies**: Show what each API depends on
6. **Note Patterns**: Identify consistent patterns (error handling, auth, etc.)
7. **Include Edge Cases**: Document error scenarios and handling

## Tools to Use

- **Grep**: Find API definitions, route declarations, function signatures
  - Use `-i` for case-insensitive
  - Use `output_mode: "content"` with `-n` for line numbers
  - Use `-C 5` for context around matches

- **Glob**: Locate API-related files
  - `**/*{api,service,handler,controller}*`
  - `**/*.{graphql,proto}`

- **Read**: Examine complete files for context

- **Bash**: Use `git grep` for more complex searches

## Success Criteria

A successful API context analysis provides:
- Complete inventory of relevant internal APIs
- Function signatures with full type information
- Real usage examples from the codebase with file paths
- Authentication and authorization requirements
- Error handling patterns
- Communication patterns between services
- Clear recommendations for what APIs to use/create/modify
- Testing patterns for APIs
