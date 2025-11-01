# Build PRD Command

Build complete, production-ready software from a Product Requirements Document using an intelligent team of specialized builder agents.

## What This Command Does

This command launches the **PRD Builder** agent which assembles and coordinates a team of specialized agents to build your software:

1. **Analyzes** your PRD to understand what needs to be built
2. **Assembles a research team** - 7 context agents gather comprehensive codebase understanding in parallel
3. **Creates** the build plan - detailed architecture and task assignments
4. **Directs the builder team** - coordinates specialized builder agents working in parallel waves
5. **Quality assurance** - validates the build against all PRD requirements
6. **Delivers** production-ready software with comprehensive documentation and build report

## How to Use

### Basic Usage

```bash
/build-prd
```

Then provide the PRD when prompted, either by:
- Specifying a PRD file path (e.g., `SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md`)
- Pasting the PRD content directly
- Describing requirements for the builder to analyze

### What Happens Next

The PRD Builder will assemble and direct its team:

1. **Analyze your PRD** (5-10 minutes)
   - Parse all requirements
   - Identify components to build
   - Determine knowledge gaps

2. **Assemble research team** (5-10 minutes, parallel)
   - Deploys 7 context research agents concurrently:
     - current-state-analysis (what exists vs what's needed)
     - codebase-pattern-analysis (find reusable patterns)
     - file-structure-mapping (understand organization)
     - dependency-research (identify required libraries)
     - api-context-gathering (document APIs)
     - integration-point-mapping (map connections)
     - technical-research (gather best practices)
   - Synthesizes team findings into unified context

3. **Create build plan** (10-15 minutes)
   - Designs technical architecture
   - Breaks down into buildable components
   - Sequences work by dependencies
   - Organizes into parallel build waves

4. **Direct builder team in waves** (varies by PRD size)
   - Wave 1: Foundation team (database, types, interfaces)
   - Wave 2: Core services team (business logic, services)
   - Wave 3: Integration team (APIs, workflows, CLI)
   - Wave 4: Quality team (testing and polish)
   - Each wave: multiple builder agents working in parallel

5. **Quality assurance** (15-30 minutes)
   - Runs all tests
   - Performs integration testing
   - Validates against PRD requirements
   - Assesses build quality

6. **Delivery report** (10 minutes)
   - Requirements completion status
   - Build statistics
   - Quality metrics
   - Outstanding issues
   - Next steps

## Example Interaction

```
User: /build-prd

PRD Builder: I'm the PRD Builder. I'll build your software by assembling and coordinating a team of specialized builder agents.

Please provide your PRD by either:
1. Specifying a file path (e.g., "SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md")
2. Pasting the PRD content
3. Describing the requirements

User: SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

Orchestrator: Reading PRD from SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md...

**PRD Analysis Complete**

This is a large PRD for building Spectra-Red Intel Mesh with:
- Must Have (P0): 7 features
- Should Have (P1): 5 features
- Nice to Have (P2): 4 features

I've identified 6 major components to implement:
1. SurrealDB graph database with schema
2. Mesh ingest API (fast path)
3. Restate workflows for enrichment
4. Query and planning APIs
5. CLI tool (spectra)
6. Authentication and security

**Launching 7 context-gathering agents in parallel...**

[Agents complete in 5-10 minutes]

**Context Research Complete**

Key findings:
- Current implementation: 0% (greenfield project)
- Required dependencies: 13 Go packages
- Architecture: Matches PRD specification
- Estimated effort: 20 weeks (160 hours)

**Creating implementation plan with architect-planner agent...**

[Plan created in 10-15 minutes]

**Implementation Plan Complete**

Breaking down into 4 waves with 47 total tasks:
- Wave 1 (Foundation): 12 tasks - 8 can run in parallel
- Wave 2 (Core Services): 15 tasks - 10 can run in parallel
- Wave 3 (Integration): 14 tasks - 8 can run in parallel
- Wave 4 (Testing): 6 tasks - 4 can run in parallel

**Starting Wave 1 - Foundation Layer**

Launching 8 feature-implementer agents in parallel...

[Tasks complete]

**Wave 1 Complete** ✓
- All 12 tasks completed
- 145 tests passing
- Code coverage: 87%

**Starting Wave 2 - Core Services**

...

[Process continues through all waves]

**All Waves Complete**

Performing integration testing...

**Validation Complete** ✓

All PRD requirements met:
- Must Have (P0): 7/7 ✓
- Should Have (P1): 5/5 ✓
- Nice to Have (P2): 3/4 ✓

Quality metrics:
- Test coverage: 89%
- All 387 tests passing
- Build successful
- Linting clean

**Execution Summary**

Successfully implemented Spectra-Red Intel Mesh per PRD specification.

[Detailed completion report follows...]
```

## When to Use This Command

### ✅ Use /execute-prd when:

- You have a complete PRD and want to implement it
- You need to transform requirements into working code
- You want to maximize parallelization and efficiency
- You need comprehensive context-gathering before implementation
- You want high-quality, tested, production-ready code

### ❌ Don't use /execute-prd when:

- You just want to create a PRD (use `/create-prd` instead)
- You only want an implementation plan (use `/plan-implementation`)
- You only need context for a single feature (use `/research-context`)
- You want to implement a single small task (just ask directly)
- You don't have clear requirements defined

## Expected Duration

Execution time depends on PRD scope:

### Small PRD (1-2 weeks of work)
- Context + Planning: 30 minutes
- Implementation: 2-4 hours
- Validation: 30 minutes
- **Total: 3-5 hours**

### Medium PRD (3-8 weeks of work)
- Context + Planning: 45 minutes
- Implementation: 6-12 hours
- Validation: 1 hour
- **Total: 8-14 hours**

### Large PRD (8-20 weeks of work)
- Context + Planning: 1 hour
- Implementation: 20-40 hours
- Validation: 2 hours
- **Total: 23-43 hours**

*Note: The orchestrator maximizes parallelization to minimize wall-clock time*

## What You'll Receive

Upon completion, you'll get:

1. **Working Implementation**
   - All code files created/modified
   - Complete test suite
   - Documentation
   - Build passing

2. **Execution Summary Report**
   - Requirements completion checklist
   - Implementation statistics
   - Quality metrics
   - Outstanding issues (if any)
   - Next steps

3. **Technical Artifacts**
   - Architecture documentation
   - API specifications
   - Database schemas
   - Workflow definitions

## Tips for Best Results

### Prepare Your PRD

Ensure your PRD includes:
- Clear requirements with priorities (P0, P1, P2)
- Technical specifications or architecture
- Success criteria and acceptance tests
- Dependencies and constraints

### Provide Context

Help the orchestrator by:
- Pointing to related PRD files
- Mentioning existing codebase areas
- Noting any constraints or preferences
- Specifying target completion criteria

### Monitor Progress

The orchestrator will:
- Show which wave is executing
- Report completion of each wave
- Highlight any blockers
- Provide status updates

You can ask for status at any time.

### Review Outputs

After each wave:
- Review implemented code
- Verify tests are passing
- Check acceptance criteria
- Provide feedback if needed

## Integration with Other Commands

### Complete Product Development Flow

```bash
# Step 1: Create PRD
/create-prd
> [Describe feature idea]
> [Orchestrator creates comprehensive PRD]

# Step 2: Execute PRD
/execute-prd
> [Provide PRD from step 1]
> [Orchestrator implements everything]

# Complete: PRD → Working Code
```

### Iterative Development Flow

```bash
# For individual features within a larger project:

# Research context first
/research-context
> [Describe feature]

# Then execute with context
/execute-prd
> [Provide feature spec + context]
```

### Planning-First Flow

```bash
# Plan first, execute later

# Step 1: Create implementation plan
/plan-implementation
> [Provide PRD]
> [Get detailed task breakdown]

# Step 2: Execute plan
/execute-prd
> [Provide PRD + implementation plan]
> [Orchestrator executes the plan]
```

## Advanced Usage

### Partial Execution

If you want to execute only part of a PRD:

```
/execute-prd

Orchestrator: Provide your PRD...

User: SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

But only implement Phase 1 (Weeks 1-8) - the core mesh infrastructure.

Orchestrator: Understood. I'll focus execution on Phase 1 requirements only:
- Must Have (P0) features from Phase 1
- Excluding Phase 2 and Phase 3 features
...
```

### Resume Execution

If execution was interrupted:

```
/execute-prd

User: Resume execution of SPECTRA_RED PRD from Wave 3

Orchestrator: Analyzing current state to resume from Wave 3...
[Checks which tasks are complete]
[Resumes from the appropriate point]
```

### Custom Validation

Specify additional validation criteria:

```
/execute-prd

User: SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

Additional requirements:
- Test coverage must be >90% (vs default 80%)
- Must include integration tests for all APIs
- Must include performance benchmarks
- Code must pass strict linting

Orchestrator: Noted. I'll apply enhanced quality standards:
- Target coverage: 90%+
- Integration tests: Required for all APIs
- Benchmarks: Required
- Linting: Strict mode
...
```

## Troubleshooting

### Execution Blocked

If the orchestrator reports a blocker:
- Review the blocker description
- Provide necessary information or context
- The orchestrator will automatically resume

### Quality Issues

If tests fail or quality is insufficient:
- The orchestrator will attempt to fix
- You can request specific improvements
- The orchestrator can launch additional test-focused agents

### Missing Context

If the orchestrator needs more information:
- It will ask specific questions
- Provide the requested context
- It may launch additional research agents

## Related Commands

- **`/create-prd`** - Generate a PRD from an idea
- **`/plan-implementation`** - Create an implementation plan from a PRD
- **`/research-context`** - Gather context for a specific feature

## System Requirements

For optimal execution:
- Sufficient token budget for agent parallelization
- Access to codebase for context gathering
- Ability to write files for implementation
- Ability to run tests for validation

---

**Ready to transform your PRD into working code? Run `/execute-prd` to begin.**
