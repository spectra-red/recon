# Claude Code Agent & Skill System

This repository contains a comprehensive multi-agent system for automated context research and PRD generation, built for use with Claude Code.

## System Overview

Two major agent orchestration systems:

### 1. Context Research System
**Purpose**: Gather comprehensive implementation context for features/tasks

**Command**: `/research-context`

**Architecture**: Orchestrator + 6 concurrent research agents

### 2. PRD Generation System
**Purpose**: Create production-ready Product Requirements Documents

**Command**: `/create-prd`

**Architecture**: Orchestrator + 10 concurrent research agents (6 local + 4 web)

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

## File Structure

```
.claude/
├── README.md (this file)
│
├── agents/
│   ├── context-orchestrator.md
│   └── prd-orchestrator.md
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
    └── create-prd.md
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

### v1.0.0 (2025-11-01)
- Initial implementation
- Context Research System (6 agents)
- PRD Generation System (10 research + 2 synthesis agents)
- 12 total skills created
- 2 slash commands
- Complete documentation
