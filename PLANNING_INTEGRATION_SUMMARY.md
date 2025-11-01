# Planning Integration for PRD Builder System

**Date**: 2025-11-01
**Branch**: `algoflows/build-with-planning`
**Status**: ✅ Complete

---

## Overview

Enhanced the **PRD Builder System** (`/build-prd`) to intelligently use existing planning artifacts from `/plan-implementation`, eliminating duplicate planning effort and providing builder agents with comprehensive architectural context.

---

## What Changed

### 1. `/build-prd` Command Enhancement

**File**: `.claude/commands/build-prd.md`

**Changes**:
- Added section about providing planning artifacts (DETAILED_IMPLEMENTATION_PLAN.md, IMPLEMENTATION_ROADMAP.md)
- Updated Phase 3 workflow to check for existing plans first
- Modified example to show planning artifact detection
- Enhanced "Planning-First Flow" section with recommendations

**Key Addition**:
```markdown
**Optional**: You can also provide planning artifacts from `/plan-implementation`:
- `DETAILED_IMPLEMENTATION_PLAN.md` - Complete technical architecture and task breakdown
- `IMPLEMENTATION_ROADMAP.md` - Milestones and timeline

The builder will automatically detect and use these planning files to inform the build process.
```

### 2. PRD Builder Agent Enhancement

**File**: `.claude/agents/prd-builder.md`

**Changes**:
- Added **Phase 3.0**: Check for existing planning artifacts FIRST
- Updated workflow to skip re-planning when plans exist
- Enhanced Phase 4.2 to pass planning context to builder agents
- Added detailed example of providing planning context to builders

**Key Addition**:
```markdown
**3.0 Check for Existing Planning Artifacts (FIRST!)**

**ALWAYS check for these files first** before creating a new plan:
- `DETAILED_IMPLEMENTATION_PLAN.md`
- `IMPLEMENTATION_ROADMAP.md`

**If planning artifacts exist:**
- Read and analyze existing plan
- Validate against PRD requirements
- Use as primary planning source
- **Skip re-planning (saves 10-15 minutes!)**
- Proceed directly to Phase 4 with existing plan
```

**Builder Agent Launch Enhancement**:
```markdown
**When launching each builder agent, provide:**
1. Task specification from implementation plan
2. Planning context - Reference to DETAILED_IMPLEMENTATION_PLAN.md
3. Architecture context - Relevant architectural decisions
4. Pattern guidance - Conventions from codebase analysis
5. Integration points - How this task connects to others
```

### 3. Builder Agent Enhancement

**File**: `.claude/agents/builder-agent.md`

**Changes**:
- Added new section: "Working with Implementation Plans"
- Added **Phase 1.0**: Review planning context first
- Updated Phase 1.3 to include planning docs in reading list
- Enhanced task specification parsing to include planning context

**Key Addition**:
```markdown
### Working with Implementation Plans

**When you receive planning context:**
- **DETAILED_IMPLEMENTATION_PLAN.md** contains your task's full specification
- **Architecture decisions** have been made and documented
- **File locations** have been predetermined
- **Patterns and conventions** have been identified

**Your job is to:**
- Read and understand the specific task section from the implementation plan
- Follow the architecture specified in the plan
- Implement to spec rather than making new architectural decisions
- Use the patterns identified during planning
```

---

## Workflow Integration

### Recommended Flow: Plan → Build

```bash
# Step 1: Create comprehensive PRD
/create-prd
> Describe your feature or product
> Outputs: SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

# Step 2: Create detailed implementation plan
/plan-implementation
> SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md
> Outputs: DETAILED_IMPLEMENTATION_PLAN.md, IMPLEMENTATION_ROADMAP.md

# Step 3: Build from plan
/build-prd
> SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md
> PRD Builder detects existing plans
> Uses plans instead of re-planning
> Builder agents receive full planning context
> Result: Production-ready code
```

### Time Savings

**Before (without planning integration)**:
1. Context gathering: 5-10 min
2. Planning with architect-planner: 10-15 min
3. Building: Varies
4. Total planning: **15-25 minutes**

**After (with planning integration)**:
1. Context gathering: 5-10 min
2. **Detect existing plans: <1 min**
3. **Load plans: <1 min**
4. Building: Varies
5. Total planning: **~2 minutes**

**Savings: 13-23 minutes** per build when using `/plan-implementation` first

### Additional Benefits

**Better Builder Context**:
- Builder agents receive detailed task specifications from plan
- Architecture decisions are explicit, not inferred
- File locations predetermined
- Patterns and conventions documented
- Integration points mapped

**Consistency**:
- All builders work from same architectural vision
- Decisions made once, applied uniformly
- Reduces architectural drift
- Better team coordination

**Quality**:
- Planning includes comprehensive risk assessment
- Task dependencies clearly mapped
- Testing strategy predefined
- Acceptance criteria explicit

---

## Example: Building SPECTRA-RED

### Traditional Flow (Single Command)

```bash
/build-prd SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

# PRD Builder:
# - Reads PRD (5 min)
# - Launches 7 context agents (7 min parallel)
# - Launches architect-planner (12 min)
# - Builds with builder agents (8+ hours)
# Total: ~8.5 hours
```

### Planning-First Flow (Recommended)

```bash
# Step 1: Plan
/plan-implementation SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md
# - Creates DETAILED_IMPLEMENTATION_PLAN.md (30 min)
# - Creates IMPLEMENTATION_ROADMAP.md

# Step 2: Build (can be done later, different session)
/build-prd SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

# PRD Builder:
# - Reads PRD (5 min)
# - Launches 7 context agents (7 min parallel)
# - Detects DETAILED_IMPLEMENTATION_PLAN.md (<1 min)
# - Loads existing plan (<1 min)
# - Builds with builder agents, passing plan context (8+ hours)
# Total: ~8.2 hours + 30 min planning
# But: Planning can be reviewed/approved before building!
```

**Key Advantage**: Separation of planning from building
- Review/approve architecture before implementation
- Make architectural changes before coding
- Different team members can plan vs build
- Better governance and oversight

---

## Planning Artifacts Used

### DETAILED_IMPLEMENTATION_PLAN.md

Contains:
- Complete technical architecture with diagrams
- All component specifications
- Granular task breakdown (atomic, testable, sequenced)
- File paths for all changes
- Testing requirements
- Integration points
- Acceptance criteria
- Risk assessment with mitigations

**Used by**:
- PRD Builder: Task sequencing and wave planning
- Builder Agents: Detailed task specifications

### IMPLEMENTATION_ROADMAP.md

Contains:
- Milestones with success criteria
- Timeline visualization
- Parallel work opportunities
- Critical path analysis
- Resource requirements

**Used by**:
- PRD Builder: Understanding phases and dependencies
- Builder Agents: Context about where their task fits

### Supporting Documentation

Also available to builder agents:
- `GO_PATTERNS_REFERENCE.md` - Go coding patterns
- `SURREALDB_SCHEMA_GUIDE.md` - Database patterns
- `API_CONTEXT_CATALOG.md` - API documentation
- `RESTATE_DEEP_DIVE.md` - Workflow patterns

---

## Implementation Details

### PRD Builder Detection Logic

1. **Start Phase 3** (Architecture & Planning)
2. **Check workspace root** for planning files:
   ```
   DETAILED_IMPLEMENTATION_PLAN.md
   IMPLEMENTATION_ROADMAP.md
   ```
3. **If found**:
   - Read both files
   - Validate against PRD requirements
   - Extract task list and wave structure
   - Skip architect-planner agent
   - Proceed to Phase 4 with loaded plan
4. **If not found**:
   - Launch architect-planner agent
   - Create new plan
   - Proceed to Phase 4 with new plan

### Builder Agent Context Passing

When PRD Builder launches a builder agent:

```markdown
[Task tool for builder-agent]

You are building Task T-5 from DETAILED_IMPLEMENTATION_PLAN.md:

**Task**: Implement HTTP ingest API with Chi router

**From Implementation Plan (DETAILED_IMPLEMENTATION_PLAN.md:245-280)**:
- Component: API Layer (internal/api/handlers/)
- Architecture: Chi router with middleware chain
- Files to create:
  - internal/api/handlers/ingest.go
  - internal/api/handlers/ingest_test.go
- Acceptance Criteria:
  - [ ] POST /v1/mesh/ingest accepts scan results
  - [ ] Returns 202 Accepted with job ID
  - [ ] Validates Ed25519 signature
  - [ ] Handles rate limiting (60 req/min)
- Integration: Calls Restate workflow for processing

**Architecture Context**:
- Section 1.1: API Gateway design (lines 32-45)
- Section 2.3: Authentication with Ed25519 (lines 450-470)
- Section 2.5: Rate limiting strategy (lines 510-525)

**Patterns**:
- Follow Chi patterns from API_CONTEXT_CATALOG.md
- Use middleware chain per GO_PATTERNS_REFERENCE.md
- Table-driven tests per testing standards

Build this component following the detailed specification in the implementation plan.
```

---

## Testing the Enhancement

To test the planning integration:

1. **Generate a PRD**:
   ```bash
   /create-prd
   > Build a simple user authentication system
   ```

2. **Create implementation plan**:
   ```bash
   /plan-implementation
   > [Paste PRD or specify file]
   > Verify DETAILED_IMPLEMENTATION_PLAN.md created
   ```

3. **Build from plan**:
   ```bash
   /build-prd
   > [Paste same PRD]
   > Watch for "Found DETAILED_IMPLEMENTATION_PLAN.md!"
   > Verify planning phase skipped
   > Verify builder agents receive planning context
   ```

4. **Compare to building without plan**:
   - Delete DETAILED_IMPLEMENTATION_PLAN.md
   - Run `/build-prd` again
   - Observe architect-planner is launched
   - Note the time difference

---

## Benefits Summary

### For PRD Builder (Team Lead)

✅ **Faster planning phase** (2 min vs 15 min)
✅ **Better task coordination** - detailed plan available
✅ **Consistent architecture** - single source of truth
✅ **No duplicate effort** - planning done once
✅ **Can resume builds** - plan persists across sessions

### For Builder Agents (Team Members)

✅ **Clear specifications** - detailed task breakdown
✅ **Architectural context** - understand the big picture
✅ **Predetermined locations** - no guessing where code goes
✅ **Pattern guidance** - explicit conventions to follow
✅ **Integration clarity** - know what connects where
✅ **Testing requirements** - clear quality standards

### For Users

✅ **Faster builds** - 13-23 min saved per build
✅ **Better quality** - comprehensive planning upfront
✅ **Review opportunities** - approve architecture before coding
✅ **Governance** - separate planning from implementation
✅ **Flexibility** - plan once, build multiple times (iterations)
✅ **Documentation** - implementation plan serves as design doc

---

## Files Modified

1. `.claude/commands/build-prd.md` (+35 lines)
   - Planning artifact detection
   - Enhanced workflow description
   - Recommended planning-first flow

2. `.claude/agents/prd-builder.md` (+65 lines)
   - Phase 3.0: Planning artifact detection
   - Enhanced builder agent launching with context
   - Skip re-planning logic

3. `.claude/agents/builder-agent.md` (+55 lines)
   - Working with implementation plans section
   - Phase 1.0: Review planning context
   - Enhanced task specification parsing

4. `PLANNING_INTEGRATION_SUMMARY.md` (NEW)
   - This documentation

**Total**: ~155 lines added, 0 lines removed

---

## Future Enhancements

### Possible Improvements

1. **Plan validation**:
   - Check plan freshness (PRD modified since plan created?)
   - Detect plan-PRD drift
   - Suggest re-planning when needed

2. **Partial plan updates**:
   - Update specific sections of plan
   - Add new tasks without full re-plan
   - Track which tasks are complete

3. **Plan templates**:
   - Common architecture patterns
   - Standard task breakdowns
   - Reusable component specs

4. **Multi-PRD builds**:
   - Build from multiple PRDs
   - Merge multiple plans
   - Coordinate cross-PRD dependencies

5. **Plan versioning**:
   - Track plan changes
   - A/B test different architectures
   - Rollback to previous plans

---

## Conclusion

The planning integration enhancement makes the PRD Builder System more efficient and effective by:

1. **Eliminating duplicate work** - no re-planning when plan exists
2. **Providing better context** - builder agents get detailed specifications
3. **Enabling governance** - plan can be reviewed before building
4. **Separating concerns** - planning vs implementation
5. **Improving quality** - comprehensive architecture upfront

**Recommended workflow**: Always use `/plan-implementation` before `/build-prd` for best results.

---

**Status**: ✅ Ready for testing and deployment
