---
name: create-skill
description: Guide for creating new Claude Code skills with proper structure, metadata, and best practices. Use when user wants to create, build, or learn about making Claude Code skills.
---

# Create Claude Code Skills

This skill teaches you how to create custom Claude Code skills with proper structure and best practices.

## Skill Location

Skills can be stored in two locations:

**Personal Skills**: `~/.claude/skills/skill-name/`
- Available across all projects for the current user
- Not shared with team members

**Project Skills**: `.claude/skills/skill-name/`
- Specific to the current project
- Automatically shared via git when committed

## Required File Structure

Every skill needs a `SKILL.md` file as the entry point:

```
.claude/skills/your-skill-name/
├── SKILL.md           # Required: Main skill file
├── templates/         # Optional: Template files
├── scripts/          # Optional: Helper scripts
└── examples.md       # Optional: Additional documentation
```

## SKILL.md Format

The SKILL.md file must use YAML frontmatter followed by Markdown content:

```yaml
---
name: your-skill-name
description: Brief description of what this skill does and when to use it
---

# Your Skill Name

## Instructions
Step-by-step guidance on how to use this skill

## Examples
Concrete usage examples
```

## Metadata Fields

### Required Fields

**name**:
- Lowercase letters, numbers, and hyphens only
- Maximum 64 characters
- Example: `create-skill`, `pdf-extractor`, `test-runner`

**description**:
- Maximum 1024 characters
- Critical for Claude to discover when to use your skill
- Must explain BOTH:
  - What the skill does
  - When Claude should use it
- Include trigger terms users would mention
- Be specific, not vague

### Optional Fields

**allowed-tools**:
- Restricts which tools Claude can use when the skill is active
- Example: `allowed-tools: Read, Grep, Glob` (read-only access)
- Useful for safety-critical or read-only operations

## Best Practices

### 1. Keep Skills Focused
Each skill should address ONE specific capability, not multiple broad domains.

**Good**: "Extract text and tables from PDF files"
**Bad**: "Helps with documents"

### 2. Write Specific Descriptions
Include trigger terms that users would naturally mention:
- For PDF skills: mention "PDF files", "extract", "parse"
- For Excel skills: mention "Excel", "spreadsheets", "XLSX"
- For testing skills: mention "tests", "test runner", "pytest"

### 3. Avoid Vague Language
**Good**: "Run Jest tests and generate coverage reports for React applications"
**Bad**: "Helps with testing"

### 4. Test Activation
Verify the skill activates when expected by:
- Using different phrasings of the trigger terms
- Testing with team members
- Checking that unrelated requests don't trigger it

### 5. Document Versions
Track changes in a version history section within SKILL.md:

```markdown
## Version History

### v1.1.0 (2025-01-15)
- Added support for nested directories
- Improved error handling

### v1.0.0 (2025-01-01)
- Initial release
```

## Supporting Files

Claude loads supporting files progressively, only when needed:

**templates/**: Template files that can be referenced in your skill
**scripts/**: Helper scripts or utilities
**examples.md**: Extended examples and use cases
**reference.md**: Technical reference documentation

## Example: Creating a PDF Processing Skill

```yaml
---
name: pdf-processor
description: Extract text, images, and tables from PDF files. Use when user mentions PDFs, needs to parse PDF content, or wants to analyze PDF documents.
allowed-tools: Read, Bash, Write
---

# PDF Processor

## Instructions

1. Check if required dependencies are installed (pdfplumber, PyPDF2)
2. Read the PDF file path from user request
3. Use appropriate library based on task:
   - Text extraction: pdfplumber
   - Metadata: PyPDF2
   - Images: pdf2image
4. Process the PDF and save results
5. Report summary to user

## Examples

### Extract All Text
User: "Extract text from invoice.pdf"
Action: Use pdfplumber to extract all text, save to invoice.txt

### Extract Tables
User: "Get tables from report.pdf"
Action: Use pdfplumber.extract_tables(), save as CSV

## Dependencies

```bash
pip install pdfplumber PyPDF2 pdf2image
```
```

## Team Sharing

Project skills (in `.claude/skills/`) are automatically shared when:
1. Committed to git
2. Team members pull latest code
3. No additional installation needed

## Common Pitfalls to Avoid

1. **Overly broad descriptions**: Make descriptions specific enough to avoid false activations
2. **Missing trigger terms**: Include terms users would naturally use
3. **Too many responsibilities**: Split complex skills into focused ones
4. **Poor naming**: Use descriptive, searchable names
5. **No examples**: Always include concrete examples

## Verification Checklist

Before committing a skill, verify:

- [ ] Name uses only lowercase, numbers, and hyphens
- [ ] Description is specific and includes trigger terms
- [ ] SKILL.md has proper YAML frontmatter
- [ ] Instructions are clear and actionable
- [ ] At least one concrete example is provided
- [ ] Skill activates correctly when tested
- [ ] No unintended activations on unrelated requests

## Resources

Official documentation: https://docs.claude.com/en/docs/claude-code/skills
