# Current State Analysis Skill

## Objective

Analyze the current state of the codebase to determine what already exists versus what needs to be built for a given PRD or feature specification.

## Input Required

- PRD or feature specification
- List of components/features mentioned in requirements
- Technology stack from requirements (e.g., Go, SurrealDB, Restate)

## Research Process

### 1. Inventory Existing Implementation

**For each component mentioned in the requirements:**

a. **Search for existing code:**
   - Use Glob to find relevant files by pattern
   - Use Grep to search for related functionality
   - Check common directories (cmd/, internal/, pkg/, api/)

b. **Document what exists:**
   - File paths and structure
   - Implementation status (complete, partial, missing)
   - Code quality and patterns used
   - Test coverage

c. **Identify gaps:**
   - Features mentioned in PRD but not implemented
   - Partial implementations that need completion
   - Missing tests or documentation
   - Technical debt or refactoring needs

### 2. Assess Current Architecture

**Analyze the existing system structure:**

a. **Project Structure:**
   - Examine go.mod, go.work files
   - Check directory organization
   - Review build system (Makefile, scripts)

b. **Dependencies:**
   - List installed packages
   - Identify missing required dependencies
   - Note version constraints

c. **Configuration:**
   - Find config files (YAML, JSON, env)
   - Document current settings
   - Identify configuration gaps

### 3. Evaluate Implementation Readiness

**For each PRD requirement, determine:**

a. **Status Categories:**
   - ✅ **Complete**: Fully implemented and tested
   - 🚧 **Partial**: Started but incomplete
   - ⚠️ **Needs Refactor**: Exists but needs rework
   - ❌ **Missing**: Not implemented at all

b. **Effort Assessment:**
   - Small (< 4 hours)
   - Medium (4-16 hours)
   - Large (> 16 hours)

c. **Blockers:**
   - Missing dependencies
   - Prerequisite work needed
   - Technical decisions required

### 4. Create Gap Analysis

**Build a comprehensive comparison:**

```markdown
## Component: [Name]

**PRD Requirements:**
- [Requirement 1]
- [Requirement 2]

**Current State:**
- ✅ [What exists]
- 🚧 [What's partial]
- ❌ [What's missing]

**Gap Summary:**
- Missing: [List]
- Needs Work: [List]
- Estimated Effort: [Hours/Days]

**Files Involved:**
- Existing: [paths]
- To Create: [paths]
- To Modify: [paths]
```

## Tools to Use

- **Glob**: Find files matching patterns (e.g., `**/*_test.go`, `cmd/**/main.go`)
- **Grep**: Search code for keywords, functions, types
- **Read**: Examine specific files for detailed analysis
- **Bash**: Run commands like `go list`, `git log`, `wc -l`

## Output Format

Deliver a structured **Current State Analysis** document:

```markdown
# Current State Analysis
*Generated: [timestamp]*

## Executive Summary

**Overall Implementation Status:**
- Complete: X%
- Partial: Y%
- Missing: Z%

**Estimated Total Effort to Complete PRD:**
- [X] weeks / [Y] hours

**Major Blockers:**
1. [Blocker 1]
2. [Blocker 2]

---

## Component Analysis

### 1. [Component Name]

**PRD Requirements:**
- [Req 1]
- [Req 2]

**Current Implementation:**
- Status: [Complete/Partial/Missing]
- Files: [paths]
- Tests: [coverage %]
- Quality: [assessment]

**Gap Analysis:**
- ❌ Missing: [list]
- 🚧 Incomplete: [list]
- ⚠️ Needs Refactor: [list]

**Effort Estimate:** [hours]

---

### 2. [Next Component]

[Same structure]

---

## Architecture Assessment

**Current Structure:**
```
project/
├── cmd/
├── internal/
└── pkg/
```

**Matches PRD Architecture:** [Yes/Partial/No]

**Structural Changes Needed:**
- [Change 1]
- [Change 2]

---

## Dependency Analysis

**Installed:**
- [package@version]

**Required (from PRD):**
- [package] - Status: [Installed/Missing/Wrong Version]

**To Install:**
- [package@version]

---

## Database Schema

**Current Tables/Collections:**
- [table1]: [status vs PRD]
- [table2]: [status vs PRD]

**Missing Schema:**
- [table/field list]

---

## API Endpoints

**Existing Endpoints:**
- GET /api/v0/endpoint1 - [status]
- POST /api/v0/endpoint2 - [status]

**Missing Endpoints (from PRD):**
- [endpoint] - [description]

---

## Test Coverage

**Current Coverage:**
- Unit Tests: X%
- Integration Tests: Y%
- E2E Tests: Z%

**Testing Gaps:**
- [component without tests]

---

## Configuration

**Current Config:**
- [config file paths]
- [key settings]

**Missing Config:**
- [settings needed for PRD features]

---

## Recommendations

### Priority 1 (Blockers)
1. [Action item with reasoning]

### Priority 2 (Core Features)
1. [Action item]

### Priority 3 (Polish)
1. [Action item]

---

## Implementation Sequence Suggestion

Based on current state, suggested build order:

**Phase 1: Foundation (Week 1-2)**
1. [Task based on what's missing]
2. [Task]

**Phase 2: Core Features (Week 3-6)**
1. [Task]
2. [Task]

**Phase 3: Advanced Features (Week 7+)**
1. [Task]
2. [Task]

---

## Files to Create

- [ ] path/to/new/file1.go
- [ ] path/to/new/file2.go

## Files to Modify

- [ ] existing/path/file1.go - [changes needed]
- [ ] existing/path/file2.go - [changes needed]

## Files to Delete/Refactor

- [ ] legacy/path/old.go - [reason]

---

## Risk Assessment

**High Risk:**
- [Risk description] - Mitigation: [approach]

**Medium Risk:**
- [Risk description] - Mitigation: [approach]

**Low Risk:**
- [Risk description] - Mitigation: [approach]
```

## Success Criteria

A successful current state analysis should:

1. **Comprehensively cover** all PRD components
2. **Accurately assess** implementation status (no guessing)
3. **Quantify gaps** with effort estimates
4. **Identify blockers** that must be resolved first
5. **Provide actionable** file-level guidance
6. **Sequence work** based on dependencies
7. **Include evidence** (file paths, code snippets) for assessments

## Common Patterns to Check

### Go Projects
- `cmd/` - CLI entry points
- `internal/` - Private application code
- `pkg/` - Public libraries
- `api/` - API definitions (protobuf, OpenAPI)
- `*_test.go` - Test files
- `go.mod` - Dependencies
- `Makefile` - Build automation

### Database Files
- `*.sql` - Schema definitions
- `migrations/` - Migration scripts
- `schema/` - Schema files

### Config Files
- `*.yaml`, `*.yml` - YAML config
- `*.json` - JSON config
- `.env*` - Environment variables
- `config/` - Config directory

### Documentation
- `README.md` - Project overview
- `docs/` - Documentation
- `*.md` files throughout

## Time Allocation

- **File discovery**: 15-20 minutes (Glob/Grep extensive searches)
- **Code analysis**: 20-30 minutes (Read key files, assess quality)
- **Gap identification**: 15-20 minutes (Compare PRD vs current)
- **Documentation**: 15-20 minutes (Write structured report)

**Total: 60-90 minutes** for comprehensive analysis

## Quality Checklist

Before delivering, verify:

- [ ] All PRD components analyzed
- [ ] Status categories are evidence-based (not guessed)
- [ ] File paths are actual paths found in codebase
- [ ] Effort estimates are realistic
- [ ] Dependencies checked against go.mod
- [ ] Schema compared to actual database files
- [ ] API endpoints verified against routing code
- [ ] Test coverage assessed from `*_test.go` files
- [ ] Recommendations are prioritized and actionable
