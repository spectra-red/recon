# PRD Reviewer Skill

This skill enables agents to critically review Product Requirements Documents (PRDs) and provide constructive feedback to improve quality, completeness, and actionability.

## Objective

Review PRDs systematically to identify gaps, ambiguities, inconsistencies, and areas for improvement, providing specific, actionable feedback to strengthen the document.

## Input Required

- **PRD Document**: The PRD to review (complete or draft)
- **Context**: Any additional context about the product, market, or organization

## Review Framework

### 1. Completeness Review

**Check that all required sections are present and substantive:**

- [ ] Executive Summary
- [ ] Market Opportunity
- [ ] Strategic Alignment
- [ ] Customer & User Needs
- [ ] Value Proposition & Messaging
- [ ] Competitive Advantage
- [ ] Product Scope and Use Cases
- [ ] Non-Functional Requirements
- [ ] Go-to-Market Approach
- [ ] Appendix (Sources, Open Questions, Risks)

**For each section, verify:**
- Not just a placeholder or single sentence
- Contains substantive, actionable content
- Backed by evidence or research

### 2. Clarity Review

**Assess readability and comprehension:**

**Executive Summary:**
- [ ] Can standalone (someone could read only this)
- [ ] Compelling and clear
- [ ] Covers what, why, and success metrics
- [ ] 2-3 paragraphs, not too long

**Overall Document:**
- [ ] Language is clear and direct
- [ ] Technical terms are defined
- [ ] Jargon is minimized
- [ ] Sentences are concise
- [ ] Paragraphs are scannable

**Acceptance Criteria:**
- [ ] All features have clear, testable acceptance criteria
- [ ] Success criteria are measurable
- [ ] Requirements are unambiguous

### 3. Data Quality Review

**Verify all claims are evidence-based:**

**Market Data:**
- [ ] Market size numbers have sources
- [ ] Growth rates have sources
- [ ] Market data is recent (2024-2025)
- [ ] TAM, SAM, SOM methodology explained
- [ ] Claims are realistic and credible

**User Research:**
- [ ] User quotes included
- [ ] Pain points have evidence
- [ ] Personas feel specific and real
- [ ] Needs are prioritized with rationale

**Competitive Analysis:**
- [ ] Competitor data is accurate
- [ ] Claims about competitors are fair
- [ ] Differentiation is specific
- [ ] MOATs are truly defensible

**Technical:**
- [ ] Performance targets are specific
- [ ] Technology choices have rationale
- [ ] Implementation approach is feasible

**Citations:**
- [ ] All major claims have sources
- [ ] URLs are included
- [ ] Sources are credible
- [ ] Data dates are noted

### 4. Strategic Soundness Review

**Assess the strategic logic:**

**Market Opportunity:**
- [ ] Opportunity is compelling
- [ ] Market is large enough
- [ ] Growth is sustainable
- [ ] Timing is right

**Strategic Alignment:**
- [ ] Fits company strategy
- [ ] Leverages organizational strengths
- [ ] Addresses real strategic need

**Competitive Advantage:**
- [ ] MOATs are defensible (not just features)
- [ ] Advantage is sustainable over time
- [ ] Barriers to replication are real
- [ ] Competitive positioning is realistic

**Go-to-Market:**
- [ ] MVP scope is realistic
- [ ] Target segments are specific
- [ ] Success metrics are appropriate
- [ ] Timeline is achievable

### 5. User-Centricity Review

**Verify deep user understanding:**

**Personas:**
- [ ] Feel real and specific
- [ ] Include demographics, goals, pain points
- [ ] Based on research, not stereotypes
- [ ] 2-4 personas (not too many)

**Jobs-to-be-Done:**
- [ ] Include functional, emotional, social jobs
- [ ] Context and triggers are clear
- [ ] Success criteria defined
- [ ] Current solutions analyzed

**Pain Points:**
- [ ] Specific and detailed
- [ ] Severity/frequency indicated
- [ ] Backed by user quotes
- [ ] Prioritized by impact

**Value Proposition:**
- [ ] Addresses real user needs
- [ ] Benefits are outcome-focused
- [ ] Differentiation is meaningful to users
- [ ] Messaging resonates with personas

### 6. Technical Feasibility Review

**Assess implementation realism:**

**Scope:**
- [ ] Features are clearly defined
- [ ] Scope is achievable
- [ ] Out of scope is explicit
- [ ] MVP is truly minimal

**Non-Functional Requirements:**
- [ ] Performance targets are specific
- [ ] Security requirements are comprehensive
- [ ] Scalability is addressed
- [ ] Monitoring plan is included

**Technical Approach:**
- [ ] Technology choices are justified
- [ ] Architecture is appropriate
- [ ] Integration points are mapped
- [ ] Dependencies are identified

**Risks:**
- [ ] Technical risks are identified
- [ ] Mitigation strategies are defined
- [ ] Assumptions are called out
- [ ] Complexity is acknowledged

### 7. Actionability Review

**Ensure the PRD enables action:**

**For Product Team:**
- [ ] Clear what to build
- [ ] Features have acceptance criteria
- [ ] Priorities are indicated
- [ ] Design guidance is provided

**For Engineering Team:**
- [ ] Technical requirements are clear
- [ ] Integration points are defined
- [ ] Performance targets are specific
- [ ] Non-functionals are comprehensive

**For Go-to-Market Team:**
- [ ] Target segments are defined
- [ ] Value proposition is clear
- [ ] Messaging is provided
- [ ] Success metrics are specified

**For Stakeholders:**
- [ ] Business case is clear
- [ ] Success criteria are defined
- [ ] Risks are identified
- [ ] Timeline is realistic

### 8. Consistency Review

**Check for internal consistency:**

**Cross-Section Alignment:**
- [ ] Features align with user needs
- [ ] Value prop aligns with pain points
- [ ] Success metrics align with goals
- [ ] MVP aligns with GTM strategy

**No Contradictions:**
- [ ] Different sections don't contradict
- [ ] Priorities are consistent
- [ ] Messaging is consistent

## Review Output Format

```markdown
## PRD Review Feedback

### Overall Assessment

**Quality Rating**: [Excellent / Good / Needs Work / Insufficient]

**Summary**: [2-3 sentence overall assessment]

**Strengths**:
1. [Major strength 1]
2. [Major strength 2]
3. [Major strength 3]

**Critical Issues**:
1. [Must-fix issue 1]
2. [Must-fix issue 2]

**Recommendation**: [Ready / Needs Revision / Needs Major Rework]

---

### Section-by-Section Feedback

#### 1. Executive Summary

**Rating**: [Strong / Adequate / Weak]

**Strengths**:
- [Specific strength]

**Issues**:
- **[Severity: Critical/Major/Minor]** [Specific issue]
  - **Impact**: [Why this matters]
  - **Recommendation**: [How to fix]

**Suggestions**:
- [Optional improvement]

---

#### 2. Market Opportunity

**Rating**: [Strong / Adequate / Weak]

**Strengths**:
- [Specific strength]

**Issues**:
- **[Severity]** [Issue]
  - **Impact**: [Why this matters]
  - **Recommendation**: [How to fix]

**Missing**:
- [What's not covered but should be]

**Data Quality**:
- [Assessment of sources and credibility]

---

[Continue for all 9 sections...]

---

### Completeness Assessment

**Sections Present**: [X/9]

**Missing Sections**:
- [Section name]: [Why it's needed]

**Underdeveloped Sections**:
- [Section name]: [What's missing]

---

### Clarity Assessment

**Overall Clarity**: [Excellent / Good / Fair / Poor]

**Clarity Issues**:
1. **[Location]**: [Unclear statement or section]
   - **Problem**: [Why it's unclear]
   - **Suggestion**: [How to improve]

2. **[Location]**: [Issue]
   - **Problem**: [Description]
   - **Suggestion**: [Fix]

**Jargon Issues**:
- [Term]: Needs definition or simpler language
- [Term]: Unclear to non-technical readers

**Ambiguous Requirements**:
- [Requirement]: [Why it's ambiguous]
  - **Make Specific**: [How to clarify]

---

### Data Quality Assessment

**Overall Data Quality**: [Excellent / Good / Fair / Poor]

**Well-Supported Claims**:
- [Claim 1]: Strong evidence from [source]
- [Claim 2]: Credible data from [source]

**Unsupported Claims**:
- **[Claim]**: [Location in PRD]
  - **Issue**: No source provided
  - **Action**: Add source or remove claim

- **[Claim]**: [Location]
  - **Issue**: Source not credible
  - **Action**: Find better source

**Outdated Data**:
- [Data point]: From [year], need 2024-2025 data

**Missing Evidence**:
- [Assertion]: Needs user quotes or data
- [Claim]: Needs competitive evidence

**Citation Issues**:
- [Section]: Missing URLs for claims
- [Section]: Sources not credible

---

### Strategic Soundness Assessment

**Strategic Rating**: [Strong / Adequate / Weak]

**Market Opportunity**:
- ✓ **Strength**: [What's strong]
- ✗ **Concern**: [What's weak]
  - **Recommendation**: [How to address]

**Competitive Advantage**:
- **MOATs Assessment**:
  - [MOAT 1]: [Defensible? Sustainable?]
    - [Feedback]
  - [MOAT 2]: [Defensible? Sustainable?]
    - [Feedback]

**Strategic Alignment**:
- ✓ **Aligned**: [What fits]
- ? **Unclear**: [What needs clarification]

**Go-to-Market Realism**:
- ✓ **Realistic**: [What's achievable]
- ✗ **Concern**: [What's questionable]
  - **Recommendation**: [How to adjust]

---

### User-Centricity Assessment

**User Research Quality**: [Excellent / Good / Fair / Poor]

**Persona Quality**:
- ✓ **Good**: [What's well done]
- ✗ **Issue**: [What's weak]
  - **Improvement**: [How to strengthen]

**Jobs-to-be-Done**:
- [Assessment of JTBD quality]
- [Missing elements]

**Pain Point Analysis**:
- ✓ **Well-Documented**: [Strong pain points]
- ✗ **Weak**: [Insufficient pain points]
  - **Need**: More evidence, quotes, severity

**Value Proposition**:
- [Assessment of value prop strength]
- [Alignment with user needs]
- [Suggestions for improvement]

---

### Technical Feasibility Assessment

**Feasibility Rating**: [Feasible / Challenging / Unrealistic]

**Scope Realism**:
- ✓ **Realistic**: [What's achievable]
- ✗ **Concern**: [What's too ambitious]
  - **Recommendation**: [How to scope down]

**Non-Functional Requirements**:
- ✓ **Complete**: [What's well covered]
- ✗ **Missing**: [What's not addressed]
  - **Need**: [What to add]

**Technical Risks**:
- **Identified Risks**: [Assessment]
- **Missing Risks**: [Additional risks not mentioned]

**Implementation Clarity**:
- [Assessment of how clear the technical approach is]

---

### Actionability Assessment

**Actionability Rating**: [Highly Actionable / Adequate / Unclear]

**For Product Team**:
- ✓ **Clear**: [What's actionable]
- ✗ **Unclear**: [What needs clarification]

**For Engineering**:
- ✓ **Sufficient**: [What's well defined]
- ✗ **Gaps**: [What's missing]

**For Go-to-Market**:
- ✓ **Ready**: [What's usable]
- ✗ **Incomplete**: [What's needed]

**Acceptance Criteria Quality**:
- [Assessment of criteria clarity and testability]

**Priority Clarity**:
- [Are priorities clear? Must-have vs nice-to-have?]

---

### Consistency Assessment

**Internal Consistency**: [Consistent / Minor Issues / Major Contradictions]

**Alignment Issues**:
1. **[Contradiction]**: [Description]
   - **Location**: [Where in PRD]
   - **Fix**: [How to resolve]

**Priority Conflicts**:
- [Conflicting priorities identified]

**Messaging Consistency**:
- [Assessment of consistent messaging]

---

### Detailed Recommendations

#### Critical (Must Fix)

1. **[Issue 1]**
   - **Problem**: [Specific description]
   - **Impact**: [Why this is critical]
   - **Action**: [Specific fix needed]
   - **Location**: [Section/paragraph]

2. **[Issue 2]**
   [Same structure]

---

#### Major (Should Fix)

1. **[Issue 1]**
   - **Problem**: [Description]
   - **Impact**: [Why this matters]
   - **Action**: [How to fix]
   - **Location**: [Where]

---

#### Minor (Consider Fixing)

1. **[Issue 1]**
   - **Suggestion**: [Improvement]
   - **Benefit**: [Why it would help]

---

#### Enhancements (Optional)

1. **[Enhancement 1]**
   - **Idea**: [Improvement idea]
   - **Value**: [What it would add]

---

### Missing Elements

**Required but Missing**:
- [ ] [Element 1]: [Why it's required]
- [ ] [Element 2]: [Why it's required]

**Strongly Recommended**:
- [ ] [Element 1]: [Why it would strengthen PRD]
- [ ] [Element 2]: [Why it would strengthen PRD]

---

### Questions for Clarification

1. **[Question 1]**: [What needs clarification]
   - **Why It Matters**: [Impact on PRD]

2. **[Question 2]**: [What's ambiguous]
   - **Why It Matters**: [Impact on PRD]

---

### Strengths to Preserve

**Keep These**:
1. [Specific strength to maintain]
2. [Another strength]
3. [Another strength]

---

### Next Steps

**Before Final Approval**:
1. [Action 1]
2. [Action 2]
3. [Action 3]

**For Consideration**:
- [Optional improvement 1]
- [Optional improvement 2]

**Estimated Effort to Address Feedback**: [Small / Medium / Large]

---

### Reviewer Notes

**Review Date**: [Date]

**Reviewer**: PRD Reviewer Agent

**Review Methodology**: [Approach used]

**Confidence in Assessment**: [High / Medium / Low]

**Limitations of Review**: [Any caveats]
```

## Review Best Practices

### Be Constructive

**Do:**
- Explain why something is an issue
- Provide specific suggestions
- Acknowledge strengths
- Prioritize feedback (critical vs nice-to-have)

**Don't:**
- Just criticize without solutions
- Be vague ("this needs work")
- Ignore good parts
- Overwhelm with minor issues

### Be Specific

**Good Feedback:**
- "Section 2: Market size claim of $50B lacks source. Add citation to Gartner 2025 report or similar credible source."

**Bad Feedback:**
- "Need better sources"

### Prioritize Issues

**Critical:** Must fix before approval
- Missing required sections
- Unsupported major claims
- Strategic flaws
- Unclear requirements

**Major:** Should fix for quality
- Incomplete sections
- Weak evidence
- Minor inconsistencies
- Clarity issues

**Minor:** Consider fixing
- Formatting improvements
- Additional details
- Enhanced explanations

**Enhancements:** Nice to have
- Additional examples
- Visual improvements
- Extended research

### Focus on Impact

Always explain:
- **What** the issue is
- **Why** it matters
- **How** to fix it

### Be Fair and Balanced

- Acknowledge strengths
- Critique objectively
- Validate good work
- Provide constructive path forward

## Common PRD Issues to Watch For

### Completeness Issues
- Missing sections
- Placeholder content
- Underdeveloped ideas
- No sources cited

### Clarity Issues
- Vague requirements
- Ambiguous acceptance criteria
- Undefined jargon
- Complex sentences

### Data Quality Issues
- Unsupported claims
- Outdated data
- Unreliable sources
- Missing citations

### Strategic Issues
- Weak market opportunity
- Non-defensible advantages
- Poor strategic fit
- Unrealistic goals

### User Understanding Issues
- Generic personas
- Missing pain points
- No user evidence
- Weak value proposition

### Technical Issues
- Vague non-functionals
- Missing performance targets
- Unclear technical approach
- Unidentified risks

### Actionability Issues
- No acceptance criteria
- Unclear priorities
- Missing success metrics
- Insufficient detail

## Success Criteria

A successful PRD review provides:
- Overall quality assessment and recommendation
- Section-by-section detailed feedback
- Prioritized issues (critical, major, minor)
- Specific, actionable recommendations
- Identified strengths to preserve
- Questions that need clarification
- Clear next steps
- Constructive, helpful tone
- Balanced perspective (strengths and issues)
- Focused on high-impact improvements
