# Competitor Analysis Skill

This skill enables agents to conduct comprehensive competitive analysis using web research to understand the competitive landscape, identify key competitors, and discover differentiation opportunities.

## Objective

Research and analyze competitors to understand their offerings, positioning, strengths, weaknesses, and identify opportunities for competitive advantage and differentiation.

## Input Required

- **Product/Feature Description**: What we're building
- **Market/Category**: Industry or product category
- **Known Competitors**: Any competitors already identified (optional)

## Research Process

### 1. Identify Competitors

**Use WebSearch to discover competitors:**

**Direct Competitors:**
- "[product category] companies"
- "[problem space] solutions"
- "alternatives to [known competitor]"
- "best [product category] tools 2025"

**Indirect Competitors:**
- "[problem] alternative solutions"
- "[use case] different approaches"
- DIY or manual process alternatives

**Search in:**
- Product comparison sites (G2, Capterra, TrustRadius)
- "Alternative to" sites (AlternativeTo.net)
- Tech news sites (TechCrunch, ProductHunt)
- Review aggregators
- Industry analyst reports

### 2. Competitor Profile Research

**For each major competitor, use WebSearch + WebFetch to research:**

**Company Information:**
- Company size, founding date, funding
- Leadership and key team members
- Company strategy and vision
- Recent news and developments

**Product Offerings:**
- Core features and capabilities
- Product variants and tiers
- Platform and integrations
- Technology stack (if public)

**Market Positioning:**
- Target customer segments
- Value proposition and messaging
- Brand positioning
- Marketing channels and tactics

**Business Model:**
- Pricing structure
- Revenue model (SaaS, transaction, freemium, etc.)
- Customer acquisition strategy
- Partnership approach

**Performance Metrics:**
- Customer count (if public)
- Revenue (if public)
- Market share
- Growth rate
- Funding raised

### 3. Feature Comparison Research

**WebSearch for:**
- "[competitor] features list"
- "[competitor] vs [other competitor]"
- "[product category] comparison"
- "[competitor] review"

**Analyze:**
- Core features offered
- Unique capabilities
- Feature gaps
- User experience approach
- Technical differentiation

### 4. Positioning & Messaging Analysis

**Research how competitors position themselves:**

**WebFetch competitor websites for:**
- Homepage messaging
- Value propositions
- Target personas mentioned
- Use cases highlighted
- Differentiation claims

**Analyze:**
- Positioning strategy
- Messaging hierarchy
- Emotional vs rational appeals
- Visual brand identity
- Content marketing approach

### 5. Pricing Research

**WebSearch and WebFetch:**
- "[competitor] pricing"
- "[competitor] plans"
- "[product category] pricing comparison"

**Capture:**
- Pricing tiers and structure
- Free trial or freemium offerings
- Enterprise pricing (if available)
- Add-on or usage-based fees
- Discounting patterns

### 6. Strengths & Weaknesses Analysis

**Research customer sentiment:**

**WebSearch for reviews:**
- "[competitor] reviews G2"
- "[competitor] Capterra reviews"
- "[competitor] complaints"
- "[competitor] Reddit"
- "[competitor] alternatives why"

**Analyze:**
- Common praise points (strengths)
- Common complaints (weaknesses)
- Feature requests
- Switching reasons
- NPS or satisfaction scores

### 7. Competitive Advantage Research

**Identify defensible advantages:**

**Look for:**
- **Data advantages**: Proprietary data, network effects
- **Technology advantages**: Patents, unique algorithms, performance
- **Partnership advantages**: Exclusive integrations, channel partners
- **Scale advantages**: Economies of scale, market dominance
- **Brand advantages**: Recognition, trust, loyalty

### 8. Market Share & Momentum

**Research market position:**

**WebSearch:**
- "[category] market share"
- "[competitor] number of customers"
- "[competitor] growth rate"
- "[category] leader Gartner/Forrester"

**Indicators of momentum:**
- Recent funding rounds
- Customer acquisition trends
- Product launch velocity
- Media coverage volume
- Social media growth

## Output Format

```markdown
## Competitive Analysis

### Executive Summary

[2-3 paragraphs summarizing:
- Number of competitors identified
- Competitive landscape characterization (crowded, emerging, consolidated)
- Key competitive threats
- Differentiation opportunities identified]

---

### Competitive Landscape Overview

**Market Structure**: [Fragmented / Consolidated / Emerging / Mature]

**Number of Competitors Identified**:
- Direct competitors: [X]
- Indirect competitors: [X]
- Emerging competitors: [X]

**Competitive Intensity**: [Low / Medium / High]

**Evidence**: [Data supporting the assessment]

---

### Direct Competitors

#### Competitor 1: [Company Name]

**Overview:**
- **Founded**: [Year]
- **Headquarters**: [Location]
- **Size**: [Employees, if known]
- **Funding**: $[X] ([Stage], [Year])
- **Website**: [URL]

**Product Offering:**
- **Core Product**: [Description]
- **Key Features**:
  - Feature 1: Description
  - Feature 2: Description
  - Feature 3: Description
- **Unique Capabilities**: [What sets them apart]
- **Technology Stack**: [If known]
- **Integrations**: [Key integrations]

**Market Position:**
- **Target Customers**: [Segments]
- **Customer Count**: [If known]
- **Market Share**: [If known]
- **Revenue**: [If known]

**Pricing:**
| Tier | Price | Features |
|------|-------|----------|
| Tier 1 | $X/mo | Features |
| Tier 2 | $X/mo | Features |
| Tier 3 | $X/mo | Features |

**Positioning & Messaging:**
- **Value Proposition**: "[Their main pitch]"
- **Key Messages**:
  - Message 1
  - Message 2
- **Differentiation Claims**: [How they claim to be different]

**Strengths:**
1. **[Strength 1]**: Description
   - Evidence: [Review quotes, data, or sources]
2. **[Strength 2]**: Description
   - Evidence: [Review quotes, data, or sources]

**Weaknesses:**
1. **[Weakness 1]**: Description
   - Evidence: [Review quotes, data, or sources]
2. **[Weakness 2]**: Description
   - Evidence: [Review quotes, data, or sources]

**Competitive Advantages (MOATs):**
- [Advantage 1]: Description and sustainability
- [Advantage 2]: Description and sustainability

**Recent Developments:**
- [Recent news, product launches, funding]

**Sources:**
- [URL 1]
- [URL 2]

---

#### Competitor 2: [Company Name]
[Same structure as Competitor 1]

---

#### Competitor 3: [Company Name]
[Same structure as Competitor 1]

---

### Indirect Competitors

#### [Alternative Approach/Category]

**Description**: [What this alternative is]

**Examples**:
- [Example 1]: Description
- [Example 2]: Description

**Why Users Choose This**:
- [Reason 1]
- [Reason 2]

**Limitations vs Our Approach**:
- [Limitation 1]
- [Limitation 2]

---

### Emerging Competitors

#### [Startup/New Entrant Name]

**Why Watching**: [What makes them interesting/threatening]

**Overview**: [Brief description]

**Differentiation**: [Their unique approach]

**Momentum Indicators**:
- [Funding, growth, traction]

---

### Feature Comparison Matrix

| Feature/Capability | Us (Planned) | Competitor 1 | Competitor 2 | Competitor 3 |
|-------------------|--------------|--------------|--------------|--------------|
| Feature 1 | ✓ Planned | ✓ | ✗ | ✓ |
| Feature 2 | ✓ Planned | Limited | ✓ | ✗ |
| Feature 3 | ✓ Planned | ✗ | ✗ | ✗ |
| Unique Capability | ✓ Planned | ✗ | ✗ | ✗ |

**Legend**: ✓ Full support, Limited, ✗ Not available

**Analysis**:
- **Table stakes features**: [Features all competitors have]
- **Differentiating features**: [Features only some have]
- **White space opportunities**: [Features no one has well]

---

### Pricing Comparison

| Competitor | Entry Price | Mid Tier | Enterprise | Free Tier |
|------------|-------------|----------|------------|-----------|
| Competitor 1 | $X/mo | $X/mo | Custom | Yes/No |
| Competitor 2 | $X/mo | $X/mo | Custom | Yes/No |
| Competitor 3 | $X/mo | $X/mo | Custom | Yes/No |

**Pricing Strategy Analysis**:
- **Market rate range**: $[X] - $[X] per month
- **Common model**: [SaaS, usage-based, freemium, etc.]
- **Pricing differentiation**: [How pricing varies and why]
- **Our opportunity**: [Pricing strategy recommendation]

---

### Positioning Map

**Positioning Dimensions**: [Axis 1] vs [Axis 2]

```
High [Axis 2]
    │
    │  [Competitor A]
    │           [Competitor C]
    │
    │  [OUR POSITION]
    │      [Competitor B]
    │
Low [Axis 2]
    └────────────────────────────
   Low [Axis 1]      High [Axis 1]
```

**Analysis**:
- **Crowded quadrant**: [Where competition is intense]
- **White space**: [Underserved positioning]
- **Our position**: [Where we can/should position]

---

### Competitive Advantages Analysis

#### Competitor Advantages (Their MOATs)

**Competitor 1:**
- **[Advantage]**: Description, sustainability assessment
- **[Advantage]**: Description, sustainability assessment

**Competitor 2:**
- **[Advantage]**: Description, sustainability assessment

#### Our Differentiation Opportunities

**1. [Opportunity 1]**: Description
- **Why it's valuable**: [Customer value]
- **Why it's defensible**: [Hard to replicate because...]
- **How to build it**: [Approach]

**2. [Opportunity 2]**: Description
- **Why it's valuable**: [Customer value]
- **Why it's defensible**: [Hard to replicate because...]
- **How to build it**: [Approach]

---

### Customer Sentiment Analysis

**What Users Love About Competitors:**
1. **[Common praise 1]**: "[Quote from review]"
   - Competitors doing this well: [List]
2. **[Common praise 2]**: "[Quote from review]"
   - Competitors doing this well: [List]

**What Users Complain About:**
1. **[Common complaint 1]**: "[Quote from review]"
   - Competitors with this issue: [List]
   - **Our opportunity**: [How we can avoid this]
2. **[Common complaint 2]**: "[Quote from review]"
   - Competitors with this issue: [List]
   - **Our opportunity**: [How we can avoid this]

**Most Requested Features:**
1. **[Feature request 1]**: Description
   - Who's asking: [User segment]
   - Who offers it: [Competitors or no one]
2. **[Feature request 2]**: Description
   - Who's asking: [User segment]
   - Who offers it: [Competitors or no one]

**Switching Reasons:**
- **Why users switch TO competitors**: [Reasons]
- **Why users switch FROM competitors**: [Reasons]
- **Our opportunity**: [How to attract switchers]

---

### Market Share & Momentum

| Competitor | Est. Market Share | Customer Count | Recent Growth | Momentum |
|------------|------------------|----------------|---------------|----------|
| Competitor 1 | X% | ~X,000 | +X%/year | ↑ High |
| Competitor 2 | X% | ~X,000 | +X%/year | → Stable |
| Competitor 3 | X% | ~X,000 | +X%/year | ↓ Declining |

**Market Leader**: [Company]
- **Why they lead**: [Reasons]
- **Vulnerability**: [Where they could be challenged]

**Fastest Growing**: [Company]
- **Why growing**: [Reasons]
- **Threat level**: [High/Medium/Low]

---

### Competitive Strategy Recommendations

#### 1. Positioning Strategy
**Recommendation**: [How to position against competitors]
- **Rationale**: [Why this positioning]
- **Differentiation**: [Key differentiators]

#### 2. Feature Strategy
**Must-Have Features** (table stakes):
- [Feature 1]
- [Feature 2]

**Differentiation Features** (unique value):
- [Feature 1]: Why this matters
- [Feature 2]: Why this matters

**Features to Skip** (not worth it):
- [Feature 1]: Why not important

#### 3. Pricing Strategy
**Recommendation**: [Pricing approach]
- **Rationale**: [Why this pricing]
- **Competitive positioning**: [Premium/Mid/Value]

#### 4. Go-to-Market Strategy
**Target Segments**: [Where to focus first]
- **Why**: [Underserved or better fit]

**Competitive Displacement Strategy**: [How to win vs incumbents]

**Messaging Strategy**: [Key messages vs competition]

---

### Competitive Threats & Risks

**Immediate Threats:**
1. **[Threat 1]**: Description and potential impact
   - **Mitigation**: [How to address]

**Emerging Threats:**
1. **[Threat 1]**: Description and timeline
   - **Monitoring**: [What to watch]

**Competitive Dynamics:**
- **Likely responses to our launch**: [What competitors might do]
- **Price war risk**: [High/Medium/Low and why]
- **Partnership conflicts**: [Any issues]

---

### Key Insights

1. **[Insight 1]**: [Important finding about competitive landscape]
2. **[Insight 2]**: [Important finding about differentiation opportunity]
3. **[Insight 3]**: [Important finding about competitive strategy]

---

### Sources

**Competitor Websites:**
- [Competitor 1]: [URL]
- [Competitor 2]: [URL]

**Review Platforms:**
- G2 Crowd: [URLs]
- Capterra: [URLs]
- TrustRadius: [URLs]

**Market Analysis:**
- [Report/Article]: [URL]
- [Report/Article]: [URL]

**News & Media:**
- [Article]: [URL]
- [Article]: [URL]

---

### Research Methodology

**Competitors Analyzed**: [X] direct, [X] indirect

**Sources Consulted**: [Number and types]

**Search Queries Used**:
- "[Query 1]"
- "[Query 2]"

**Limitations**:
- [Limitation 1]
- [Limitation 2]

**Date of Research**: [Date]
```

## Search Strategies

### Competitor Discovery
```
WebSearch queries:
- "[product category] competitors"
- "[problem space] solutions 2025"
- "best [category] tools"
- "alternatives to [known competitor]"
- "[category] G2 leaders"
```

### Company Research
```
WebSearch queries:
- "[competitor] about company"
- "[competitor] funding crunchbase"
- "[competitor] news"
- "[competitor] wikipedia"
```

### Feature Research
```
WebSearch queries:
- "[competitor] features"
- "[competitor] vs [other competitor]"
- "[competitor] review"
- "[competitor] demo video"
```

### Pricing Research
```
WebSearch queries:
- "[competitor] pricing"
- "[competitor] plans cost"
- "[category] pricing comparison"
```

### Customer Sentiment
```
WebSearch queries:
- "[competitor] reviews G2"
- "[competitor] reviews Capterra"
- "[competitor] Reddit"
- "why I switched from [competitor]"
- "[competitor] alternatives"
```

## Best Practices

1. **Cast Wide Net**: Research 5-10 competitors minimum
2. **Mix Direct + Indirect**: Don't ignore alternative approaches
3. **Use Reviews Heavily**: Customer feedback is gold
4. **Look for Patterns**: What do all competitors miss?
5. **Check Multiple Sources**: Triangulate information
6. **Use WebFetch for Depth**: Fetch competitor websites, review pages
7. **Track Momentum**: Recent funding, product launches, growth
8. **Find Switching Reasons**: Why customers leave competitors
9. **Identify White Space**: What no one does well
10. **Be Objective**: Report strengths honestly

## Tools to Use

- **WebSearch**: Primary discovery tool
  - Competitor discovery
  - Feature research
  - Pricing information
  - Review aggregation

- **WebFetch**: Deep analysis tool
  - Competitor homepages for positioning
  - Review pages for detailed feedback
  - Pricing pages for structure
  - About pages for company info

## Success Criteria

A successful competitive analysis provides:
- 5-10 competitors profiled in detail
- Feature comparison matrix
- Pricing comparison
- Strengths and weaknesses for each competitor
- Customer sentiment analysis with quotes
- Clear differentiation opportunities identified
- Defensible competitive advantages mapped
- Strategic recommendations for positioning and features
- 10+ credible sources cited
- Actionable insights for product and GTM strategy
