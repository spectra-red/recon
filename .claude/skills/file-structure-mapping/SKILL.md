# File Structure Mapping Skill

This skill enables agents to analyze repository structure and identify all relevant files, directories, and organizational patterns for a new feature or task.

## Objective

Map the repository structure to identify where code should be placed, which files need modification, and how the codebase is organized.

## Input Required

- **Feature/Task Description**: What is being built or modified
- **Technology Stack**: Programming languages and frameworks used
- **Scope**: What parts of the system are affected (backend, frontend, database, etc.)

## Research Process

### 1. Repository Overview

**High-Level Structure:**
```bash
# Use Glob to identify main directories
- Application source directories (src/, lib/, app/, etc.)
- Test directories (test/, tests/, __tests__/, etc.)
- Configuration directories (config/, etc/)
- Documentation (docs/, README files)
- Build/deployment (build/, dist/, deploy/, docker/, etc.)
```

### 2. Technology-Specific Organization

**Identify patterns like:**
- MVC structure (models/, views/, controllers/)
- Feature-based (features/, modules/, packages/)
- Layer-based (services/, repositories/, handlers/)
- Component-based (components/, pages/, layouts/)
- Domain-driven (domains/, aggregates/, entities/)

### 3. Locate Relevant Files

**For the specific feature, find:**

**Entry Points:**
- Main application files
- Route definitions
- API endpoint declarations
- Service registrations

**Feature Files:**
- Where similar features are implemented
- File naming conventions
- Directory structure for features

**Supporting Files:**
- Configuration files
- Schema/migration files
- Type definitions
- Utility/helper files
- Middleware files

**Test Files:**
- Test directory structure
- Test file naming conventions
- Test helper/fixture locations

### 4. Identify Dependencies

**File-to-file relationships:**
- Import/require patterns
- Module dependencies
- Shared utilities
- Common interfaces/types

### 5. Map Integration Points

**Where the new code will connect:**
- Routers/handlers to modify
- Services to extend
- Schemas to update
- Config files to change

## Output Format

```markdown
## File Structure Analysis

### Repository Structure Overview

```
[project-root]/
├── src/
│   ├── api/          # API layer
│   ├── services/     # Business logic
│   ├── models/       # Data models
│   ├── utils/        # Utilities
│   └── config/       # Configuration
├── tests/
│   ├── unit/
│   └── integration/
└── docs/
```

### Organizational Pattern
- **Type**: [MVC / Feature-based / Layer-based / etc.]
- **Description**: How code is organized in this repository

### Relevant Directories

#### [Directory Name]
- **Path**: `path/to/directory/`
- **Purpose**: What this directory contains
- **Relevance**: Why it matters for this task
- **File Count**: Number of files
- **Key Files**:
  - `file1.ext` - Description
  - `file2.ext` - Description

### Files to Create

#### [New File Name]
- **Location**: `path/to/new/file.ext`
- **Purpose**: What this file will do
- **Rationale**: Why it should go here
- **Template/Example**: Similar file to use as template

### Files to Modify

#### [Existing File Name]
- **Location**: `path/to/file.ext:line`
- **Current Purpose**: What it currently does
- **Required Changes**: What needs to be modified
- **Impact**: How significant the changes will be

### Naming Conventions

- **Files**: [convention observed, e.g., kebab-case, camelCase, PascalCase]
- **Directories**: [convention observed]
- **Tests**: [convention observed, e.g., *.test.js, *_test.go]
- **Examples**:
  - Feature file: `user-authentication.service.ts`
  - Test file: `user-authentication.service.test.ts`

### Directory Recommendations

**Where to place new code:**
1. **Main Implementation**: `path/to/directory/`
   - Rationale: [Why this location]

2. **Tests**: `path/to/test/directory/`
   - Rationale: [Why this location]

3. **Configuration**: `path/to/config/`
   - Rationale: [Why this location]

### Configuration Files

#### [Config File Name]
- **Location**: `path/to/config.ext`
- **Purpose**: What it configures
- **Changes Needed**: What to add/modify

### Schema/Migration Files

#### [Schema File Name]
- **Location**: `path/to/schema.ext`
- **Purpose**: Data structure definition
- **Changes Needed**: What to add/modify

### Import Path Patterns

**How modules are imported:**
- Absolute imports: `import { X } from '@/services/X'`
- Relative imports: `import { X } from '../services/X'`
- Alias configuration: Where aliases are defined

### File Organization Insights

- How features are typically structured
- Whether files are co-located or separated by type
- Testing file placement (alongside source vs separate directory)
- Documentation expectations

### Integration Points

**Files that will need to import/use the new code:**
1. `path/to/file1.ext:line` - Router registration
2. `path/to/file2.ext:line` - Service initialization
3. `path/to/file3.ext:line` - Type exports

### Recommendations

1. **File Placement**: [Specific recommendation]
2. **Naming**: [Follow these conventions]
3. **Organization**: [Organize code like this]
4. **Dependencies**: [Import from these locations]
```

## Search Strategies

### Identify Main Directories
```bash
# Find all top-level directories
Glob: **/
# Look for common source directories
Glob: {src,lib,app,api,server,client}/**/*
```

### Find Similar Features
```bash
# Search for feature files
Glob: **/*[keyword]*.*
# Search in specific directories
Glob: src/features/**/*
```

### Locate Configuration
```bash
# Find config files
Glob: **/{config,*.config.*,.*rc,*.json,*.yaml,*.yml}
```

### Map Test Structure
```bash
# Find test directories
Glob: {test,tests,__tests__,spec}/**/*
# Find test files
Glob: **/*.{test,spec}.{js,ts,go,py}
```

### Find Type Definitions
```bash
# TypeScript
Glob: **/*.d.ts
Glob: **/types/**/*
# Go
Glob: **/types.go
```

## Best Practices

1. **Start at Root**: Begin with top-level structure before diving deep
2. **Follow Conventions**: Identify and respect existing patterns
3. **Check Multiple Examples**: Look at several features to confirm patterns
4. **Note Inconsistencies**: Document when structure varies
5. **Consider Growth**: Think about how the structure will scale
6. **Map Dependencies**: Understand file relationships, not just locations
7. **Include Tests**: Always map test file locations and conventions

## Tools to Use

- **Glob**: Primary tool for discovering files and directories
  - Use `**/` to find all directories
  - Use `**/*` to find all files recursively
  - Use `**/*.ext` to find files by extension

- **Bash**: For more complex directory analysis
  - `tree` command if available
  - `find` for specific queries
  - `ls -R` for recursive listing

- **Read**: Examine key files like:
  - `package.json`, `go.mod`, `requirements.txt` for project structure clues
  - Main entry files to understand bootstrapping
  - README files for architecture documentation

- **Grep**: Search for import patterns
  - Find how modules reference each other
  - Identify import alias usage
  - Locate module exports

## Success Criteria

A successful file structure analysis provides:
- Clear directory tree showing repository organization
- Specific recommendations for where to place new files
- List of existing files that need modification
- Naming conventions to follow
- Import/export patterns to use
- Configuration files that need updates
- Test file placement guidance
