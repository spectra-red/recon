# Requirements Analysis Skill

This skill enables agents to parse, analyze, and extract actionable technical requirements from PRDs, feature descriptions, tasks, issues, or user stories.

## Objective

Transform high-level requirements, PRDs, or feature descriptions into structured technical specifications that guide implementation and context gathering.

## Input Required

- **Requirement Source**: PRD document, GitHub issue, Linear issue, feature description, or user story
- **Context**: Any additional background information or constraints

## Analysis Process

### 1. Parse the Requirement

**Extract core information:**
- **What**: What is being built or changed
- **Why**: The business value or problem being solved
- **Who**: Target users or stakeholders
- **When**: Timeline or deadlines (if mentioned)
- **Where**: Which part of the system is affected

### 2. Identify Requirement Type

**Classify the requirement:**
- **New Feature**: Building something new from scratch
- **Enhancement**: Improving existing functionality
- **Bug Fix**: Fixing broken behavior
- **Refactoring**: Improving code structure without changing behavior
- **Performance**: Optimizing speed, memory, or resources
- **Security**: Adding or improving security measures
- **Infrastructure**: Changes to deployment, CI/CD, or infrastructure
- **Documentation**: Adding or updating documentation

### 3. Extract Functional Requirements

**What the system must do:**
- User-facing features and capabilities
- Business logic requirements
- Data processing requirements
- Integration requirements
- Validation rules
- Edge cases to handle

### 4. Extract Non-Functional Requirements

**How the system must perform:**
- **Performance**: Response times, throughput, scalability
- **Security**: Authentication, authorization, encryption, data protection
- **Reliability**: Uptime, error handling, recovery
- **Usability**: User experience, accessibility
- **Maintainability**: Code quality, testing, documentation
- **Compliance**: Legal, regulatory, or policy requirements

### 5. Identify Technical Scope

**Determine what parts of the system are affected:**
- **Frontend**: UI components, pages, user interactions
- **Backend**: APIs, services, business logic
- **Database**: Schema changes, migrations, queries
- **Infrastructure**: Deployment, configuration, scaling
- **Third-party integrations**: External APIs, services
- **Testing**: Unit, integration, e2e tests

### 6. Extract Constraints

**Limitations and boundaries:**
- Technology constraints (must use specific libraries/frameworks)
- Time constraints (deadlines)
- Resource constraints (budget, team size)
- Business constraints (compliance, legal requirements)
- Technical debt considerations
- Backward compatibility requirements

### 7. Define Success Criteria

**How to know when it's done:**
- Acceptance criteria
- Measurable outcomes
- Test scenarios
- Performance benchmarks
- User acceptance criteria

### 8. Identify Dependencies

**What this requirement depends on or affects:**
- Other features or systems
- External services
- Database schema
- Configuration changes
- Team coordination needs

### 9. Surface Ambiguities and Questions

**What needs clarification:**
- Unclear specifications
- Missing information
- Conflicting requirements
- Technical uncertainties
- Implementation choices that need decision

## Output Format

```markdown
## Requirements Analysis

### Overview

**Requirement Type**: [New Feature / Enhancement / Bug Fix / etc.]

**Summary**: [One-paragraph summary of what's being built and why]

**Business Value**: [Why this matters to users/business]

---

### Functional Requirements

#### Core Functionality

1. **[Requirement 1]**
   - Description: Detailed explanation
   - User Story: As a [user], I want to [action] so that [benefit]
   - Acceptance Criteria:
     - [ ] Criterion 1
     - [ ] Criterion 2
     - [ ] Criterion 3

2. **[Requirement 2]**
   - Description: Detailed explanation
   - User Story: As a [user], I want to [action] so that [benefit]
   - Acceptance Criteria:
     - [ ] Criterion 1
     - [ ] Criterion 2

#### Edge Cases & Special Scenarios

1. **[Edge Case 1]**: How to handle [scenario]
2. **[Edge Case 2]**: What happens when [condition]
3. **[Edge Case 3]**: Behavior for [unusual situation]

#### Data Requirements

**Input Data:**
- Field 1: Type, validation rules, constraints
- Field 2: Type, validation rules, constraints

**Output Data:**
- Field 1: Type, format, transformation needed
- Field 2: Type, format, transformation needed

**Validation Rules:**
1. Rule 1: Description
2. Rule 2: Description

---

### Non-Functional Requirements

#### Performance
- Response time: [Target, e.g., < 200ms]
- Throughput: [Target, e.g., 1000 requests/second]
- Scalability: [Requirements, e.g., support 10k concurrent users]
- Database query limits: [Constraints]

#### Security
- Authentication: [Required? Method?]
- Authorization: [Roles/permissions needed]
- Data encryption: [In transit? At rest?]
- Input validation: [Requirements]
- Rate limiting: [Needed? Limits?]
- Sensitive data handling: [Requirements]

#### Reliability
- Uptime target: [e.g., 99.9%]
- Error handling: [Requirements]
- Graceful degradation: [How to handle failures]
- Recovery procedures: [What happens after failures]
- Monitoring: [What to monitor]

#### Usability
- User experience: [Requirements]
- Accessibility: [WCAG compliance? Level?]
- Internationalization: [Required? Languages?]
- Browser/device support: [Requirements]

#### Maintainability
- Code quality standards: [Requirements]
- Testing coverage: [Target percentage or requirements]
- Documentation: [What needs to be documented]
- Logging: [What to log, log level requirements]

#### Compliance
- Regulatory requirements: [GDPR, HIPAA, etc.]
- Legal requirements: [Terms, privacy, etc.]
- Industry standards: [Which standards to follow]

---

### Technical Scope

#### Frontend Changes
- [ ] Component 1: Description
- [ ] Component 2: Description
- [ ] Page 1: Description
- [ ] State management: Changes needed
- [ ] Routing: Changes needed

#### Backend Changes
- [ ] API endpoint 1: `METHOD /path` - Purpose
- [ ] API endpoint 2: `METHOD /path` - Purpose
- [ ] Service 1: Changes needed
- [ ] Business logic: Changes needed

#### Database Changes
- [ ] Table/Collection 1: Changes needed
- [ ] Migration required: Yes/No
- [ ] Indexes needed: Description
- [ ] Data backfill: Yes/No

#### Infrastructure Changes
- [ ] Configuration changes: Description
- [ ] Environment variables: What to add
- [ ] Deployment changes: Description
- [ ] Scaling considerations: Description

#### Third-Party Integrations
- [ ] Service 1: What's needed
- [ ] API 1: Integration requirements
- [ ] Webhook: Setup required

#### Testing Requirements
- [ ] Unit tests: What to cover
- [ ] Integration tests: What scenarios
- [ ] E2E tests: What flows
- [ ] Performance tests: What to measure
- [ ] Security tests: What to verify

---

### Constraints

#### Technical Constraints
- Must use: [Specific technologies required]
- Cannot use: [Technologies to avoid]
- Must be compatible with: [Systems/versions]
- Architecture pattern: [Pattern to follow]

#### Time Constraints
- Deadline: [Date]
- Phases: [If multi-phase implementation]
- Estimated effort: [Small/Medium/Large]

#### Resource Constraints
- Team: [Team members involved]
- Budget: [If relevant]
- External dependencies: [What we depend on]

#### Business Constraints
- Backward compatibility: [Requirements]
- Migration strategy: [If breaking changes]
- Feature flags: [Needed? Strategy?]
- Rollout strategy: [Phased? All at once?]

---

### Success Criteria

#### Acceptance Criteria
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]

#### Measurable Outcomes
- **Metric 1**: Target value
- **Metric 2**: Target value
- **Metric 3**: Target value

#### Test Scenarios

##### Happy Path
1. **Scenario 1**: Description
   - Given: [Initial state]
   - When: [Action]
   - Then: [Expected outcome]

2. **Scenario 2**: Description
   - Given: [Initial state]
   - When: [Action]
   - Then: [Expected outcome]

##### Error Paths
1. **Error Scenario 1**: Description
   - Given: [Initial state]
   - When: [Action that causes error]
   - Then: [Expected error handling]

##### Edge Cases
1. **Edge Case 1**: Description
   - Given: [Initial state]
   - When: [Edge condition]
   - Then: [Expected behavior]

---

### Dependencies

#### Depends On (Blockers)
- [ ] Dependency 1: Description - Status
- [ ] Dependency 2: Description - Status

#### Affects (Downstream Impact)
- System 1: How it's affected
- Feature 2: How it's affected
- Team 3: Coordination needed

#### External Dependencies
- Third-party service 1: Requirement
- External API 2: Requirement

---

### Questions & Ambiguities

#### Technical Decisions Needed
1. **[Decision Point 1]**: Question or choice to make
   - Option A: Pros and cons
   - Option B: Pros and cons
   - Recommendation: [If any]

2. **[Decision Point 2]**: Question or choice to make
   - Options and trade-offs

#### Unclear Specifications
1. **[Ambiguity 1]**: What's unclear
   - Assumed: [Current assumption]
   - Need clarification on: [Specific question]

2. **[Ambiguity 2]**: What's unclear
   - Assumed: [Current assumption]
   - Need clarification on: [Specific question]

#### Missing Information
- [ ] Information 1: What's missing
- [ ] Information 2: What's missing

---

### Implementation Phases (if applicable)

#### Phase 1: [Name]
- Scope: What to build
- Duration: Estimate
- Deliverables: What gets shipped
- Success metrics: How to measure

#### Phase 2: [Name]
- Scope: What to build
- Duration: Estimate
- Deliverables: What gets shipped
- Success metrics: How to measure

---

### Context for Research

Based on this analysis, the following research is needed:

#### Codebase Patterns
- Look for: [What patterns to search for]
- Focus on: [Which areas of codebase]

#### File Structure
- Identify: [What file/directory organization to understand]
- Focus on: [Which parts of the repo]

#### Dependencies
- Research: [Which libraries/services to investigate]
- Focus on: [What functionality is needed]

#### API Context
- Find: [Which internal APIs to understand]
- Focus on: [What integrations are needed]

#### Integration Points
- Map: [What systems connect]
- Focus on: [How data flows]

---

### Risk Assessment

#### Technical Risks
1. **[Risk 1]**: Description
   - Likelihood: High/Medium/Low
   - Impact: High/Medium/Low
   - Mitigation: Strategy

2. **[Risk 2]**: Description
   - Likelihood: High/Medium/Low
   - Impact: High/Medium/Low
   - Mitigation: Strategy

#### Business Risks
1. **[Risk 1]**: Description and mitigation

---

### References

- Original requirement: [Link or path]
- Related documents: [Links]
- Similar features: [References]
- External resources: [Links]
```

## Analysis Techniques

### For PRD Documents
1. Read the entire document
2. Identify sections: overview, goals, requirements, constraints, success metrics
3. Extract user stories and acceptance criteria
4. Identify technical implications
5. Surface ambiguities

### For GitHub/Linear Issues
1. Read issue description and all comments
2. Extract labels/tags for context (bug, feature, priority)
3. Identify acceptance criteria or definition of done
4. Look for linked issues or related work
5. Check for attached files or mockups

### For User Stories
1. Parse "As a [user], I want [action], so that [benefit]"
2. Extract the user persona and their goal
3. Identify the business value
4. Break down into technical requirements
5. Add acceptance criteria if missing

### For Feature Descriptions
1. Identify the core feature being requested
2. Extract implied requirements
3. Add non-functional requirements that might be implicit
4. Define scope boundaries
5. Create acceptance criteria

## Best Practices

1. **Be Specific**: Convert vague requirements into specific, measurable criteria
2. **Ask Questions**: Surface ambiguities rather than making assumptions
3. **Think Holistically**: Consider security, performance, testing, not just functionality
4. **Use Examples**: Include concrete examples for clarity
5. **Identify Gaps**: Note what's missing from the requirement
6. **Consider Edge Cases**: Think through unusual scenarios
7. **Define Success**: Clear, measurable success criteria
8. **Think Long-term**: Consider maintenance and scalability

## Tools to Use

- **Read**: For reading PRD documents, issue descriptions, or markdown files
- **WebFetch**: If requirements are in online documents or issues
- **Grep**: To find related issues or requirements in codebase
- **Linear MCP Tools**: If working with Linear issues
  - `mcp__linear__get_issue`: Get detailed issue information
  - `mcp__linear__list_comments`: Read discussion thread

## Success Criteria

A successful requirements analysis provides:
- Clear understanding of what needs to be built and why
- Specific functional and non-functional requirements
- Defined success criteria and acceptance tests
- Identified constraints and dependencies
- List of ambiguities and questions for clarification
- Technical scope broken down by system component
- Structured format ready for implementation planning
