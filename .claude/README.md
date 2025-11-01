# Claude Code Agent & Skill System

This repository contains a comprehensive multi-agent system for automated context research, PRD generation, implementation planning, and **full PRD execution**, built for use with Claude Code.

## System Overview

Four major agent orchestration systems:

### 1. Context Research System
**Purpose**: Gather comprehensive implementation context for features/tasks

**Command**: `/research-context`

**Architecture**: Orchestrator + 6 concurrent research agents

### 2. PRD Generation System
**Purpose**: Create production-ready Product Requirements Documents

**Command**: `/create-prd`

**Architecture**: Orchestrator + 10 concurrent research agents (6 local + 4 web)

### 3. Implementation Planning System
**Purpose**: Transform PRDs into technical architectures and actionable task breakdowns

**Command**: `/plan-implementation`

**Architecture**: Architect-Planner agent + optional research agents

### 4. PRD Builder System 🆕
**Purpose**: Build complete, production-ready software from PRDs using a team of intelligent agents

**Command**: `/build-prd`

**Architecture**: PRD Builder + 7 context research agents + team of parallel builder agents

---

## Context Research System

### Components

#### Agent
- **`context-orchestrator.md`** - Coordinates parallel context gathering

#### Skills (6)
1. **`codebase-pattern-analysis`** - Find similar implementations and patterns
2. **`file-structure-mapping`** - Map repository organization
3. **`dependency-research`** - Research libraries and dependencies
4. **`api-context-gathering`** - Document internal APIs
5. **`requirements-analysis`** - Parse requirements into specs
6. **`integration-point-mapping`** - Map system integrations

#### Command
- **`/research-context`** - Invoke the orchestrator

### Workflow

1. **Analyze** requirements (30s)
2. **Launch 6 agents** in parallel (2-3 min)
3. **Synthesize** findings (1-2 min)
4. **Deliver** comprehensive context document

**Total Time**: 5-7 minutes

### Output

Comprehensive context document with:
- Requirements analysis
- Similar codebase implementations
- File structure recommendations
- Dependency requirements
- API integration points
- Implementation blueprint
- Testing strategy

---

## PRD Generation System

### Components

#### Agent
- **`prd-orchestrator.md`** - Coordinates PRD creation

#### Skills (10)

**Local Repository Research (6 skills - reused from Context Research)**:
1. `requirements-analysis`
2. `codebase-pattern-analysis`
3. `file-structure-mapping`
4. `dependency-research`
5. `api-context-gathering`
6. `integration-point-mapping`

**Web Research (4 new skills)**:
7. **`market-research`** - Research market size, growth, opportunity
8. **`competitor-analysis`** - Analyze competitive landscape
9. **`technical-research`** - Research technical approaches and best practices
10. **`user-research`** - Research user needs, personas, pain points

**PRD Creation**:
11. **`prd-writer`** - Synthesize research into PRD
12. **`prd-reviewer`** - Review and critique PRD quality

#### Command
- **`/create-prd`** - Invoke the PRD orchestrator

### Workflow

1. **Gather** requirements via conversation (2-3 min)
2. **Launch 10 agents** in parallel (3-5 min):
   - 6 local repo research agents
   - 4 web research agents
3. **Generate** PRD from research (2-3 min)
4. **Review & Refine** iteratively (2-4 min, 1-2 cycles)
5. **Deliver** production-ready PRD

**Total Time**: 10-18 minutes

### Output

Complete PRD with 9 sections (OpenAI template):
1. Executive Summary
2. Market Opportunity (with web research data)
3. Strategic Alignment
4. Customer & User Needs (with web research)
5. Value Proposition & Messaging
6. Competitive Advantage (with competitor analysis)
7. Product Scope and Use Cases (with codebase context)
8. Non-Functional Requirements (with technical research)
9. Go-to-Market Approach

---

## Implementation Planning System

### Components

#### Agent
- **`architect-planner.md`** - Transforms PRDs into technical architectures and implementation plans

#### Command
- **`/plan-implementation`** - Invoke the architect-planner

### Workflow

1. **Ingest** PRD and analyze requirements (2-3 min)
2. **Research** missing context with agents if needed (3-5 min, optional)
3. **Design** technical architecture (5-10 min)
4. **Breakdown** into granular, sequenced tasks (5-10 min)
5. **Create** implementation roadmap with milestones (3-5 min)

**Total Time**: 15-30 minutes

### Output

Complete **Implementation Plan** with:
- **Technical Architecture**: System design, components, data flow, integrations
- **Component Specifications**: Purpose, interfaces, data models, implementation approach
- **Task Breakdown**: Granular, testable, sequenced tasks with acceptance criteria
- **Implementation Roadmap**: Milestones, timeline, dependencies, parallel work
- **Risk Assessment**: Technical risks and mitigation strategies
- **Testing Strategy**: Unit, integration, E2E test plans
- **Deployment Plan**: Environments, release strategy, rollback plan
- **Resource Plan**: Team size, skills, external dependencies

### Key Features

- **PRD to Tasks**: Transforms "what to build" into "how to build it"
- **Granular Decomposition**: Atomic tasks (1-4 hours each), testable, independent
- **Smart Sequencing**: Risk-first, dependency-aware, enables incremental validation
- **Optional Research**: Launches context research agents if technical details missing
- **Implementation-Ready**: Step-by-step guidance for developers
- **Milestone-Based**: Clear checkpoints with success criteria

---

## PRD Builder System 🆕

### Components

#### Agents
- **`prd-builder.md`** - The builder who assembles and coordinates a team of agents to build software
- **`builder-agent.md`** - Team member who builds specific components from specifications

#### Skill
- **`current-state-analysis`** - Analyzes codebase to determine what exists vs what needs to be built

#### Command
- **`/build-prd`** - Build a complete software implementation from a PRD

### Workflow

1. **Analyze** PRD and understand what to build (5-10 min)
2. **Assemble research team** - Deploy 7 context agents in parallel (5-10 min):
   - current-state-analysis (what exists vs needed)
   - codebase-pattern-analysis (find reusable patterns)
   - file-structure-mapping (understand organization)
   - dependency-research (identify libraries)
   - api-context-gathering (document APIs)
   - integration-point-mapping (map connections)
   - technical-research (best practices)
3. **Create build plan** with architect-planner (10-15 min):
   - Design technical architecture
   - Break into buildable components
   - Sequence by dependencies
   - Organize into parallel build waves
4. **Direct builder team** in parallel waves (varies by PRD size):
   - Wave 1: Foundation team (database, types, interfaces)
   - Wave 2: Core services team (business logic, services)
   - Wave 3: Integration team (APIs, workflows, CLI)
   - Wave 4: Quality team (testing and polish)
   - Each wave: multiple builder agents working in parallel
5. **Quality assurance** (15-30 min):
   - Run all tests
   - Perform integration testing
   - Verify PRD requirements
   - Assess build quality
6. **Deliver** with comprehensive build report (10 min)

**Total Time**:
- Small PRD (1-2 weeks): 3-5 hours
- Medium PRD (3-8 weeks): 8-14 hours
- Large PRD (8-20 weeks): 23-43 hours

### Output

Complete **Working Software** with:
- **Source Code**: All files created/modified with proper Go conventions
- **Tests**: Comprehensive test suite with >80% coverage
- **Documentation**: Godoc comments, README updates, architecture docs
- **Build**: Fully buildable, all tests passing
- **Build Report**:
  - Requirements completion checklist (P0/P1/P2)
  - Build statistics (files, LOC, coverage)
  - Quality metrics (tests, linting, performance)
  - Outstanding issues and blockers
  - Next steps and recommendations

### Key Features

- **Team-Based Building**: PRD Builder coordinates specialized builder agents
- **Intelligent Coordination**: Maximizes parallelization across 10+ agents
- **Context-Aware**: Research team gathers comprehensive codebase understanding
- **Quality-Driven**: Tests, linting, and validation built-in
- **Wave-Based Building**: Sequences work by dependencies, parallel execution within waves
- **Production-Ready**: Code follows conventions, includes tests, builds successfully
- **Progress Tracking**: Clear visibility into build progress and team status
- **Adaptive**: Handles blockers, adjusts plans, resolves issues automatically

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    PRD Builder                               │
│  - Analyzes what needs to be built                           │
│  - Assembles and directs specialized agent teams             │
│  - Monitors build progress and quality                       │
└─────────────────┬───────────────────────────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
  Research Team        Builder Team
    (Parallel)          (Wave-Based)
        │                   │
        │                   │
┌───────▼──────────┐  ┌────▼─────────────────────────────────┐
│  7 Research      │  │  Wave 1: Foundation Team             │
│  Agents          │  │  ├─ builder-agent (DB schema)        │
│  (Concurrent)    │  │  ├─ builder-agent (types)            │
│                  │  │  └─ builder-agent (config)           │
│  1. current-     │  │                                      │
│     state        │  │  Wave 2: Core Services Team          │
│  2. codebase     │  │  ├─ builder-agent (parser)           │
│  3. file-struct  │  │  ├─ builder-agent (enricher)         │
│  4. dependencies │  │  └─ builder-agent (query)            │
│  5. api-context  │  │                                      │
│  6. integration  │  │  Wave 3: Integration Team            │
│  7. technical    │  │  ├─ builder-agent (API)              │
│                  │  │  ├─ builder-agent (CLI)              │
│  Runs: 5-10 min  │  │  └─ builder-agent (workflows)        │
└──────────────────┘  │                                      │
                      │  Wave 4: Quality Team                │
                      │  ├─ builder-agent (tests)            │
                      │  └─ builder-agent (docs)             │
                      │                                      │
                      │  Each wave: Team members work in     │
                      │  parallel, wave completes together   │
                      └──────────────────────────────────────┘
```

### Build Example

```
User: /build-prd SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md

PRD Builder:
├─ Phase 1: Analyzing PRD (5 min)
│  └─ Identified: 16 components, 47 tasks
│
├─ Phase 2: Assembling Research Team (7 min parallel)
│  ├─ current-state-analysis → Gap analysis complete
│  ├─ codebase-pattern-analysis → Patterns found
│  ├─ file-structure-mapping → Structure mapped
│  ├─ dependency-research → 13 deps needed
│  ├─ api-context-gathering → APIs documented
│  ├─ integration-point-mapping → Integrations mapped
│  └─ technical-research → Best practices gathered
│
├─ Phase 3: Creating Build Plan (12 min)
│  └─ architect-planner → 4 waves, 47 tasks
│
├─ Phase 4: Directing Builder Team
│  │
│  ├─ Wave 1: Foundation (2 hours, 8 parallel)
│  │  ├─ Task 1-8: Database schema ✓
│  │  ├─ Task 9-12: Core types ✓
│  │  └─ All tests passing (87% coverage)
│  │
│  ├─ Wave 2: Services (3 hours, 10 parallel)
│  │  ├─ Task 13-22: Core logic ✓
│  │  └─ All tests passing (91% coverage)
│  │
│  ├─ Wave 3: Integration (2.5 hours, 8 parallel)
│  │  ├─ Task 23-36: APIs, CLI, workflows ✓
│  │  └─ All tests passing (89% coverage)
│  │
│  └─ Wave 4: Polish (1 hour, 4 parallel)
│     ├─ Task 37-42: Tests and docs ✓
│     └─ All tests passing (92% coverage)
│
├─ Phase 5: Validation (25 min)
│  ├─ Integration tests: ✓ 47/47 passing
│  ├─ Build: ✓ Successful
│  └─ Requirements: ✓ 16/16 complete
│
└─ Phase 6: Completion Report
   └─ Delivered: Working implementation with 92% coverage
```

---

## Key Design Principles

### 1. Concurrent Fan-Out Pattern
- Launch ALL research agents in **single message** with multiple Task calls
- Maximizes parallelization (up to 10 agents simultaneously)
- Reduces total time by ~90% vs sequential execution

### 2. Dual Research Strategy
- **Local repo context**: Codebase patterns, structure, APIs, integrations
- **Web-based context**: Market data, competitors, users, technical best practices
- Combines internal and external intelligence

### 3. Tool Combination
- **Local research**: Glob, Grep, Read, Bash
- **Web research**: WebSearch (discovery) + WebFetch (deep analysis)
- **Orchestration**: Task tool with subagent_type="Explore"

### 4. Model Selection
- **Research agents**: Haiku (fast, cost-effective)
- **Writer agents**: Sonnet (higher quality output)
- **Orchestrators**: Sonnet (complex coordination)

### 5. Iterative Refinement
- Generate → Review → Refine → Repeat
- Quality improves with each cycle
- Typically 1-2 cycles sufficient

---

## Usage Examples

### Context Research

```
/research-context

I need to add JWT authentication to our API. We need to:
- Protect all /api/* endpoints
- Support refresh tokens
- Store tokens securely
```

**Output**: Complete context for implementation including:
- Similar auth patterns in codebase
- Where to place auth middleware
- Dependencies needed (JWT libraries)
- Existing APIs to protect
- Integration with current auth system

---

### PRD Generation

```
/create-prd

I want to build a feature that lets users export their data to CSV, Excel, and PDF formats.
```

**Orchestrator asks clarifying questions**, then:

**Launches 10 agents in parallel**:
- Local: Analyzes codebase for export patterns, file structure, dependencies, APIs
- Web: Researches market demand, competitors, technical approaches, user needs

**Generates PRD** with:
- Market data on export feature adoption
- Competitor export feature comparison
- User pain points with current solutions
- Technical implementation approach
- Codebase integration points
- Go-to-market strategy

---

### Implementation Planning

```
/plan-implementation

[Provide PRD or feature description]
```

**Agent workflow**:
1. Ingests PRD and analyzes requirements
2. Identifies knowledge gaps
3. Launches research agents if needed (optional, parallel)
4. Designs technical architecture
5. Breaks down into granular tasks
6. Creates implementation roadmap

**Generates Implementation Plan** with:
- Complete system architecture with diagrams

---

### PRD Building 🆕

```
/build-prd

SPECTRA_RED_PRD_ENGINEERING_FOCUSED.md
```

**PRD Builder workflow**:
1. Analyzes what needs to be built (5-10 min)
2. Assembles research team - 7 agents in parallel (5-10 min)
3. Creates build plan with architect-planner (10-15 min)
4. Directs builder teams in parallel waves (varies):
   - Wave 1: Foundation team - 8 parallel builders
   - Wave 2: Services team - 10 parallel builders
   - Wave 3: Integration team - 8 parallel builders
   - Wave 4: Quality team - 4 parallel builders
5. Quality assurance validation (15-30 min)
6. Generates build report (10 min)

**Delivers Working Software** with:
- All source files created/modified
- Comprehensive test suite (>80% coverage)
- Complete documentation
- Build passing, all tests green
- Build report with metrics
- Component specifications (purpose, interfaces, data models)
- 20-50 granular tasks with:
  - Step-by-step implementation guidance
  - File paths to create/modify
  - Dependencies and APIs to use
  - Testing approach
  - Estimated effort
- Milestones with success criteria
- Task sequencing and dependencies
- Timeline visualization
- Risk assessment
- Testing and deployment strategy

---

## Complete Workflow Examples

### Full Product Development Cycle (Idea → Working Code)

```bash
# Step 1: Create PRD
/create-prd
> Build analytics dashboard with user engagement metrics

# Gets: Market-researched PRD with competitive analysis

# Step 2: Build from PRD (NEW! 🆕)
/build-prd
> [Paste the PRD from Step 1]

# The PRD Builder assembles teams and builds the software:
#   - Research team gathers context (7 agents, 7 min)
#   - Architect plans the build (15 min)
#   - Builder teams work in parallel waves (8-14 hours)
#   - Quality team validates (30 min)
#
# Gets: Complete working software
#   - All source code files
#   - Comprehensive test suite
#   - Documentation
#   - Passing build
# Total time: ~8-14 hours (parallel team execution)
```

### Traditional Development Cycle (Planning → Manual Implementation)

```bash
# Step 1: Create PRD
/create-prd
> Build analytics dashboard with user engagement metrics

# Step 2: Plan Implementation
/plan-implementation
> [Paste the PRD from Step 1]

# Gets: Technical architecture + 30 sequenced tasks

# Step 3: Research Context (for individual tasks)
/research-context
> Implement task T-5: Dashboard data API endpoint

# Gets: Codebase patterns, integration points, implementation guide

# Step 4: Manual implementation by developer
```

---

## File Structure

```
.claude/
├── README.md (this file)
│
├── agents/
│   ├── context-orchestrator.md
│   ├── prd-orchestrator.md
│   ├── architect-planner.md
│   ├── prd-builder.md 🆕
│   └── builder-agent.md 🆕
│
├── skills/
│   ├── codebase-pattern-analysis/
│   │   └── SKILL.md
│   ├── file-structure-mapping/
│   │   └── SKILL.md
│   ├── dependency-research/
│   │   └── SKILL.md
│   ├── api-context-gathering/
│   │   └── SKILL.md
│   ├── requirements-analysis/
│   │   └── SKILL.md
│   ├── integration-point-mapping/
│   │   └── SKILL.md
│   ├── current-state-analysis/ 🆕
│   │   └── SKILL.md
│   ├── market-research/
│   │   └── SKILL.md
│   ├── competitor-analysis/
│   │   └── SKILL.md
│   ├── technical-research/
│   │   └── SKILL.md
│   ├── user-research/
│   │   └── SKILL.md
│   ├── prd-writer/
│   │   └── SKILL.md
│   └── prd-reviewer/
│       └── SKILL.md
│
└── commands/
    ├── research-context.md
    ├── create-prd.md
    ├── plan-implementation.md
    └── build-prd.md 🆕
```

---

## Research Citations

This system was built based on research from:

### Multi-Agent Orchestration
- Anthropic's multi-agent research system (90% time reduction with parallel agents)
- Cuong Tham's Claude Code subagent deep dive
- Zach Wills' parallel development patterns

### PRD Best Practices
- Miqdad Jaffer (OpenAI Product Lead) AI PRD Template
- IBM's MetaGPT multi-agent PRD automation
- Kovyrin's PRD-driven development workflow

### Claude Code Patterns
- Lee Hanchung's Claude skills deep dive
- Mikhail Shilkov's web tools analysis
- Claude Agent SDK documentation

---

## Performance Characteristics

### Context Research System
- **Agents**: 6 concurrent
- **Time**: 5-7 minutes
- **Token Usage**: ~150k-200k tokens
- **Output**: 3000-5000 word context document

### PRD Generation System
- **Agents**: 10 concurrent + 2 sequential (writer/reviewer)
- **Time**: 10-18 minutes
- **Token Usage**: ~250k-350k tokens
- **Output**: 5000-10000 word comprehensive PRD

### PRD Builder System 🆕
- **Agents**: 7 concurrent (research team) + 4-30 concurrent (builder team waves)
- **Time**:
  - Small PRD (1-2 weeks): 3-5 hours
  - Medium PRD (3-8 weeks): 8-14 hours
  - Large PRD (8-20 weeks): 23-43 hours
- **Token Usage**: ~500k-2M tokens (varies by PRD complexity)
- **Output**:
  - Complete working software
  - Test suite with >80% coverage
  - Documentation and build report
  - 1000-10000+ lines of production code

---

## Customization

### Adding New Research Skills

1. Create skill directory: `.claude/skills/[skill-name]/`
2. Add `SKILL.md` with:
   - Objective
   - Input required
   - Research process
   - Output format
3. Update orchestrator to include in parallel launch

### Modifying PRD Template

Edit `.claude/skills/prd-writer/SKILL.md` to:
- Change section structure
- Adjust output format
- Modify quality standards

### Tuning Performance

- **Speed**: Use haiku for all agents
- **Quality**: Use sonnet for key agents (writer, reviewer)
- **Cost**: Reduce number of agents or use haiku throughout
- **Depth**: Add more web research iterations

---

## Best Practices

1. **Always launch agents in parallel** - Use single message with multiple Task calls
2. **Combine local + web research** - Never skip either stream
3. **Iterate on quality** - Use review → refine cycles
4. **Cite everything** - Include sources for all claims
5. **Be specific** - Quantify, don't generalize
6. **Use appropriate models** - Haiku for research, Sonnet for synthesis

---

## Troubleshooting

### Agents Not Running in Parallel
- Ensure single message with multiple Task tool calls
- Check that subagent_type="Explore"
- Verify no dependencies between agents

### Research Quality Issues
- Add more specific prompts to agent tasks
- Increase iteration cycles for review/refinement
- Use WebFetch for deeper analysis of promising sources

### Output Too Long
- Reduce number of agents
- Constrain research scope in prompts
- Use haiku model for all agents

### Missing Context
- Add more research agents
- Specify more detailed research objectives
- Include follow-up targeted research for gaps

---

## Future Enhancements

Potential additions:
- **Financial Analysis Agent** - Revenue models, pricing strategy
- **Risk Analysis Agent** - Technical and business risks
- **Legal/Compliance Agent** - Regulatory requirements
- **Data Analysis Agent** - Analytics and metrics strategy
- **Design Research Agent** - UI/UX patterns and best practices

---

## License

This agent system was created for use within the Recon project's Conductor workspace.

---

## Changelog

### v2.0.0 (2025-11-01) 🆕
- **PRD Builder System** - Build complete software from PRDs using intelligent agent teams
- Added `prd-builder` agent - Assembles and coordinates specialized builder agent teams
- Added `builder-agent` - Team member who builds specific components from specifications
- Added `current-state-analysis` skill - Analyzes what exists vs what needs to be built
- Added `/build-prd` command - Build complete software from a PRD
- Enhanced architecture with 4 major systems (was 3)
- Total: 5 agents, 13 skills, 4 slash commands

### v1.0.0 (2025-11-01)
- Initial implementation
- Context Research System (6 agents)
- PRD Generation System (10 research + 2 synthesis agents)
- Implementation Planning System
- 12 total skills created
- 3 slash commands
- Complete documentation
