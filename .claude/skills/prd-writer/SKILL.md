# PRD Writer Skill

This skill enables agents to synthesize research findings into a comprehensive, well-structured Product Requirements Document (PRD) following industry best practices.

## Objective

Transform research findings (market research, competitive analysis, technical research, user research, and codebase analysis) into a complete, actionable PRD following OpenAI's proven 9-section template.

## Input Required

- **Product/Feature Description**: What is being built
- **Research Findings**: All research from various agents:
  - Market research
  - Competitive analysis
  - Technical research
  - User research
  - Requirements analysis
  - Codebase pattern analysis
  - File structure mapping
  - Dependency research
  - API context
  - Integration point mapping

## PRD Template Structure (OpenAI Standard)

### Section 1: Executive Summary
- Strategic alignment and why this matters
- Market opportunity in 2-3 sentences
- Key success metrics
- High-level solution approach

### Section 2: Market Opportunity
- Market size (TAM, SAM, SOM) with sources
- Market growth rate (CAGR) with sources
- Market stage assessment (emerging/growing/mature)
- Long-term business value

### Section 3: Strategic Alignment
- Company vision alignment
- Product objectives this supports
- Organizational fit and competencies
- Why we're positioned to build this

### Section 4: Customer & User Needs
- Target segments and personas
- Jobs-to-be-done
- Pain points with evidence
- User research findings

### Section 5: Value Proposition & Messaging
- Problem statement
- Solution overview
- Key benefits and outcomes
- Differentiation vs competitors
- Messaging framework per segment

### Section 6: Competitive Advantage
- Competitive landscape overview
- Our defensible advantages (MOATs)
- Sustainability assessment
- Competitor comparison

### Section 7: Product Scope and Use Cases
- Core capabilities and features
- Primary use cases with flows
- Feature specifications with acceptance criteria
- Out of scope
- High-risk assumptions

### Section 8: Non-Functional Requirements
- Performance, security, reliability, usability, maintainability
- AI-specific requirements (if applicable): accuracy, monitoring, ethics
- Technical implementation notes from codebase analysis

### Section 9: Go-to-Market Approach
- Release phases (MVP to full launch)
- Target segments and rollout plan
- Success metrics
- Feedback loops

## Writing Process

### 1. Organize Research Findings

**Review all input research:**
- Group by PRD section
- Identify key insights
- Note data sources for citations
- Flag any gaps

### 2. Write Executive Summary

**Craft a compelling 3-paragraph summary:**
- **Paragraph 1**: What we're building and why (the vision)
- **Paragraph 2**: Market opportunity and business value (the case)
- **Paragraph 3**: Approach and success metrics (the plan)

**Tips:**
- Make it standalone (executives may only read this)
- Lead with impact
- Quantify the opportunity
- Be concise but compelling

### 3: Write Market Opportunity

**Synthesize market research findings:**
- Present market size data with citations
- Show growth trajectory with data
- Assess market stage with evidence
- Articulate long-term value

**Quality checklist:**
- [ ] TAM, SAM, SOM all defined with sources
- [ ] CAGR with time period and source
- [ ] Market stage clearly identified
- [ ] All claims backed by credible sources

### 4. Write Strategic Alignment

**Connect to company strategy:**
- How this fits company vision
- Which OKRs/goals this supports
- Why we're positioned to win
- Strategic rationale

**Use information from:**
- User requirements/context
- Company strategy (if provided)
- Competitive advantage research

### 5. Write Customer & User Needs

**Synthesize user research:**
- Target segments (2-4 max)
- User personas with quotes
- Jobs-to-be-done framework
- Pain points with severity and evidence

**Quality checklist:**
- [ ] Personas feel real and specific
- [ ] JTBD includes functional, emotional, social
- [ ] Pain points backed by user quotes
- [ ] Clear priority/severity indicated

### 6. Write Value Proposition & Messaging

**Craft clear positioning:**
- Problem statement (specific, relatable)
- Solution overview (how we solve it)
- Key benefits (outcomes, not features)
- Differentiation (why choose us)
- Messages tailored per segment

**Use insights from:**
- User pain points
- Competitive analysis
- Jobs-to-be-done

### 7. Write Competitive Advantage

**Synthesize competitive research:**
- Landscape overview
- Key competitors profiled
- Our MOATs (defensible advantages)
- Sustainability assessment

**Quality checklist:**
- [ ] All claims evidence-based
- [ ] MOATs are truly defensible
- [ ] Competitor data accurate
- [ ] Realistic about challenges

### 8. Write Product Scope and Use Cases

**Define what we're building:**
- Core capabilities list
- Primary use cases (3-5)
  - User story format
  - Step-by-step flow
  - Success criteria
- Feature specifications
  - Description
  - Acceptance criteria
  - Technical approach (from codebase research)
- Explicitly out of scope
- High-risk assumptions to validate

**Use information from:**
- Requirements analysis
- User needs
- Technical research
- Codebase patterns

### 9. Write Non-Functional Requirements

**Cover all quality attributes:**
- **Performance**: Targets with measurement methods
- **Security**: Authentication, authorization, data protection
- **Reliability**: Uptime, error handling, recovery
- **Usability**: Accessibility, UX standards
- **Maintainability**: Code quality, testing, documentation
- **AI-Specific** (if applicable):
  - Accuracy and reliability targets
  - Evaluation metrics
  - Monitoring and drift detection
  - Ethical guardrails

**Technical Implementation Notes:**
- Patterns to follow (from codebase analysis)
- File placement (from file structure mapping)
- Dependencies needed (from dependency research)
- APIs to use (from API context)
- Integration points (from integration mapping)

### 10. Write Go-to-Market Approach

**Plan the rollout:**
- **Phase 1: MVP**
  - Core features only
  - Target early adopters
  - Success metrics
- **Phase 2+**: Expansion phases
- Launch strategy
- Feedback mechanisms

**Use information from:**
- Market research (segments, timing)
- User research (early adopters)
- Competitive analysis (market positioning)

### 11. Add Appendix

**Include supporting information:**
- Research sources (all URLs)
- Open questions
- Risks and mitigations table
- Revision history

## Writing Best Practices

### Style Guidelines

**Be Clear:**
- Use simple, direct language
- Avoid jargon unless necessary
- Define technical terms
- Short sentences and paragraphs

**Be Specific:**
- Quantify everything possible
- Use concrete examples
- Provide acceptance criteria
- Include success metrics

**Be Actionable:**
- Focus on "what" and "why," not just "how"
- Make requirements testable
- Provide clear priorities
- Enable immediate implementation

**Be Evidence-Based:**
- Cite all data sources
- Include user quotes
- Reference research
- Link to codebase examples

### Formatting Best Practices

**Use Markdown Effectively:**
- Clear heading hierarchy (##, ###, ####)
- Tables for comparisons
- Lists for features/requirements
- Code blocks for technical details
- Checkboxes for acceptance criteria

**Make It Scannable:**
- Bold key points
- Use tables for dense data
- Break up long paragraphs
- Add whitespace

**Include Visual Elements:**
- ASCII diagrams where helpful
- Data in tables
- Comparisons side-by-side

### Citation Best Practices

**Always Cite Sources:**
- Include URL after claims
- Group sources at section end
- Note date of data
- Indicate confidence level

**Citation Formats:**
- Inline: "Market growing at 23% CAGR [Source: Gartner 2025]"
- After paragraph: **Source**: [https://...]
- Section end: **Sources**: [List of URLs]

## Quality Checklist

Before finalizing, verify:

### Completeness
- [ ] All 9 sections present
- [ ] Each section has substantive content
- [ ] All research incorporated
- [ ] Sources cited throughout

### Clarity
- [ ] Executive summary is standalone
- [ ] Technical terms defined
- [ ] Acceptance criteria clear
- [ ] Priorities indicated

### Data Quality
- [ ] Market data is recent (2024-2025)
- [ ] All numbers have sources
- [ ] User quotes included
- [ ] Competitor data accurate

### Actionability
- [ ] Features have acceptance criteria
- [ ] Success metrics defined
- [ ] Priorities clear
- [ ] Implementation guidance included

### Technical Completeness
- [ ] Non-functional requirements comprehensive
- [ ] Technical approach documented
- [ ] Integration points mapped
- [ ] Risks identified

### Strategic Soundness
- [ ] Market opportunity compelling
- [ ] Competitive advantage defensible
- [ ] Go-to-market realistic
- [ ] Success measurable

## Common Pitfalls to Avoid

1. **Too Vague**: "Fast performance" → "< 200ms response time"
2. **Missing Sources**: Claims without citations
3. **Feature List**: PRD is not just a feature list
4. **No Priorities**: Everything can't be P0
5. **Ignoring Research**: Use the research provided
6. **Too Long**: Aim for clarity over comprehensiveness
7. **No Success Criteria**: How do we know when we're done?
8. **Weak Competitive Advantage**: MOATs must be defensible
9. **Unrealistic Timeline**: Be honest about scope
10. **Missing Non-Functionals**: Performance, security matter

## Output Template

Use this structure for the final PRD:

````markdown
# Product Requirements Document: [Feature Name]

**Version:** 1.0
**Date:** [Date]
**Author:** PRD Writer Agent
**Status:** Draft

---

## 1. Executive Summary

[3 compelling paragraphs]

---

## 2. Market Opportunity

### Market Size & Growth
[TAM, SAM, SOM with sources]

### Market Stage
[Assessment with evidence]

### Long-term Business Value
[Strategic value]

**Sources:**
- [URLs]

---

## 3. Strategic Alignment

### Company Vision Alignment
[How this fits]

### Product Objectives
[Which goals this supports]

### Organizational Fit
[Why we can win]

---

## 4. Customer & User Needs

### Target Segments
[2-4 segments defined]

### User Personas
[Detailed personas with quotes]

### Jobs-to-be-Done
[JTBD framework applied]

### Pain Points
[Evidence-based pain points]

---

## 5. Value Proposition & Messaging

### Problem Statement
[Clear problem articulation]

### Solution Overview
[How we solve it]

### Key Benefits
[Customer outcomes]

### Differentiation
[Why choose us]

### Messaging Framework
[Messages per segment]

---

## 6. Competitive Advantage

### Competitive Landscape
[Overview with key players]

### Our Defensible Advantages (MOATs)
[Specific, sustainable advantages]

### Competitor Comparison
[Table or detailed comparison]

**Sources:**
- [URLs]

---

## 7. Product Scope and Use Cases

### Core Capabilities
[Feature list]

### Primary Use Cases
[Detailed use cases with flows]

### Feature Specifications
[Features with acceptance criteria]

### Out of Scope
[Explicitly not included]

### High-Risk Assumptions
[Assumptions to validate]

---

## 8. Non-Functional Requirements

### Performance
[Targets with measurement]

### Security
[Authentication, authorization, protection]

### Reliability
[Uptime, error handling]

### Usability
[UX standards]

### Maintainability
[Code quality, testing]

### AI-Specific (if applicable)
[Accuracy, monitoring, ethics]

### Technical Implementation Notes
[From codebase research]

---

## 9. Go-to-Market Approach

### Release Phases
[MVP and beyond]

### Launch Strategy
[Rollout plan]

### Success Metrics
[How to measure success]

### Feedback Loops
[How to iterate]

---

## Appendix

### Research Sources
[All URLs organized]

### Open Questions
[Items needing clarification]

### Risks & Mitigations
[Risk table]

### Revision History
[Version tracking]
````

## Success Criteria

A successful PRD provides:
- Compelling executive summary that sells the vision
- Quantified market opportunity with credible sources
- Clear strategic rationale
- Deep user understanding with evidence
- Strong value proposition and differentiation
- Defensible competitive advantages
- Detailed product scope with acceptance criteria
- Comprehensive non-functional requirements
- Realistic go-to-market plan
- All research synthesized and cited
- Actionable for implementation
- Professional formatting and clarity
