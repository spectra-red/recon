# Integration Point Mapping Skill

This skill enables agents to identify and document how a new feature or change will integrate with existing systems, services, and code within the repository.

## Objective

Map all the connection points, dependencies, and data flows between the new feature and existing systems to ensure proper integration and identify potential impacts.

## Input Required

- **Feature/Task Description**: What is being built or changed
- **Technical Scope**: Which systems are involved
- **Architecture Context**: How the system is currently organized

## Research Process

### 1. Identify Integration Categories

**Types of integrations to map:**
- **Code-level**: Function calls, class instantiation, module imports
- **Data-level**: Database access, shared state, data transformations
- **API-level**: REST endpoints, GraphQL, gRPC services
- **Event-level**: Event emitters, listeners, message queues
- **Infrastructure-level**: Config files, environment variables, deployment
- **UI-level**: Component composition, routing, state management

### 2. Map Incoming Integration Points

**Where existing code will call the new code:**
- Which modules/services will import the new functionality
- Which endpoints will route to new handlers
- Which events will trigger new listeners
- Which UI components will use new components

### 3. Map Outgoing Integration Points

**What the new code will need to call:**
- Existing APIs and services to invoke
- Database tables/collections to access
- External services to integrate with
- Shared utilities and helpers to use
- Configuration to read

### 4. Trace Data Flow

**Follow data through the system:**
- Where data originates (user input, external API, database)
- How data is transformed at each step
- Where data is stored or persisted
- Where data is sent (response, event, external service)
- What data validation occurs at each point

### 5. Identify Side Effects

**Changes that will affect other parts of the system:**
- Modified shared state
- Database schema changes affecting other features
- API contract changes affecting consumers
- Configuration changes affecting other services
- Breaking changes requiring updates elsewhere

### 6. Map Authentication & Authorization Flow

**How auth integrates:**
- Where auth checks happen
- What permissions are required
- How user context is passed through the system
- Session management integration

### 7. Identify Testing Integration Points

**How new code integrates with testing:**
- Test utilities to use
- Mocking strategies for dependencies
- Test data setup and teardown
- Integration test scenarios

## Output Format

```markdown
## Integration Point Mapping

### System Context

**Architecture Style**: [Monolith / Microservices / Serverless / etc.]

**Communication Patterns**: [Synchronous / Asynchronous / Event-driven / etc.]

**Current Flow Diagram**:
```
[ASCII or markdown diagram showing current system]
```

---

### Integration Overview

**New Feature Integration Points**: [Number] points identified

**Impact Level**: [Low / Medium / High]

**Breaking Changes**: [Yes / No]

---

### Incoming Integration Points

*Where existing code will call or use the new feature*

#### [Integration Point 1]

**Caller**: `path/to/calling/code.ext:line`

**Integration Type**: [Direct Function Call / API Request / Event Subscription / etc.]

**How It Integrates**:
```[language]
// Example of how existing code will call new code
import { newFeature } from './new-feature'

existingFunction() {
  const result = newFeature(params)
  // use result
}
```

**Changes Required in Caller**:
- [ ] Import new module
- [ ] Update function signature
- [ ] Handle new return type
- [ ] Add error handling

**Data Passed In**:
- Parameter 1: Type, description
- Parameter 2: Type, description

**Data Received Back**:
- Return value: Type, description

**Error Handling**:
- How errors should be handled
- What errors can be thrown

---

#### [Integration Point 2]

**Trigger**: `path/to/trigger.ext:line`

**Integration Type**: [Event / Webhook / Schedule / etc.]

**When It's Called**: [Condition or event that triggers it]

**Data Flow**:
```
Trigger Event → Event Bus → New Handler → Process → Response/Side Effect
```

**Example**:
```[language]
// How the integration works
eventBus.on('user.created', newUserHandler)
```

---

### Outgoing Integration Points

*What the new feature needs to call or access*

#### [Integration Point 1]

**Service/API Called**: [Service Name]

**Location**: `path/to/service.ext:line`

**Integration Type**: [Direct Call / HTTP Request / Database Query / etc.]

**Purpose**: Why the new feature needs this

**How to Integrate**:
```[language]
// Example of calling existing service
import { existingService } from './existing-service'

newFeature() {
  const data = await existingService.getData(params)
  // use data
}
```

**Data Sent**:
- Parameter: Type, description

**Data Received**:
- Return: Type, description

**Error Scenarios**:
- Error 1: How to handle
- Error 2: How to handle

**Dependencies**:
- Requires: [What needs to be initialized first]
- Assumes: [What state must exist]

---

#### Database Integration

**Table/Collection**: `table_name`

**Schema Location**: `path/to/schema.ext:line`

**Access Pattern**: [Read / Write / Read-Write]

**Queries Needed**:

##### Query 1: Fetch User Data
```sql
SELECT id, name, email FROM users WHERE id = ?
```
- **Purpose**: Get user information
- **Used by**: `path/to/code:line`

##### Query 2: Update User Status
```sql
UPDATE users SET status = ? WHERE id = ?
```
- **Purpose**: Update user status
- **Used by**: `path/to/code:line`

**Transaction Requirements**:
- Needs transaction: Yes/No
- Isolation level: [If applicable]

**Impact on Existing Queries**:
- Query X may be slower due to: [Reason]
- Index needed: [Description]

---

#### External Service Integration

**Service**: [Service Name, e.g., Stripe, SendGrid, AWS S3]

**Purpose**: What functionality it provides

**Authentication**: [API Key / OAuth / etc.]

**Configuration Required**:
```bash
# Environment variables needed
EXTERNAL_SERVICE_API_KEY=xxx
EXTERNAL_SERVICE_ENDPOINT=https://...
```

**API Calls**:

##### API Call 1
- **Endpoint**: `POST /api/endpoint`
- **Purpose**: What it does
- **Request**:
  ```json
  {
    "field": "value"
  }
  ```
- **Response**:
  ```json
  {
    "result": "value"
  }
  ```

**Error Handling**:
- Rate limits: [How to handle]
- Network errors: [Retry strategy]
- Invalid responses: [How to handle]

**Cost Implications**: [If applicable]

---

### Data Flow Mapping

#### End-to-End Flow

```
User Request
    ↓
[1] API Endpoint (path/to/handler.ext:line)
    ↓ validates request
[2] Service Layer (path/to/service.ext:line)
    ↓ business logic
[3] Repository (path/to/repo.ext:line)
    ↓ database query
[4] Database
    ↑ returns data
[5] Service Layer
    ↓ transforms data
[6] Event Emitter (path/to/event.ext:line)
    ↓ triggers event
[7] Response to User
```

**Step-by-Step Detail**:

##### Step 1: API Endpoint
- **Location**: `path/to/handler.ext:line`
- **Input**: Request object
- **Processing**: Validates input, extracts parameters
- **Output**: Validated parameters
- **Integration**: Calls service layer

##### Step 2: Service Layer
- **Location**: `path/to/service.ext:line`
- **Input**: Validated parameters
- **Processing**: Business logic, orchestration
- **Output**: Processed data
- **Integration**: Calls repository, external services

##### Step 3: Repository
- **Location**: `path/to/repo.ext:line`
- **Input**: Query parameters
- **Processing**: Database operations
- **Output**: Raw data
- **Integration**: Database connection

##### Step 4: Event Emission
- **Location**: `path/to/event.ext:line`
- **Event**: `event.name`
- **Payload**: Event data structure
- **Subscribers**: Who listens to this event

---

### State Management Integration

#### Shared State

**State Location**: `path/to/state.ext:line`

**State Modified**:
```[language]
interface StateChanges {
  field1: NewType  // What changes
  field2: NewType  // What changes
}
```

**Who Else Reads This State**:
- Component 1: `path/to/component.ext:line`
- Component 2: `path/to/component.ext:line`

**Potential Conflicts**:
- Race condition: [Description and mitigation]
- Stale data: [Description and mitigation]

**State Update Pattern**:
```[language]
// How state should be updated
dispatch({ type: 'ACTION', payload: data })
```

---

### Configuration Integration

#### Config Files to Modify

##### [Config File 1]

**Location**: `path/to/config.ext`

**Changes Needed**:
```[language]
// New configuration
{
  newFeature: {
    enabled: true,
    setting1: "value",
    setting2: 123
  }
}
```

**Used By**:
- `path/to/code1.ext:line`
- `path/to/code2.ext:line`

##### Environment Variables

**New Variables**:
```bash
NEW_FEATURE_ENABLED=true
NEW_FEATURE_API_KEY=xxx
```

**Where Set**:
- Development: `.env.local`
- Staging: `config/staging.env`
- Production: `config/production.env`

**Where Used**:
- `path/to/code.ext:line`

---

### UI/Frontend Integration

#### Component Integration

**New Component**: `NewComponent`

**Used By**:
- Page: `path/to/page.ext:line`
- Parent Component: `path/to/parent.ext:line`

**Props Interface**:
```[language]
interface NewComponentProps {
  prop1: Type1
  prop2: Type2
  onAction: (data: Type3) => void
}
```

**State Integration**:
```[language]
// How it connects to state management
const data = useSelector(selectData)
const dispatch = useDispatch()
```

#### Routing Integration

**New Routes**:
```[language]
{
  path: '/new-feature',
  component: NewFeatureComponent,
  guard: AuthGuard
}
```

**Route File**: `path/to/routes.ext:line`

**Navigation**:
- Links from: `path/to/component1.ext`, `path/to/component2.ext`

---

### Event-Driven Integration

#### Events Emitted

##### Event: `feature.action.completed`

**Emitted By**: `path/to/emitter.ext:line`

**When**: [Trigger condition]

**Payload**:
```[language]
{
  eventType: 'feature.action.completed',
  timestamp: Date,
  data: {
    field1: value,
    field2: value
  }
}
```

**Current Subscribers**: None / List if any

**Recommended Subscribers**: [Who should listen]

#### Events Subscribed To

##### Event: `user.updated`

**Subscribed By**: `path/to/subscriber.ext:line`

**Handler**:
```[language]
function handleUserUpdate(event) {
  // Process event
}
```

**Purpose**: Why the new feature needs this event

**Publisher**: `path/to/publisher.ext:line`

---

### Authentication & Authorization Integration

#### Auth Check Integration

**Where Auth Happens**: `path/to/auth/middleware.ext:line`

**How to Add Auth to New Endpoints**:
```[language]
// Example
router.post('/new-endpoint', authMiddleware, handler)
```

**Required Permissions**:
- Permission 1: Description
- Permission 2: Description

**Permission Check**:
```[language]
// How to check permissions
if (!user.hasPermission('feature.action')) {
  throw new ForbiddenError()
}
```

**Token/Session Access**:
```[language]
// How to get user from request
const user = req.user // from auth middleware
```

---

### Testing Integration Points

#### Unit Test Integration

**Test Utilities to Use**:
- `testHelper1` from `path/to/helper.ext`
- `mockFactory` from `path/to/factory.ext`

**Mocking Strategy**:
```[language]
// How to mock dependencies
jest.mock('./dependency', () => ({
  method: jest.fn().mockResolvedValue(mockData)
}))
```

**Test Data**:
- Fixtures: `path/to/fixtures.ext`
- Factories: `path/to/factories.ext`

#### Integration Test Integration

**Test Scenarios**:
1. **End-to-end flow**: Description
   - Setup: `path/to/test.ext:line`
   - Pattern to follow: `path/to/similar-test.ext:line`

**Database Test Integration**:
```[language]
// How tests handle database
beforeEach(async () => {
  await setupTestDatabase()
})

afterEach(async () => {
  await cleanupTestDatabase()
})
```

---

### Deployment Integration

#### Build Process

**Build Steps Affected**:
- Step 1: Description
- Step 2: Description

**Build Config**: `path/to/build-config.ext`

**Changes Needed**:
```[language]
// Build configuration updates
```

#### CI/CD Integration

**Pipeline File**: `path/to/ci-config.yml`

**Changes Needed**:
- [ ] Add new test step
- [ ] Add new environment variables
- [ ] Update deployment script

#### Infrastructure

**Services Affected**:
- Service 1: How it's affected
- Service 2: How it's affected

**Scaling Considerations**:
- Will this increase load on [service]?
- Need to scale [resource]?

---

### Migration & Backward Compatibility

#### Breaking Changes

**Change 1**: Description
- **Affected**: What code is affected
- **Migration**: How to update
- **Location**: `path/to/affected-code.ext:line`

#### Data Migration

**Required**: Yes/No

**Migration Script**: `path/to/migration.ext`

**Steps**:
1. Step 1: Description
2. Step 2: Description

**Rollback Plan**: How to rollback if needed

#### Feature Flags

**Flag Name**: `enable-new-feature`

**Location**: `path/to/flags.ext`

**Usage**:
```[language]
if (featureFlags.isEnabled('enable-new-feature')) {
  // New code path
} else {
  // Old code path
}
```

---

### Side Effects & Impact Analysis

#### Systems Affected

| System/Service | Impact Level | Description | Action Needed |
|----------------|--------------|-------------|---------------|
| System 1 | High | Description | Update required |
| System 2 | Medium | Description | Testing needed |
| System 3 | Low | Description | Monitor |

#### Performance Impact

**Database**:
- Query X will be affected: [How]
- Need new index: [Where]
- Expected load increase: [Estimate]

**API**:
- Endpoint X response time: [Impact]
- Expected traffic increase: [Estimate]

**Cache**:
- Cache invalidation needed: [Where]
- New cache keys: [What]

#### Monitoring Integration

**Metrics to Track**:
- Metric 1: Description, threshold
- Metric 2: Description, threshold

**Alerts to Add**:
- Alert 1: Condition, severity
- Alert 2: Condition, severity

**Logging Integration**:
```[language]
// How to add logging
logger.info('feature.action', { context: data })
```

**Log Aggregation**: Where logs will appear

---

### Integration Checklist

#### Before Implementation
- [ ] All integration points identified
- [ ] Data flow mapped end-to-end
- [ ] Dependencies reviewed and approved
- [ ] Breaking changes documented
- [ ] Migration plan created (if needed)
- [ ] Feature flag strategy defined (if needed)

#### During Implementation
- [ ] Each integration point tested
- [ ] Error handling added for all external calls
- [ ] Configuration updated
- [ ] Tests cover integration scenarios
- [ ] Logging added at key points

#### Before Deployment
- [ ] All affected teams notified
- [ ] Documentation updated
- [ ] Monitoring and alerts configured
- [ ] Rollback plan ready
- [ ] Performance testing completed

---

### Recommendations

1. **Integration Approach**: [Recommended strategy]
2. **Risk Mitigation**: [How to reduce integration risks]
3. **Testing Strategy**: [How to test integrations]
4. **Rollout Plan**: [Phased rollout or all-at-once]
5. **Monitoring**: [What to monitor closely]

---

### Risks & Concerns

1. **[Risk 1]**: Description
   - Mitigation: Strategy
   - Owner: Who should handle this

2. **[Risk 2]**: Description
   - Mitigation: Strategy
   - Owner: Who should handle this

---

### References

- Architecture docs: `path/to/docs`
- Related features: Links
- External integration docs: URLs
```

## Search Strategies

### Find Callers
```bash
# Find where existing code might call new feature
Grep: "import.*NewFeature"
Grep: "require.*new-feature"
Grep: "from.*newFeature"
```

### Find Similar Integrations
```bash
# Look for similar feature integrations
Grep: "similar-feature"
# Read how they integrate
Read: path/to/similar-feature
```

### Find Event Patterns
```bash
Grep: "emit\\("
Grep: "on\\(|addEventListener"
Grep: "publish\\(|subscribe\\("
```

### Find Config Files
```bash
Glob: **/*.{config,env}.*
Glob: config/**/*
Glob: .env*
```

### Find Auth Patterns
```bash
Grep: "middleware.*auth"
Grep: "requireAuth|isAuthenticated"
Grep: "permission|authorize"
```

## Best Practices

1. **Trace Real Flows**: Follow actual request/data paths, not just file structure
2. **Map Bidirectionally**: Both what calls new code AND what new code calls
3. **Consider Side Effects**: Think about unintended consequences
4. **Document Breaking Changes**: Be explicit about what will break
5. **Test Integration Points**: Each point needs testing
6. **Plan for Failure**: How will failures at each point be handled?
7. **Monitor Key Points**: Add observability at integration boundaries

## Tools to Use

- **Grep**: Find import statements, function calls, event handlers
- **Glob**: Locate config files, test files, related features
- **Read**: Examine integration points in detail
- **Bash**: Use `git grep` for complex searches

## Success Criteria

A successful integration point mapping provides:
- Complete list of where new code will be called from
- Complete list of what new code will call
- End-to-end data flow diagram
- Configuration changes needed
- Breaking changes identified
- Migration plan (if needed)
- Testing strategy for each integration point
- Deployment considerations
- Clear recommendations for safe integration
