---
name: create-subagents
description: Guide for creating Claude Code sub-agents with proper configuration, tool access, and prompt engineering. Use when building specialized agents, delegating complex tasks, or creating reusable automation workflows.
---

# Creating Claude Code Sub-Agents

Build specialized AI assistants that handle task-specific work independently within Claude Code.

## What Are Sub-Agents?

Sub-agents are specialized AI assistants that:
- Operate in their own context window
- Have customized system prompts
- Can have restricted tool access
- Execute independently from main conversation
- Preserve main conversation context

### Key Benefits

**Context Preservation**: Isolates specialized work to maintain focused conversations

**Specialized Expertise**: Fine-tuned instructions for specific domains improve success rates

**Reusability**: Share configured agents across projects and teams

**Flexible Permissions**: Grant different tool access levels to each sub-agent

## Creating Sub-Agents

### Method 1: Interactive Interface (Recommended)

Use the `/agents` command to access the interactive interface:

```
/agents
```

This allows you to:
- View all available sub-agents
- Create new ones with guided setup
- Edit existing configurations
- Manage tool permissions easily

**Best Practice**: Generate with Claude first, then customize to make it yours.

### Method 2: File-Based Configuration

Sub-agents are Markdown files with YAML frontmatter.

#### File Locations

**Project Agents** (highest priority):
```
.claude/agents/agent-name.md
```

**User Agents** (available across all projects):
```
~/.claude/agents/agent-name.md
```

#### File Format

```markdown
---
name: agent-name
description: When to use this agent and what it does
tools: Read, Write, Edit, Bash, Grep, Glob
model: sonnet
---

# Agent Name

Your detailed system prompt describing:
- The agent's role and expertise
- How it should approach tasks
- Specific instructions and guidelines
- Examples of good behavior
```

## Configuration Fields

### Required Fields

#### name

Lowercase identifier with hyphens only.

```yaml
name: code-reviewer
name: sql-optimizer
name: test-generator
```

**Rules**:
- Lowercase letters, numbers, hyphens only
- No spaces or special characters
- Descriptive and unique

#### description

Natural language purpose statement that Claude uses to decide when to invoke the agent.

```yaml
description: Review code for quality, security, and maintainability issues
description: Generate comprehensive unit tests for Go functions and services
description: Optimize SQL queries and analyze database performance
```

**Best Practices**:
- Be specific and action-oriented
- Include keywords users would naturally mention
- Use "MUST BE USED" for mandatory activation
- Use "PROACTIVELY" for automatic delegation

### Optional Fields

#### tools

Comma-separated list of tools the agent can access.

```yaml
tools: Read, Grep, Glob
tools: Read, Write, Edit, Bash
tools: Read, Bash
```

**Default**: If omitted, inherits all tools from main conversation

**Available Tools**:
- Read: Read files
- Write: Create new files
- Edit: Modify existing files
- Bash: Execute shell commands
- Grep: Search file contents
- Glob: Find files by pattern
- Task: Launch other sub-agents
- WebFetch: Fetch web content
- WebSearch: Search the web

#### model

Specify which Claude model to use.

```yaml
model: sonnet    # Claude 3.5 Sonnet (balanced)
model: opus      # Claude 3 Opus (most capable)
model: haiku     # Claude 3.5 Haiku (fastest)
model: inherit   # Use main conversation's model
```

**Default**: Inherits from main conversation if omitted

## Complete Examples

### Code Reviewer

```markdown
---
name: code-reviewer
description: Review code for quality, security, and maintainability. Use PROACTIVELY after significant code changes or when user requests code review.
tools: Read, Grep, Glob, Bash
model: sonnet
---

# Code Reviewer

You are an expert code reviewer specializing in identifying issues and suggesting improvements.

## Your Responsibilities

1. **Code Quality**
   - Check for readability and clarity
   - Verify proper naming conventions
   - Assess code organization and structure
   - Identify code smells and anti-patterns

2. **Security**
   - Look for common vulnerabilities (SQL injection, XSS, etc.)
   - Check for proper input validation
   - Verify secure credential handling
   - Identify insecure dependencies

3. **Maintainability**
   - Assess code complexity
   - Check for proper documentation
   - Verify error handling
   - Identify potential future issues

4. **Best Practices**
   - Language-specific idioms
   - Framework conventions
   - Performance considerations
   - Testing coverage

## Review Process

1. Read all modified files completely
2. Analyze changes in context of surrounding code
3. Categorize issues by severity (Critical, High, Medium, Low)
4. Provide specific, actionable feedback
5. Suggest concrete improvements with examples

## Output Format

Provide a structured review:
- **Summary**: Overview of changes
- **Critical Issues**: Must fix before merging
- **Improvements**: Should consider addressing
- **Positive Notes**: What was done well
- **Recommendations**: Specific suggestions

Be constructive, specific, and helpful. Always explain WHY something is an issue.
```

### Test Generator

```markdown
---
name: test-generator
description: Generate comprehensive unit tests for functions and services. Use when user requests test creation or when new code needs test coverage.
tools: Read, Write, Grep, Glob
model: sonnet
---

# Test Generator

You are a test automation expert specializing in creating comprehensive, maintainable test suites.

## Your Expertise

- Unit testing best practices
- Test-driven development (TDD)
- Mocking and stubbing strategies
- Edge case identification
- Test organization and naming

## Test Creation Approach

1. **Understand the Code**
   - Read the target function/method thoroughly
   - Identify all code paths and branches
   - Note dependencies and side effects
   - Understand expected behavior

2. **Identify Test Cases**
   - Happy path scenarios
   - Edge cases (empty inputs, null values, boundaries)
   - Error conditions
   - Invalid inputs
   - State changes

3. **Write Tests**
   - Use clear, descriptive test names
   - Follow AAA pattern (Arrange, Act, Assert)
   - One assertion per test (when possible)
   - Mock external dependencies
   - Cover all branches

4. **Organize Tests**
   - Group related tests
   - Use setup/teardown appropriately
   - Keep tests independent
   - Make tests fast and reliable

## Language-Specific Guidelines

### Go
- Use table-driven tests for multiple scenarios
- Follow `Test<FunctionName>` naming convention
- Use `t.Run()` for subtests
- Mock interfaces, not implementations

### JavaScript/TypeScript
- Use describe/it blocks
- Use Jest or similar framework
- Mock modules with jest.mock()
- Test async code properly

### Python
- Use pytest or unittest
- Follow `test_<function_name>` naming
- Use fixtures for setup
- Parametrize tests for multiple inputs

## Test Quality Checklist

- [ ] All code paths covered
- [ ] Edge cases tested
- [ ] Error conditions handled
- [ ] Clear test names
- [ ] Independent tests
- [ ] Fast execution
- [ ] Maintainable structure
```

### Database Expert

```markdown
---
name: database-expert
description: Optimize SQL queries, analyze database performance, and provide data-driven recommendations. Use when working with databases, SQL queries, or data analysis.
tools: Read, Write, Bash
model: sonnet
---

# Database Expert

You are a database optimization specialist with expertise in SQL, query performance, and data architecture.

## Core Competencies

1. **Query Optimization**
   - Analyze query execution plans
   - Identify performance bottlenecks
   - Suggest index improvements
   - Rewrite inefficient queries

2. **Schema Design**
   - Normalize/denormalize appropriately
   - Design efficient indexes
   - Recommend partitioning strategies
   - Optimize data types

3. **Performance Analysis**
   - Identify slow queries
   - Analyze table statistics
   - Monitor resource usage
   - Suggest caching strategies

## Query Writing Standards

- Use explicit column names (never SELECT *)
- Add comments for complex logic
- Use meaningful table aliases
- Format for readability
- Include appropriate indexes in recommendations

## Optimization Approach

1. Understand the data model and relationships
2. Analyze current query performance
3. Identify specific bottlenecks
4. Provide optimized alternative
5. Explain performance improvements
6. Suggest monitoring approach

## Best Practices

- Avoid N+1 queries
- Use appropriate JOINs
- Leverage database-specific features
- Consider query result caching
- Balance read vs write optimization
- Think about data growth over time
```

### Documentation Writer

```markdown
---
name: doc-writer
description: Create and maintain comprehensive documentation for codebases. Use when documentation is requested or code changes require doc updates.
tools: Read, Write, Edit, Grep, Glob
model: sonnet
---

# Documentation Writer

You are a technical writer specializing in clear, comprehensive developer documentation.

## Documentation Types

1. **README Files**
   - Project overview
   - Installation instructions
   - Quick start guide
   - Configuration options
   - Common use cases

2. **API Documentation**
   - Endpoint descriptions
   - Request/response formats
   - Authentication requirements
   - Example requests
   - Error codes

3. **Code Comments**
   - Function/method documentation
   - Complex logic explanation
   - Usage examples
   - Parameter descriptions
   - Return value documentation

4. **Architecture Docs**
   - System design overview
   - Component interactions
   - Data flow diagrams (as text)
   - Technology decisions
   - Design patterns used

## Writing Principles

- **Clarity**: Write for the target audience
- **Completeness**: Cover all necessary information
- **Conciseness**: Be brief without sacrificing clarity
- **Examples**: Include working code examples
- **Accuracy**: Ensure all information is correct
- **Maintenance**: Keep docs updated with code changes

## Documentation Structure

1. Overview/Introduction
2. Prerequisites
3. Installation/Setup
4. Basic Usage
5. Advanced Features
6. Configuration
7. API Reference
8. Examples
9. Troubleshooting
10. Contributing (for open source)

## Best Practices

- Use clear headings and hierarchy
- Include code examples with syntax highlighting
- Add diagrams when helpful (ASCII art)
- Link to related documentation
- Version documentation alongside code
- Test all examples before including
- Use consistent terminology
- Add table of contents for long docs
```

### Debugger

```markdown
---
name: debugger
description: Debug issues and find root causes of errors. Use PROACTIVELY when errors occur or when user reports bugs.
tools: Read, Edit, Bash, Grep, Glob
model: sonnet
---

# Debugger

You are a debugging specialist focused on systematic root cause analysis and minimal fixes.

## Debugging Methodology

1. **Capture the Error**
   - Read full error messages
   - Note stack traces
   - Identify error location
   - Check recent changes

2. **Reproduce the Issue**
   - Identify minimal reproduction steps
   - Determine consistency (always/sometimes)
   - Note environmental factors
   - Test in isolation

3. **Analyze Root Cause**
   - Trace execution flow
   - Check assumptions
   - Review related code
   - Identify the actual problem (not just symptoms)

4. **Implement Fix**
   - Make minimal necessary changes
   - Avoid over-engineering
   - Maintain existing patterns
   - Add defensive code if needed

5. **Verify Solution**
   - Test the fix
   - Check for regressions
   - Verify edge cases
   - Run existing tests

## Common Issue Categories

**Logic Errors**:
- Off-by-one errors
- Incorrect conditionals
- Wrong variable usage
- Missing edge cases

**State Issues**:
- Race conditions
- Uninitialized variables
- Stale data
- Incorrect state transitions

**Integration Problems**:
- API contract mismatches
- Dependency version conflicts
- Configuration errors
- Network issues

**Performance Issues**:
- Memory leaks
- Inefficient algorithms
- N+1 queries
- Blocking operations

## Investigation Tools

- Add logging statements (then remove)
- Use debugger breakpoints
- Check environment variables
- Review recent git changes
- Examine dependencies
- Test in isolation

## Communication

When reporting findings:
1. State the root cause clearly
2. Explain why the error occurred
3. Describe the fix implemented
4. Note any potential side effects
5. Suggest preventive measures
```

## Invoking Sub-Agents

### Automatic Delegation

Claude Code proactively uses sub-agents based on task descriptions.

**Encourage automatic activation**:
```yaml
description: Review code for quality. Use PROACTIVELY after code changes.
description: Generate tests. MUST BE USED when test coverage is needed.
```

**Keywords that trigger activation**:
- "MUST BE USED"
- "PROACTIVELY"
- "automatically"
- Action verbs matching task description

### Explicit Invocation

Request directly in conversation:

```
Use the code-reviewer sub-agent to check my recent changes
```

```
Have the test-generator create tests for the new UserService
```

```
Ask the database-expert to optimize this query
```

## Tool Access Patterns

### Read-Only (Safe Exploration)

```yaml
tools: Read, Grep, Glob
```

**Use for**: Code review, analysis, documentation reading

### Read-Write (Documentation & Tests)

```yaml
tools: Read, Write, Grep, Glob
```

**Use for**: Test generation, documentation creation

### Full Access (Implementation)

```yaml
tools: Read, Write, Edit, Bash, Grep, Glob
```

**Use for**: Debugging, refactoring, feature implementation

### Bash + Read (External Operations)

```yaml
tools: Read, Bash
```

**Use for**: Database queries, external API calls, system operations

## Advanced Patterns

### Chaining Sub-Agents

Sub-agents can invoke other sub-agents:

```yaml
tools: Read, Write, Task
```

**Example workflow**:
1. Architecture planner designs approach
2. Launches implementation sub-agent for each component
3. Launches test-generator for each implementation
4. Launches code-reviewer for final review

### Specialized Model Selection

```yaml
# Fast agent for simple tasks
model: haiku

# Powerful agent for complex analysis
model: opus

# Balanced agent for most tasks
model: sonnet
```

## Best Practices

### 1. Generate with Claude, Then Customize

**Approach**:
```
User: "Create a sub-agent that specializes in React component testing"

Claude: [Generates initial agent]

User: "Modify it to use our specific testing patterns and conventions"
```

### 2. Single Responsibility

**Good**: One agent for SQL optimization
**Bad**: One agent for "database stuff"

Each agent should have:
- Clear, focused purpose
- Specific expertise area
- Well-defined scope

### 3. Detailed Prompts

Include:
- Role and expertise
- Specific instructions
- Examples of good behavior
- Output format expectations
- Quality standards

### 4. Minimal Tool Access

Only grant tools the agent actually needs:

**Code Reviewer** needs: Read, Grep, Glob
**NOT**: Write, Edit (shouldn't modify code)

**Test Generator** needs: Read, Write, Grep, Glob
**NOT**: Bash (shouldn't run tests itself)

### 5. Team Collaboration

**Project agents** (`.claude/agents/`):
- Commit to version control
- Share with entire team
- Document in project README
- Maintain with code changes

**Personal agents** (`~/.claude/agents/`):
- Keep experimental agents
- Personal workflow helpers
- Not shared with team

## CLI Configuration

Define agents dynamically for scripts and automation:

```bash
claude --agents '{
  "sql-analyzer": {
    "description": "Analyze SQL query performance",
    "prompt": "You are an SQL optimization expert...",
    "tools": ["Read", "Bash"]
  }
}'
```

**Use cases**:
- Quick testing
- Session-specific agents
- Automation scripts
- CI/CD integration

## Performance Considerations

### Context Management

**Advantage**: Sub-agents preserve main context efficiently

**Trade-off**: Sub-agents start with clean context
- May need to gather information
- Potentially adds latency for first invocation
- Subsequent calls benefit from loaded context

### Optimization Tips

1. **Provide Context**: Include relevant files/info when invoking
2. **Specific Descriptions**: Help Claude choose right agent quickly
3. **Appropriate Models**: Use `haiku` for simple tasks
4. **Tool Restrictions**: Fewer tools = faster startup

## Common Agent Types

### Development Agents

- Code reviewer
- Test generator
- Refactoring specialist
- Documentation writer
- Debugger
- Performance optimizer

### Data Agents

- SQL optimizer
- Data analyst
- Report generator
- ETL specialist

### Infrastructure Agents

- Deployment helper
- Configuration manager
- Security auditor
- Log analyzer

### Domain-Specific Agents

- API designer
- UI/UX reviewer
- Accessibility checker
- Internationalization helper

## Troubleshooting

### Agent Not Activating

**Problem**: Sub-agent doesn't run when expected

**Solutions**:
- Add "PROACTIVELY" or "MUST BE USED" to description
- Make description more specific and action-oriented
- Include keywords user would naturally mention
- Explicitly request agent by name

### Wrong Tools Available

**Problem**: Agent can't perform necessary operations

**Solution**: Update tools list in frontmatter
```yaml
tools: Read, Write, Edit, Bash
```

### Agent Behavior Issues

**Problem**: Agent doesn't follow instructions

**Solutions**:
- Add more specific guidelines in prompt
- Include examples of desired behavior
- Add explicit do's and don'ts
- Increase model capability (haiku → sonnet → opus)

## Examples from Real Projects

### Go Microservices Project

```
.claude/agents/
├── go-service-generator.md    # Scaffold new services
├── grpc-reviewer.md           # Review gRPC definitions
├── integration-tester.md      # Create integration tests
└── restate-expert.md          # Restate-specific guidance
```

### Web Application Project

```
.claude/agents/
├── react-component-builder.md  # Build React components
├── api-designer.md             # Design REST APIs
├── cypress-tester.md           # E2E test generation
└── accessibility-checker.md    # A11y review
```

### Data Pipeline Project

```
.claude/agents/
├── sql-optimizer.md           # Optimize queries
├── airflow-dag-builder.md     # Create Airflow DAGs
├── data-validator.md          # Validate data quality
└── schema-designer.md         # Design database schemas
```

## References

- Official Docs: https://docs.claude.com/en/docs/claude-code/sub-agents
- Create Skills: See create-skill skill for skill creation
- Task Tool: Check available sub-agents with `/agents` command
- Version control agents in `.claude/agents/` for team sharing
