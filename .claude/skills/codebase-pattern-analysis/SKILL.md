# Codebase Pattern Analysis Skill

This skill enables agents to find similar implementations, patterns, and examples within the codebase that are relevant to a new feature or task.

## Objective

Search the codebase to identify existing patterns, implementations, and code examples that can inform the implementation of a new feature or task.

## Input Required

- **Feature/Task Description**: What is being built or modified
- **Technology/Framework Context**: Programming language, frameworks, libraries involved
- **Scope**: What aspects to focus on (e.g., API endpoints, database models, UI components)

## Research Process

### 1. Identify Search Keywords

Based on the feature description, extract:
- Core functionality keywords (e.g., "authentication", "payment", "notification")
- Technical components (e.g., "middleware", "controller", "service", "handler")
- Patterns (e.g., "validation", "error handling", "caching")

### 2. Execute Pattern Search

Use Grep and Glob tools to search for:

**Similar Features:**
```bash
# Example searches
- Grep for similar function names or class names
- Grep for similar API endpoint patterns
- Grep for similar database operations
- Glob for files with similar naming conventions
```

**Common Patterns:**
- Error handling patterns
- Validation approaches
- Data transformation logic
- API response formats
- Configuration patterns
- Test file structures

**Technology-Specific:**
- Framework-specific patterns (e.g., middleware chains, route handlers)
- Library usage examples
- Common utilities and helpers

### 3. Analyze Findings

For each pattern found:
- Read the relevant file sections
- Understand the implementation approach
- Identify reusable patterns or anti-patterns
- Note any configuration or dependencies

### 4. Categorize Results

Organize findings into:

**Directly Reusable:**
- Utilities, helpers, or base classes that can be used as-is
- Configuration patterns to follow
- Testing patterns to replicate

**Adaptable Patterns:**
- Similar implementations that need modification
- Architectural patterns to follow
- Code structure templates

**Anti-Patterns to Avoid:**
- Outdated approaches
- Known issues or tech debt
- Deprecated patterns

## Output Format

```markdown
## Codebase Pattern Analysis Results

### Similar Implementations Found

#### [Feature/Pattern Name]
- **Location**: `path/to/file.ext:line`
- **Description**: What this implementation does
- **Relevance**: How it relates to the new task
- **Key Patterns**:
  - Pattern 1: Description
  - Pattern 2: Description
- **Code Example**:
  ```[language]
  [relevant code snippet]
  ```

### Reusable Components

#### [Component Name]
- **Location**: `path/to/component:line`
- **Purpose**: What it does
- **Usage Example**: How to use it

### Common Patterns Identified

#### [Pattern Category]
- **Pattern**: Description
- **Used In**: List of files using this pattern
- **Recommendation**: How to apply this pattern

### Anti-Patterns to Avoid

#### [Anti-Pattern Name]
- **Location**: Where it appears
- **Issue**: What's wrong with it
- **Better Approach**: What to do instead

### Architecture Insights

- How similar features are structured
- Naming conventions observed
- File organization patterns
- Testing approaches

### Recommendations

1. [Specific recommendation based on findings]
2. [Another recommendation]
3. [Etc.]
```

## Search Strategy Examples

### For API Endpoint Feature
```
1. Grep for existing route definitions
2. Grep for similar HTTP methods
3. Search for validation middleware patterns
4. Find error response formats
5. Locate relevant controller/handler examples
```

### For Database Feature
```
1. Grep for similar model definitions
2. Search for migration patterns
3. Find query examples
4. Locate transaction handling
5. Identify indexing strategies
```

### For UI Component Feature
```
1. Glob for similar component files
2. Grep for similar props patterns
3. Search for styling approaches
4. Find state management examples
5. Locate component testing patterns
```

## Best Practices

1. **Start Broad, Then Narrow**: Begin with general searches, then focus on specific patterns
2. **Multiple Search Terms**: Use variations and synonyms for better coverage
3. **Context Matters**: Read surrounding code to understand why patterns exist
4. **Version Awareness**: Check git history if a pattern seems outdated
5. **Cross-Reference**: Look for patterns across multiple similar features
6. **Document Uncertainty**: Note when you're unsure if a pattern is current best practice

## Tools to Use

- **Grep**: Search file contents for patterns, function names, keywords
  - Use `-i` for case-insensitive searches
  - Use `output_mode: "content"` with `-n` to get line numbers
  - Use `-C` for context around matches

- **Glob**: Find files by naming patterns
  - `**/*.go` for all Go files
  - `**/controllers/**/*.ts` for controller files
  - `**/test/**/*` for test files

- **Read**: Examine specific files in detail once found

- **Bash**: Use `git log` or `git blame` to understand pattern evolution

## Success Criteria

A successful pattern analysis provides:
- At least 3-5 relevant existing implementations
- Clear categorization of patterns (reusable, adaptable, avoid)
- Specific file paths and line numbers
- Code examples demonstrating each pattern
- Actionable recommendations for the new implementation
