# ChoreSync: Working Backwards PRFAQ

**Document Status**: Approved  
**Lifecycle Milestone**: Pre-Step 1: Ideation & Working Backwards  
**Product**: ChoreSync – Equitable Household Chore Coordination System  

---

## 1. Press Release

### FOR IMMEDIATE RELEASE

**ChoreSync Unveils AI-Native Household Chore Coordination Platform to Eliminate Domestic Friction Across Flatmates, Families, and Couples**

**SAN FRANCISCO, CA — September 2026** — Today, the ChoreSync team announced the general availability of **ChoreSync**, an intelligent, equitable household management system engineered to transform domestic chores from a daily source of tension into a transparent, collaborative routine. Built from the ground up to support three distinct living arrangements—student/young professional flatshares, multi-generational families with children, and couples co-living—ChoreSync bridges the gap between individual accountability and harmonious communal living.

Household chore imbalances remain one of the leading causes of domestic friction worldwide. Traditional chore apps force a rigid, one-size-fits-all model: either an over-engineered project management tool unsuitable for casual roommates, or a child-oriented sticker chart irrelevant for adults. ChoreSync solves this dilemma with adaptive living arrangement profiles:
- **Flatmates & Roommates**: Features automated round-robin task rotation, a peer-to-peer chore swapping marketplace, and transparent contribution audit logs that eliminate uncomfortable confrontations over unwashed dishes or full trash bins.
- **Families with Children**: Introduces parent/admin approval gates with mandatory photo proof uploads, customizable point values, and redeemable real-world reward catalogs that turn responsibility into an engaging family experience.
- **Couples & Lightweight Co-living**: Provides a shared, voluntary claiming backlog with one-tap instant completion and automated gentle reminder nudges.

"Nobody wants to be the household nag, and nobody enjoys feeling like their contributions go unnoticed," said the ChoreSync Lead Product Architect. "With ChoreSync, domestic fairness is automated. Tasks rotate predictably, swaps happen with mutual consent, and parents can verify completed chores without micromanagement. It provides total visibility with zero awkwardness."

ChoreSync is designed for everyday accessibility. Its dual-mode interface operates seamlessly on shared kitchen tablets with 1-tap member switching and PIN-protected admin controls, as well as on private mobile devices.

### Availability & Deployment
ChoreSync is fully open-source and deployable in seconds via lightweight Docker Compose clusters or local Kubernetes via Kind, running with zero compiler dependencies on the host OS. For more information, visit https://github.com/pithwamanish/shared-household-chores.

---

## 2. Customer FAQ (External)

### Q1: What makes ChoreSync different from standard to-do or chore list apps?
ChoreSync does not treat chores as static checkboxes. It recognizes that households operate under different social dynamics. Instead of forcing everyone into the same workflow, ChoreSync allows households to select their arrangement (Flatmates, Families, or Couples) and configures rules accordingly—such as peer-to-peer task trading for flatmates, photo proof and reward redemption for families, and voluntary claim pools for couples.

### Q2: How does the chore rotation engine work for roommates?
When a recurring chore (like cleaning the bathroom or vacuuming the common area) is marked complete, ChoreSync's idempotent round-robin engine automatically advances the assignment to the next active household member in order. The entire household can see who is up next, eliminating confusion and debates over "whose turn it is."

### Q3: What happens if a roommate cannot complete a chore due to travel or work?
Roommates can initiate a chore swap through the built-in Peer Swap Marketplace. The user selects a target chore and sends a swap proposal to a roommate. The chore remains the original owner's responsibility until the recipient formally accepts the trade, preventing dropped tasks.

### Q4: How does ChoreSync handle chore verification for kids?
For chores flagged with requires_approval, completing the task shifts its status to pending_approval and prompts the child to upload photo proof. The chore only credits gamification points once a parent/admin verifies the submission and approves it.

### Q5: Can ChoreSync be installed on a shared iPad or tablet on our kitchen counter?
Yes. ChoreSync includes a dedicated **Kitchen Tablet / Kiosk Mode**. Household members can switch profiles with a single tap on the header bar to view their specific tasks or log completions. Admin settings and approval gates require a secure 4-digit PIN, ensuring children cannot approve their own chores or modify rewards.

### Q6: What if someone forgets their chore?
ChoreSync integrates an asynchronous reminder engine. Household members can send gentle nudge reminders, which queue through an event bus and dispatch notification emails to assignees without personal friction.

### Q7: Is my household data private?
Absolutely. ChoreSync enforces strict multi-tenant isolation. All database records, member rosters, proof photos, and activity logs are scoped strictly to your unique household ID, protected by stateless HMAC-SHA256 authentication and bcrypt password hashing.

---

## 3. Internal Engineering FAQ

### Q1: Why did we adopt the AI-Native Spec-Driven Development Methodology for ChoreSync?
The AI-native methodology guarantees rigorous separation between requirements, contracts, and code. By freezing the OpenAPI 3.1 specification (contracts/openapi.yaml) and establishing constitution.md before writing implementation code, we prevented architectural drift, eliminated hallucinated endpoints, and enabled seamless parallel multi-agent development.

### Q2: Why was Golang 1.22+ with Chi selected over Node.js or Python?
Golang provides ultra-low memory consumption (~15–25MB RAM vs ~120MB in Node/Python), sub-50ms cold starts, and single-binary packaging. This enables ChoreSync to run multiple services, databases, and OpenTelemetry sidecars comfortably within free-tier cloud quotas (Render, Neon, Fly.io).

### Q3: How does ChoreSync achieve 100% offline cloud parity during local development?
We integrated **Floci (`floci/floci:latest`)** as a local cloud emulator on port 4566. Floci provides local AWS S3 object storage for chore photo proofs and AWS SQS queueing for asynchronous reminder nudges. The Go backend uses standard AWS SDK Go v2, allowing identical code and manifests to run against Floci locally and real AWS in production without code changes.

### Q4: How does the system guard against contract drift?
The repository incorporates automated AST route auditing (`verify-spec-drift`), linting (`make lint`), and containerized mutation testing (`run-mutation-test`). Mutation testing intentionally injects faults into core logic to verify that test suites achieve >= 80% mutant kill rates, ensuring tests are not tautological.

### Q5: What operational safety safeguards protect the system during incidents?
ChoreSync enforces an outside-the-model autonomy policy (`on-call-engineer/autonomy-policy.json`). Automated on-call responders collect bounded, read-only telemetry packets (Prometheus, Loki, Tempo) without database write access. Proposed remediation actions must be evaluated against the policy engine before execution: Level 1 actions (rollback, restart) execute automatically, while Level 2 actions (patches) require human authorization.
