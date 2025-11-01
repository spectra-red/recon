# Technical Research Skill

This skill enables agents to research technical approaches, best practices, architecture patterns, and implementation strategies for building features using web search and analysis.

## Objective

Research and document technical approaches, technology choices, architecture patterns, and implementation best practices for a product or feature by leveraging web-based technical resources.

## Input Required

- **Feature/Product Description**: What is being built
- **Technology Context**: Known tech stack or preferences
- **Technical Scope**: Backend, frontend, database, infrastructure, etc.

## Research Process

### 1. Define Technical Research Scope

**Identify what to research:**
- Technology stack options
- Architecture patterns
- Implementation approaches
- Best practices and anti-patterns
- Performance considerations
- Security considerations
- Scalability approaches
- Testing strategies

### 2. Technology Stack Research

**Use WebSearch to research:**

**For each technology decision:**
- "[technology A] vs [technology B] 2025"
- "[problem] best technology stack"
- "[use case] framework comparison"
- "[technology] production experience"
- "[technology] at scale"

**Research areas:**
- **Programming languages**: Best for the use case
- **Frameworks**: Modern, maintained, community support
- **Libraries**: Battle-tested, well-documented
- **Databases**: Performance, scalability, fit
- **Infrastructure**: Cloud services, deployment options
- **Tools**: Development, testing, monitoring

**Sources to prioritize:**
- Technical blogs (Engineering blogs from tech companies)
- Stack Overflow (for practical issues)
- GitHub (for real implementations and issues)
- Documentation sites
- Technical comparison articles
- Conference talks and presentations

### 3. Architecture Pattern Research

**WebSearch for architecture patterns:**
- "[use case] architecture pattern"
- "[problem] system design"
- "[technology] architecture best practices"
- "[scale] architecture patterns"

**Common patterns to research:**
- **Monolith vs Microservices**: When to use each
- **Event-driven**: Message queues, event sourcing
- **Serverless**: FaaS patterns
- **API patterns**: REST, GraphQL, gRPC
- **Data patterns**: CQRS, event sourcing, caching
- **Frontend patterns**: SPA, SSR, islands, micro-frontends

### 4. Implementation Approach Research

**WebSearch for implementation strategies:**
- "[feature] implementation guide"
- "how to build [feature] [technology]"
- "[feature] tutorial [framework]"
- "[feature] code example [language]"

**Use WebFetch to get details from:**
- Technical tutorials
- Official documentation
- Engineering blog posts
- GitHub repositories with examples
- Technical deep-dives

**Look for:**
- Step-by-step implementation guides
- Real code examples
- Common pitfalls
- Production lessons learned
- Performance optimization tips

### 5. Best Practices Research

**WebSearch for best practices:**
- "[technology] best practices 2025"
- "[use case] production best practices"
- "[feature] security best practices"
- "[technology] performance optimization"

**Research areas:**
- **Code quality**: Design patterns, clean code
- **Security**: Authentication, authorization, data protection
- **Performance**: Optimization techniques, caching strategies
- **Reliability**: Error handling, retry logic, circuit breakers
- **Testing**: Unit, integration, e2e strategies
- **Monitoring**: Logging, metrics, alerting
- **Documentation**: API docs, code comments

### 6. Anti-Patterns Research

**WebSearch for what to avoid:**
- "[technology] anti-patterns"
- "[feature] common mistakes"
- "[technology] pitfalls"
- "lessons learned [technology]"

**Look for:**
- Common mistakes and how to avoid them
- Anti-patterns that seem good but cause problems
- Deprecated approaches
- Scalability bottlenecks
- Security vulnerabilities

### 7. Performance & Scalability Research

**WebSearch for performance insights:**
- "[technology] performance benchmarks"
- "[use case] scalability patterns"
- "[technology] at scale"
- "[feature] performance optimization"

**Research:**
- Performance benchmarks
- Scalability limits and solutions
- Caching strategies
- Database optimization
- Load testing approaches

### 8. Security Research

**WebSearch for security guidance:**
- "[technology] security best practices"
- "[use case] security considerations"
- "[technology] vulnerabilities"
- "[feature] security checklist"

**Research:**
- Common vulnerabilities (OWASP Top 10)
- Authentication and authorization patterns
- Data encryption approaches
- Secure coding practices
- Security testing methods

### 9. Real-World Examples

**WebSearch for production examples:**
- "[company] engineering blog [technology]"
- "how [company] built [feature]"
- "[technology] case study"
- "[feature] production lessons"

**Use WebFetch to analyze:**
- Engineering blog posts from successful companies
- Conference talks and presentations
- Post-mortems and lessons learned
- Open source implementations

## Output Format

```markdown
## Technical Research Findings

### Executive Summary

[2-3 paragraphs covering:
- Recommended technical approach
- Key technology choices
- Major architecture decisions
- Critical considerations]

---

### Technology Stack Recommendations

#### Programming Language

**Recommendation**: [Language]

**Rationale**:
- [Reason 1]: Description
- [Reason 2]: Description
- [Reason 3]: Description

**Alternatives Considered**:
| Alternative | Pros | Cons | Why Not Chosen |
|-------------|------|------|----------------|
| [Language 1] | [Pros] | [Cons] | [Reason] |
| [Language 2] | [Pros] | [Cons] | [Reason] |

**Sources**:
- [Source 1 with URL]
- [Source 2 with URL]

---

#### Backend Framework

**Recommendation**: [Framework]

**Rationale**:
- Performance: [Benchmark data]
- Community: [GitHub stars, npm downloads, etc.]
- Ecosystem: [Available libraries, integrations]
- Maturity: [Production adoption, stability]

**Alternatives Considered**:
[Same comparison table structure]

**Sources**:
- [Source 1 with URL]
- [Source 2 with URL]

---

#### Database

**Recommendation**: [Database]

**Rationale**:
- Data model fit: [Why it fits the use case]
- Performance: [Benchmark data]
- Scalability: [Scaling approach]
- Operations: [Ease of management]

**Alternatives Considered**:
[Same comparison table structure]

**Sources**:
- [Source 1 with URL]
- [Source 2 with URL]

---

#### [Additional Technology Choices]

[Frontend framework, caching layer, message queue, etc.]

[Same structure as above for each]

---

### Architecture Recommendations

#### Recommended Architecture Pattern

**Pattern**: [Pattern name, e.g., "Event-Driven Microservices"]

**Architecture Diagram**:
```
[ASCII or text-based architecture diagram]
```

**Description**:
[Detailed description of the architecture]

**Why This Pattern**:
- [Reason 1]: Explanation
- [Reason 2]: Explanation
- [Reason 3]: Explanation

**Trade-offs**:
- **Pros**:
  - [Pro 1]
  - [Pro 2]
- **Cons**:
  - [Con 1 and mitigation]
  - [Con 2 and mitigation]

**Production Examples**:
- **[Company 1]**: [How they use this pattern]
  - Source: [URL]
- **[Company 2]**: [How they use this pattern]
  - Source: [URL]

**Sources**:
- [Architecture pattern resource with URL]
- [Case study with URL]

---

#### Alternative Patterns Considered

##### Pattern 2: [Pattern Name]

**Why Considered**: [Use case or advantage]

**Why Not Chosen**: [Reason]

**When to Reconsider**: [Conditions that would make this better]

---

### Implementation Strategy

#### High-Level Approach

**Phase 1: [Name]**
- **Goal**: [What to accomplish]
- **Key Components**:
  - Component 1: Description
  - Component 2: Description
- **Duration**: [Estimate]

**Phase 2: [Name]**
[Same structure]

---

#### Technical Implementation Details

##### Component 1: [Name]

**Purpose**: [What this component does]

**Technology**: [Framework/library to use]

**Implementation Approach**:
1. **Step 1**: Description
   - Code pattern: `[code snippet or pseudocode]`
2. **Step 2**: Description
   - Code pattern: `[code snippet or pseudocode]`

**Code Example**:
```[language]
// Real-world example from [source]
[code snippet]
```
**Source**: [URL]

**Best Practices**:
- [Practice 1]: Description
- [Practice 2]: Description

**Pitfalls to Avoid**:
- [Pitfall 1]: How to avoid
- [Pitfall 2]: How to avoid

---

##### Component 2: [Name]
[Same structure]

---

### Best Practices

#### Code Quality

**Recommended Patterns**:
1. **[Pattern 1]**: Description
   - Example: `[code snippet]`
   - Source: [URL]
2. **[Pattern 2]**: Description
   - Example: `[code snippet]`
   - Source: [URL]

**Code Organization**:
- File structure: [Recommended structure]
- Naming conventions: [Conventions to follow]
- Module organization: [How to organize]

**Source**: [URL to style guide or best practices]

---

#### Security Best Practices

**Authentication**:
- **Recommended Approach**: [JWT, OAuth, etc.]
- **Implementation**: [How to implement]
- **Libraries**: [Recommended libraries]
- **Code Example**:
  ```[language]
  [Example code]
  ```
- **Source**: [URL]

**Authorization**:
- **Recommended Approach**: [RBAC, ABAC, etc.]
- **Implementation**: [How to implement]
- **Source**: [URL]

**Data Protection**:
- Encryption at rest: [Approach]
- Encryption in transit: [Approach]
- Sensitive data handling: [Guidelines]
- **Source**: [URL]

**Security Checklist**:
- [ ] Input validation
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] CSRF protection
- [ ] Rate limiting
- [ ] Security headers
- [ ] [Additional items]

**Source**: [OWASP or security guide URL]

---

#### Performance Optimization

**Recommended Optimizations**:

1. **[Optimization 1]**: Description
   - **Impact**: [Expected improvement]
   - **Implementation**:
     ```[language]
     [Code example]
     ```
   - **Source**: [URL]

2. **[Optimization 2]**: Description
   - **Impact**: [Expected improvement]
   - **Implementation**: [How to implement]
   - **Source**: [URL]

**Caching Strategy**:
- **Where to cache**: [Layers]
- **What to cache**: [Data types]
- **Cache invalidation**: [Strategy]
- **Technology**: [Redis, CDN, etc.]
- **Source**: [URL]

**Database Optimization**:
- **Indexing**: [Strategy]
- **Query optimization**: [Approaches]
- **Connection pooling**: [Configuration]
- **Source**: [URL]

---

#### Testing Strategy

**Unit Testing**:
- **Framework**: [Recommended framework]
- **Coverage target**: [Percentage]
- **What to test**: [Guidelines]
- **Example**:
  ```[language]
  [Test example]
  ```
- **Source**: [URL]

**Integration Testing**:
- **Approach**: [Strategy]
- **Tools**: [Recommended tools]
- **Scenarios to cover**: [List]
- **Source**: [URL]

**E2E Testing**:
- **Framework**: [Tool like Playwright, Cypress]
- **Approach**: [Strategy]
- **Source**: [URL]

**Performance Testing**:
- **Tools**: [Load testing tools]
- **Metrics**: [What to measure]
- **Targets**: [Performance targets]
- **Source**: [URL]

---

### Anti-Patterns to Avoid

#### Anti-Pattern 1: [Name]

**Description**: [What this anti-pattern is]

**Why It's Bad**:
- [Problem 1]
- [Problem 2]

**Example of Bad Code**:
```[language]
// Anti-pattern example
[code]
```

**Correct Approach**:
```[language]
// Better approach
[code]
```

**Source**: [URL]

---

#### Anti-Pattern 2: [Name]
[Same structure]

---

### Scalability Considerations

#### Horizontal Scaling

**Approach**: [Description]

**Implementation**:
- [Aspect 1]: How to handle
- [Aspect 2]: How to handle

**Technologies**: [Load balancers, orchestration, etc.]

**Source**: [URL]

---

#### Database Scaling

**Read Scaling**: [Replication, read replicas]

**Write Scaling**: [Sharding, partitioning]

**Caching**: [Strategy]

**Source**: [URL]

---

#### Performance Targets

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Response time | < X ms | [How to measure] |
| Throughput | X req/sec | [How to measure] |
| Concurrent users | X users | [How to measure] |

**Benchmarks from Similar Systems**:
- [System 1]: [Performance numbers] - Source: [URL]
- [System 2]: [Performance numbers] - Source: [URL]

---

### Real-World Examples & Case Studies

#### Example 1: [Company] - [Feature/System]

**Overview**: [What they built]

**Scale**: [Users, requests, data volume]

**Technical Approach**:
- Architecture: [Pattern used]
- Technologies: [Stack]
- Key decisions: [Important choices and rationale]

**Lessons Learned**:
1. **[Lesson 1]**: Description
2. **[Lesson 2]**: Description

**What We Can Apply**:
- [Applicable insight 1]
- [Applicable insight 2]

**Source**: [Engineering blog URL]

---

#### Example 2: [Company] - [Feature/System]
[Same structure]

---

### Technical Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation Strategy |
|------|------------|--------|---------------------|
| [Risk 1] | High/Med/Low | High/Med/Low | [Strategy] |
| [Risk 2] | High/Med/Low | High/Med/Low | [Strategy] |

**Sources**: [URLs to risk analysis or post-mortems]

---

### Monitoring & Observability

**Metrics to Track**:
- [Metric 1]: Why and how
- [Metric 2]: Why and how

**Logging Strategy**:
- **What to log**: [Guidelines]
- **Log levels**: [Usage]
- **Log aggregation**: [Tool]

**Alerting**:
- **Critical alerts**: [Conditions]
- **Warning alerts**: [Conditions]

**Tools Recommended**:
- Monitoring: [Tool and why]
- Logging: [Tool and why]
- Tracing: [Tool and why]

**Source**: [Observability guide URL]

---

### Development Workflow

**Recommended Workflow**:
1. **[Step 1]**: Description
2. **[Step 2]**: Description

**CI/CD Pipeline**:
- Build: [Approach]
- Test: [Strategy]
- Deploy: [Strategy]

**Tools**: [Recommended tools]

**Source**: [DevOps guide URL]

---

### Key Technical Insights

1. **[Insight 1]**: [Important technical finding and implication]
2. **[Insight 2]**: [Important technical finding and implication]
3. **[Insight 3]**: [Important technical finding and implication]

---

### Recommendations Summary

**Must Do**:
1. [Critical recommendation 1]
2. [Critical recommendation 2]

**Should Do**:
1. [Important recommendation 1]
2. [Important recommendation 2]

**Consider**:
1. [Optional recommendation 1]
2. [Optional recommendation 2]

**Avoid**:
1. [Anti-pattern 1]
2. [Anti-pattern 2]

---

### Sources

**Official Documentation**:
- [Technology 1 Docs]: [URL]
- [Technology 2 Docs]: [URL]

**Engineering Blogs**:
- [Company Blog Post]: [URL]
- [Company Blog Post]: [URL]

**Technical Articles**:
- [Article]: [URL]
- [Article]: [URL]

**GitHub Repositories**:
- [Example Implementation]: [URL]
- [Example Implementation]: [URL]

**Benchmarks & Comparisons**:
- [Benchmark]: [URL]
- [Comparison]: [URL]

**Video Resources**:
- [Conference Talk]: [URL]
- [Tutorial]: [URL]

---

### Research Methodology

**Technologies Evaluated**: [List]

**Sources Consulted**: [Number and types]

**Search Queries Used**:
- "[Query 1]"
- "[Query 2]"

**Code Examples Reviewed**: [Number]

**Production Case Studies Analyzed**: [Number]

**Limitations**:
- [Limitation 1]
- [Limitation 2]

**Date of Research**: [Date]
```

## Search Strategies

### Technology Comparison
```
WebSearch queries:
- "[tech A] vs [tech B] 2025"
- "best [technology] for [use case]"
- "[technology] production experience"
- "[technology] pros and cons"
```

### Architecture Research
```
WebSearch queries:
- "[use case] system design"
- "[scale] architecture patterns"
- "[technology] architecture best practices"
- "how [company] built [system]"
```

### Implementation Guidance
```
WebSearch queries:
- "how to build [feature] [technology]"
- "[feature] implementation guide [framework]"
- "[feature] tutorial 2025"
- "[technology] code examples"
```

### Best Practices
```
WebSearch queries:
- "[technology] best practices 2025"
- "[use case] production best practices"
- "[technology] security best practices"
- "[technology] performance optimization"
```

### Real-World Examples
```
WebSearch queries:
- "[company] engineering blog [technology]"
- "[technology] case study"
- "[company] tech stack"
- "lessons learned [technology]"
```

## Best Practices

1. **Focus on Production-Ready**: Prioritize battle-tested approaches over bleeding-edge
2. **Seek Real Examples**: Code examples and case studies over theory
3. **Consider Scale**: Research how approaches perform at scale
4. **Security First**: Always research security implications
5. **Performance Matters**: Benchmark data and optimization techniques
6. **Learn from Failures**: Post-mortems and lessons learned are valuable
7. **Check Dates**: Technology moves fast - prioritize recent (2024-2025) sources
8. **Diverse Sources**: Mix official docs, blogs, tutorials, and discussions
9. **Validate Claims**: Cross-reference benchmarks and recommendations
10. **Use WebFetch**: Deep dive into promising technical articles

## Tools to Use

- **WebSearch**: Primary discovery tool
  - Technology comparisons
  - Best practices
  - Code examples
  - Case studies

- **WebFetch**: Deep analysis tool
  - Engineering blog posts
  - Technical documentation
  - Tutorial articles
  - GitHub README files

## Success Criteria

A successful technical research provides:
- Clear technology stack recommendations with rationale
- Architecture pattern recommendation with diagram
- Implementation strategy with code examples
- Comprehensive best practices across code, security, performance, testing
- Anti-patterns to avoid with examples
- Scalability and performance considerations
- Real-world case studies from production systems
- At least 10-15 credible technical sources cited
- Actionable recommendations prioritized by importance
- Specific code examples demonstrating patterns
