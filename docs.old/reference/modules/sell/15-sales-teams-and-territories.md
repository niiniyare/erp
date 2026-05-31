[<-- Back to Index](README.md)

## 15. Sales Teams & Territories

### Territory Definition

```markdown
TERRITORY MANAGEMENT

Territory Structure:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Geographic Territories:

Kenya
├─ Nairobi Region
│  ├─ Nairobi CBD
│  ├─ Westlands
│  ├─ Industrial Area
│  └─ Satellite Towns (Ruiru, Thika, Kiambu)
│
├─ Coast Region
│  ├─ Mombasa
│  ├─ Malindi
│  └─ Kilifi
│
├─ Western Region
│  ├─ Kisumu
│  ├─ Eldoret
│  └─ Kakamega
│
├─ Central Region
│  ├─ Nakuru
│  ├─ Nyeri
│  └─ Nanyuki
│
└─ Other Counties

Regional East Africa:
├─ Uganda (Kampala, Entebbe)
├─ Tanzania (Dar es Salaam, Arusha)
├─ Rwanda (Kigali)
└─ South Sudan (Juba)

Territory Master Record:
┌────────────────────────────────────────────┐
│ Territory Name: Nairobi Corporate          │
│ Parent Territory: Nairobi Region           │
│ Territory Manager: Sarah Johnson           │
│                                            │
│ Coverage:                                  │
│ - Nairobi CBD                              │
│ - Westlands                                │
│ - Industrial Area                          │
│                                            │
│ Customer Segments:                         │
│ ☑ Corporate                                │
│ ☑ Large Enterprise                         │
│ □ SME                                      │
│ □ Retail                                   │
│                                            │
│ Annual Target: 200M KES                    │
│ YTD Achievement: 180M (90%)                │
│                                            │
│ Team Size: 4 sales persons                │
│ Active Customers: 85                       │
│ Pipeline Value: 120M KES                   │
└────────────────────────────────────────────┘

Territory Assignment Rules:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Automatic Assignment:
  When new customer created:
    → System checks customer address
    → Matches to territory
    → Assigns default sales person
    → Notifies sales team

Manual Override:
  Sales manager can:
    - Reassign customer to different territory
    - Transfer customers between sales persons
    - Split large accounts (team selling)

Territory Overlap:
  Some customers span multiple territories
  Solution: Primary & Secondary assignment
  
  Example:
    Customer: Nationwide Retail Chain
    Primary: Nairobi (HQ location)
    Secondary: All regions (store locations)
    Revenue split: 60% primary, 40% split

Territory Performance:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Comparative Analysis:
┌──────────────┬──────────┬──────────┬───────────┐
│ Territory    │ Target   │ Actual   │ Achievement│
├──────────────┼──────────┼──────────┼───────────┤
│ Nairobi Corp │ 200M     │ 180M     │   90%     │
│ Nairobi SME  │ 120M     │ 135M     │  113% ✓   │
│ Mombasa      │ 150M     │ 142M     │   95%     │
│ Kisumu       │  80M     │  92M     │  115% ✓   │
│ Nakuru       │  50M     │  48M     │   96%     │
└──────────────┴──────────┴──────────┴───────────┘

Market Penetration:
┌──────────────┬─────────────┬────────┬─────────┐
│ Territory    │ Total Market│ Our    │ Share   │
├──────────────┼─────────────┼────────┼─────────┤
│ Nairobi Corp │ 2,000M      │ 180M   │   9%    │
│ Mombasa      │   800M      │ 142M   │  18% ✓  │
│ Kisumu       │   400M      │  92M   │  23% ✓  │
└──────────────┴─────────────┴────────┴─────────┘

Growth Opportunity:
  Nairobi has low penetration but high potential
  Focus: Increase market share from 9% to 12%
```

### Sales Team Structure

```markdown
SALES TEAM ORGANIZATION

Team Hierarchy:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Chief Sales Officer (CSO)
│
├─ National Sales Director
│  │
│  ├─ Regional Sales Manager - Nairobi
│  │  ├─ Team Lead - Corporate
│  │  │  ├─ Senior Sales Executive
│  │  │  ├─ Sales Executive
│  │  │  └─ Junior Sales Executive
│  │  │
│  │  └─ Team Lead - SME
│  │     ├─ Sales Executive
│  │     └─ Sales Executive
│  │
│  ├─ Regional Sales Manager - Coast
│  │  ├─ Sales Executive (Mombasa)
│  │  └─ Sales Executive (Malindi)
│  │
│  └─ Regional Sales Manager - Western
│     ├─ Sales Executive (Kisumu)
│     └─ Sales Executive (Eldoret)
│
├─ Key Account Manager (Large Accounts)
│  ├─ Strategic Account Exec (Top 10 customers)
│  └─ Strategic Account Exec (Next 20 customers)
│
└─ Inside Sales Manager (Telesales/Online)
   ├─ Inside Sales Rep
   ├─ Inside Sales Rep
   └─ Inside Sales Rep

Team Specialization Models:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Model 1: Geographic (Current)
  Each sales person owns a territory
  Pros: Clear ownership, local expertise
  Cons: May lack product expertise

Model 2: Industry Vertical
  Sales teams by industry:
    - Manufacturing Team
    - Retail/Distribution Team
    - Construction Team
    - Healthcare Team
  Pros: Deep industry knowledge
  Cons: Territory conflicts

Model 3: Product Line
  Sales teams by product:
    - Machinery Team
    - Equipment Team
    - Services Team
  Pros: Product expertise
  Cons: Customer confusion (multiple reps)

Model 4: Customer Size
  Segmented by customer value:
    - Enterprise Team (>10M annual)
    - Mid-Market Team (1-10M)
    - SMB Team (<1M)
  Pros: Appropriate resource allocation
  Cons: Customers may graduate between teams

Hybrid Model (Recommended):
  - Geographic territories (primary)
  - Industry specialists (overlay)
  - Key account managers (strategic)
  - Inside sales (small customers)

Sales Team Master:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
┌────────────────────────────────────────────┐
│ Team Name: Nairobi Corporate Team          │
│ Team Manager: James Ndungu                 │
│                                            │
│ Members:                                   │
│ 1. Sarah Johnson (Team Lead)               │
│ 2. Mike Chen (Senior SE)                   │
│ 3. Jane Mwangi (SE)                        │
│ 4. Tom Omondi (Junior SE)                  │
│                                            │
│ Territory: Nairobi Corporate               │
│ Customer Segment: Enterprise (Corporate)   │
│                                            │
│ Team Target: 200M KES (2025)               │
│ Individual Targets:                        │
│   Sarah: 60M                               │
│   Mike: 55M                                │
│   Jane: 50M                                │
│   Tom: 35M                                 │
│                                            │
│ Commission Structure: Team-based           │
│ Split: 60% individual, 40% team pool       │
└────────────────────────────────────────────┘

Team Collaboration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deal Registration:
  Sales person registers opportunity
  Prevents multiple reps approaching same customer
  First to register gets the deal

Team Selling:
  Large complex deals require team:
    - Account Executive (relationship)
    - Technical Sales Engineer (solution)
    - Sales Manager (negotiations)
  
  Revenue Split:
    Account Exec: 50%
    Technical SE: 30%
    Manager: 20%

Lead Distribution:
  Inbound leads distributed via:
    - Round Robin (equal distribution)
    - Territory Match (geographic)
    - Skill Match (product/industry)
    - Load Balancing (current pipeline)

Handoff Process:
  Lead → SDR qualifies → AE closes
  Inbound → Inside Sales → Field Sales (large deal)
  Trial → Success Team → Renewal Team
```

### Sales Meetings & Cadence

```markdown
SALES RHYTHM & MEETINGS

Daily Huddle (15 minutes):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Time: 8:30 AM
Attendees: Entire sales team
Format: Stand-up

Agenda:
  □ Yesterday's wins (deals closed)
  □ Today's priorities
  □ Blockers/support needed
  □ Quick updates

Example:
  Sarah: "Closed ABC Mfg deal - 5M. Today meeting 
         XYZ Corp for final negotiation. Need pricing 
         approval for 12% discount."
  
  Manager: "Great work! I'll fast-track the approval. 
           Mike, can you join the XYZ meeting for 
           technical support?"

Weekly Sales Meeting (1 hour):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Time: Monday 9:00 AM
Attendees: Regional team + manager

Agenda:
  1. Week in Review (20 min)
     - Revenue vs target
     - Key wins and losses
     - Pipeline review
  
  2. Week Ahead (15 min)
     - Major opportunities closing
     - Customer meetings
     - Priorities
  
  3. Coaching Corner (15 min)
     - Deal strategy discussion
     - Role play scenarios
     - Best practices sharing
  
  4. Announcements (10 min)
     - New products/promotions
     - Policy updates
     - Recognition

Pipeline Review (Bi-weekly, 90 min):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deep dive on each rep's pipeline:

Review per Sales Person:
  Opportunities > 1M KES
  For each deal:
    - Customer background
    - Opportunity size
    - Stage and probability
    - Competition
    - Next steps
    - Close date
    - Support needed

Manager Actions:
  - Challenge assumptions
  - Provide coaching
  - Commit forecasts
  - Assign resources

Example Review:
  Opportunity: DEF Corporation - 8M
  Stage: Proposal
  Probability: 60%
  Competition: Competitor X
  
  Manager: "What's their main objection?"
  Rep: "Price. We're 10% higher."
  Manager: "Have you quantified TCO benefits?"
  Rep: "Not yet."
  Manager: "Work with product team on ROI 
           analysis. Present that next week."

Monthly Business Review (2 hours):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Attendees: Sales leadership

Agenda:
  1. Monthly Performance
     - Actual vs Target
     - YTD performance
     - Trend analysis
  
  2. Pipeline Health
     - Weighted pipeline
     - Coverage ratio
     - Stage distribution
  
  3. Team Performance
     - Individual rankings
     - Activity metrics
     - Win/loss analysis
  
  4. Customer Analysis
     - Top customers
     - Churn risks
     - Expansion opportunities
  
  5. Next Month Plan
     - Target allocation
     - Focus areas
     - Initiatives

Quarterly Planning (Half-day):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Strategic planning session:
  - Review previous quarter
  - Set next quarter goals
  - Territory realignment (if needed)
  - Training needs
  - Process improvements
  - Market opportunities
```

---