# Dependency Research Skill

This skill enables agents to research libraries, frameworks, external services, and dependencies needed for implementing a feature or task.

## Objective

Identify, evaluate, and gather comprehensive information about dependencies (libraries, frameworks, APIs, external services) relevant to the feature being implemented.

## Input Required

- **Feature/Task Description**: What is being built
- **Technology Stack**: Programming language and existing frameworks
- **Scope**: What functionality is needed from external dependencies

## Research Process

### 1. Identify Current Dependencies

**Analyze existing dependency files:**
- `package.json` (Node.js/JavaScript)
- `go.mod` / `go.sum` (Go)
- `requirements.txt` / `pyproject.toml` (Python)
- `Gemfile` (Ruby)
- `pom.xml` / `build.gradle` (Java)
- `Cargo.toml` (Rust)
- `composer.json` (PHP)

**Extract:**
- Current versions of relevant libraries
- Existing dependencies that might already provide needed functionality
- Dependency management patterns

### 2. Search for Existing Internal Usage

**Use Grep to find:**
- How existing dependencies are imported and used
- Configuration patterns for libraries
- Common usage patterns across the codebase
- Any custom wrappers or utilities built around dependencies

### 3. Identify New Dependencies Needed

**Determine if new libraries are required for:**
- Core functionality (e.g., JWT handling, encryption, validation)
- Testing (e.g., mocking, fixtures, assertions)
- Build/deployment tools
- Development utilities

### 4. External Research

**For each potential dependency, research:**

**Official Documentation:**
- Installation instructions
- Quick start guides
- API reference
- Configuration options
- Migration guides (if upgrading)

**Best Practices:**
- Common usage patterns
- Security considerations
- Performance implications
- Known issues and limitations

**Community Resources:**
- GitHub repository (stars, issues, recent activity)
- Stack Overflow discussions
- Blog posts and tutorials
- Real-world examples

**Evaluation Criteria:**
- Maintenance status (last commit, release frequency)
- Community size and support
- License compatibility
- Bundle size / performance impact
- Security track record
- Framework compatibility

### 5. Version Compatibility

**Check compatibility with:**
- Current language version
- Existing framework versions
- Other dependencies (peer dependencies)
- Build tools and deployment environment

## Output Format

```markdown
## Dependency Research Results

### Current Dependencies Analysis

#### Existing Relevant Dependencies

##### [Library Name] (v[version])
- **Purpose**: What it's currently used for
- **Relevant to Task**: How it might help with the new feature
- **Current Usage**: `path/to/file.ext:line` - Example usage
- **Recommendation**: Use as-is / Extend usage / Consider alternative

#### Dependency File
- **Location**: `path/to/dependency/file`
- **Package Manager**: npm / go modules / pip / etc.
- **Lock File**: Present/Absent

### New Dependencies Recommended

#### [Library Name]

**Overview:**
- **Package**: `package-name`
- **Current Version**: v[X.Y.Z]
- **License**: [License type]
- **Repository**: [GitHub URL]
- **Documentation**: [Docs URL]

**Why This Library:**
- Solves [specific problem]
- [Advantage 1]
- [Advantage 2]

**Installation:**
```bash
[installation command]
```

**Basic Usage:**
```[language]
// Example code showing basic usage
```

**Configuration:**
```[language]
// Configuration example
```

**Integration Points:**
- Where in the codebase this will be used
- What files will import this
- Any initialization needed

**Considerations:**
- Bundle size: [size info]
- Performance: [performance notes]
- Security: [security considerations]
- Breaking changes: [if upgrading]

**Alternatives Considered:**
1. **[Alternative 1]**: Why not chosen
2. **[Alternative 2]**: Why not chosen

### External Services/APIs

#### [Service Name]

**Overview:**
- **Type**: REST API / GraphQL / SDK / etc.
- **Documentation**: [URL]
- **Authentication**: [Auth method needed]

**Integration Requirements:**
- API keys / credentials needed
- Environment variables to configure
- Rate limits and quotas

**Usage Example:**
```[language]
// Example API call or SDK usage
```

**Cost Considerations:**
- Pricing tier needed
- Expected usage volume
- Cost per request/month

### Dependency Tree Impact

**New Dependencies Will Add:**
- Direct dependencies: [count]
- Transitive dependencies: [count if known]
- Total size addition: [size estimate]

**Potential Conflicts:**
- [Dependency X] requires [version range]
- [Framework Y] peer dependency requires [version]
- Resolution: [How to handle]

### Version Recommendations

| Dependency | Recommended Version | Reason |
|------------|-------------------|---------|
| [name] | vX.Y.Z | [Why this version] |
| [name] | vX.Y.Z | [Why this version] |

### Code Examples from Documentation

#### [Use Case 1]
```[language]
// Real-world example from docs or GitHub
```
**Source**: [URL]

#### [Use Case 2]
```[language]
// Another relevant example
```
**Source**: [URL]

### Best Practices & Patterns

1. **[Practice 1]**: Description and why it matters
2. **[Practice 2]**: Description and why it matters
3. **[Practice 3]**: Description and why it matters

### Known Issues & Pitfalls

#### [Issue 1]
- **Problem**: Description of the issue
- **Affected Versions**: Which versions
- **Workaround**: How to avoid or fix
- **Source**: [Link to issue/discussion]

#### [Issue 2]
- **Problem**: Description
- **Workaround**: Solution

### Security Considerations

- Known vulnerabilities: [None / List with CVEs]
- Security best practices: [List practices]
- Authentication/authorization patterns: [Patterns to follow]
- Data validation requirements: [Requirements]

### Testing Dependencies

#### [Test Library Name]
- **Purpose**: What testing functionality it provides
- **Installation**: `[command]`
- **Usage Example**:
  ```[language]
  // Test example
  ```

### Migration Guide (if upgrading existing dependency)

**From**: v[old] **To**: v[new]

**Breaking Changes:**
1. [Change 1] - How to update code
2. [Change 2] - How to update code

**Migration Steps:**
1. [Step 1]
2. [Step 2]

**Estimated Effort**: [Low/Medium/High]

### Resources

**Documentation:**
- Official docs: [URL]
- API reference: [URL]
- Migration guides: [URL]

**Tutorials & Examples:**
- [Tutorial 1]: [URL]
- [Example repo]: [URL]
- [Blog post]: [URL]

**Community:**
- GitHub issues: [URL]
- Stack Overflow tag: [URL]
- Discord/Slack: [URL]

### Implementation Checklist

- [ ] Install dependency
- [ ] Add to dependency file
- [ ] Update lock file
- [ ] Configure environment variables (if needed)
- [ ] Create utility wrappers (if needed)
- [ ] Add type definitions (if needed)
- [ ] Write integration tests
- [ ] Update documentation
- [ ] Security scan for vulnerabilities
```

## Search Strategies

### Find Dependency Files
```bash
Glob: {package.json,go.mod,requirements.txt,Gemfile,pom.xml,build.gradle,Cargo.toml,composer.json}
```

### Find Library Usage
```bash
Grep: "import.*library-name"
Grep: "require.*library-name"
Grep: "from library-name import"
```

### Find Configuration
```bash
Grep: "library-name" in {*.config.*,config/**/*}
Glob: **/*library-name*.config.*
```

## External Research Tools

### Use WebSearch for:
- "[library-name] getting started 2025"
- "[library-name] vs [alternative] comparison"
- "[library-name] best practices"
- "[library-name] [framework-name] integration"
- "[library-name] security issues"
- "[library-name] production setup"

### Use WebFetch for:
- Official documentation pages
- GitHub README files
- Popular tutorial articles
- npm/PyPI/crates.io package pages

### Use Context7 MCP Tools for:
- Library-specific documentation (if available)
- Framework integration guides
- Code examples and snippets

## Best Practices

1. **Verify Maintenance**: Check when the library was last updated
2. **Check Community**: Look for active issues, PRs, and discussions
3. **Test Compatibility**: Verify version compatibility before recommending
4. **Security First**: Always check for known vulnerabilities
5. **Consider Bundle Size**: Especially for frontend dependencies
6. **Read Real Code**: Find actual implementations, not just docs
7. **Evaluate Alternatives**: Consider at least 2-3 options before deciding
8. **Document Trade-offs**: Explain why you chose one library over another

## Tools to Use

- **Read**: For dependency files and configuration
- **Grep**: To find existing library usage in codebase
- **Glob**: To locate dependency-related files
- **WebSearch**: For current best practices and comparisons
- **WebFetch**: For official docs and detailed guides
- **MCP Context7**: For up-to-date library documentation

## Success Criteria

A successful dependency research provides:
- Complete list of dependencies needed (new and existing)
- Specific version recommendations with rationale
- Installation and configuration instructions
- Real code examples showing usage
- Security and compatibility analysis
- Migration guide if upgrading
- Links to official documentation
- Known issues and workarounds
- Testing approach for the dependency
