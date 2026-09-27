# ChoreSync — Vibe-Coding Prompt Package (Lovable / v0.dev / Bolt.new)

> [!NOTE]
> **Status: [ARCHIVED / OUTDATED - Step 2 Initial Prototype Prompts]**  
> This prompt package was an intermediate design artifact used during Step 2 of the AI-native development lifecycle to scaffold the initial frontend mock on vibe-coding platforms (Lovable/v0.dev/Bolt.new).  
> **Current Status**: The frontend is now fully implemented with React 18, Vite, TypeScript, and live Go Chi API integration. This document is preserved for historical traceability.

---

## Architecture Requirements for the Generated Prototype

When generating this prototype, adhere to the following critical conventions:
1. **Framework & Styling**: React (TypeScript), Tailwind CSS, Lucide Icons, and Radix UI / shadcn/ui primitives.
2. **Centralized Service / Mock Layer**:
   - **MANDATORY**: All state mutations and data fetching must be centralized inside a single service module (`src/services/api.ts` or `src/lib/api.ts`).
   - UI components must **never** mutate mock state directly inside `useState`; they should call typed functions like `api.getChores()`, `api.completeChore(id, proof)`, `api.createSwapRequest(...)`.
   - This ensures the client API contract can be cleanly extracted into `contracts/openapi.yaml` in Step 4.
3. **Local Persistence**: Use an in-memory or `localStorage`-backed mock store initialized with realistic seed data.
4. **Mobile-First Responsive Design**: Desktop sidebar layout with bottom bar navigation on mobile viewports.

---

## 📋 Prompt 1: Master App Scaffold & Chore Dashboard (Initial Setup)

*Copy and paste this into Lovable, v0.dev, or Bolt.new as your starting prompt:*

```markdown
Create a modern, responsive web application called "ChoreSync" for managing shared household chores across roommates, families, and couples.

### Visual Aesthetic & Theme
- Clean, warm, and cooperative UI using Tailwind CSS with pastel status badges (green for completed, amber for pending approval, red for overdue, blue for open pool).
- Mobile-first layout: Desktop features a collapsible left sidebar; mobile view uses a clean bottom navigation bar.
- Font: Inter or system sans-serif with crisp hierarchy and micro-interactions.

### Architecture Requirement: Centralized Client Service Layer
- Create `src/services/api.ts` containing all TypeScript interfaces and mock data functions:
  - `Member`: id, name, role ('admin' | 'member' | 'child'), avatar, pointsBalance, streak.
  - `Chore`: id, title, description, category ('cleaning' | 'kitchen' | 'yard' | 'pets' | 'maintenance'), points, assignmentType ('direct' | 'round_robin' | 'open_pool'), assignee (Member object or null), requiresApproval, requiresProof, status ('unassigned' | 'assigned' | 'in_progress' | 'pending_approval' | 'completed' | 'overdue'), dueDate, recurrence ('daily' | 'weekly' | 'none').
- All UI components must fetch and mutate data via `src/services/api.ts` (simulate 150ms network delay with localStorage persistence).

### Core Views & Seed Data
1. **Header & Household Mode Switcher**:
   - Household selector dropdown ("Apartment 4B - Roommates", "The Millers - Family", "Alex & Sam - Couple"). Switching changes view context and active user persona.
   - User profile switcher in header to easily test as "Admin/Parent (Sarah)", "Member/Roommate (Liam)", or "Child (Leo)".
   - Spendable Points pill and Notification bell.

2. **Chore Dashboard**:
   - **Filter Pills**: All, My Chores, Open Pool (unassigned), Needs Approval, Overdue.
   - **Chore Cards**:
     - Chore title, category icon, assignee avatar (or "Claim" button if Open Pool), points badge (+15 pts), and recurrence indicator.
     - Status badges (Overdue with days, Pending Approval, Done).
     - **Quick Actions on Card**:
       - 1-tap "Complete" button (with smooth checkmark animation).
       - If chore has `requiresApproval`, clicking Complete moves it to "Pending Approval".
       - If chore is Open Pool, card shows a prominent "Claim" button.
       - Three-dot menu: "Request Swap", "Send Nudge", "Edit Chore".

Seed with 6 diverse chores representing Roommate, Family, and Couple workflows.
```

---

## 📋 Prompt 2: Chore Creation Modal & Completion / Verification Flow

*Use this prompt after the dashboard is rendered to add modal interactions and proof submissions:*

```markdown
Enhance ChoreSync with the complete Chore Creation and Verification workflows:

### 1. "Add Chore" Modal (Comprehensive & Dynamic)
- Triggered by a floating "+" button or "New Chore" button in the header.
- **Fields**:
  - Title (required) & Description / Checklist notes.
  - Category Selector: Cleaning, Kitchen, Yard, Pet Care, Maintenance, Other.
  - Effort Points slider or stepper (e.g. 5 to 50 points, default 10).
  - **Assignment Strategy Radio Group**:
    1. *Direct Assignment*: Select specific household member from dropdown.
    2. *Automated Round-Robin*: Reorderable list of members who take turns sequentially.
    3. *Open Pool*: Unassigned; anyone can claim.
  - **Recurrence Cadence**: One-off, Daily, Weekly (select days: M, T, W, Th, F, Sa, Su), Monthly.
  - Due Date & Time picker.
  - **Toggles**:
    - "Requires Parent/Admin Approval" (switch).
    - "Requires Photo Proof" (switch).
- Saves new chore through `api.createChore()` in `src/services/api.ts`.

### 2. Completion & Verification Flow
- When a user marks a chore complete:
  - If `requiresProof` is TRUE: Open a "Submit Completion" modal allowing them to enter completion notes and attach a simulated photo proof (mock file upload preview).
  - If `requiresApproval` is TRUE: Mark status as `pending_approval` and emit an activity log event.
  - If instant: Play a subtle celebration confetti animation, mark as `completed`, and immediately credit points to the active member.

### 3. Admin / Parent Approval Queue Drawer
- Dedicated tab or drawer showing chores in `pending_approval`.
- Admin sees: Submitting member, chore title, submitted notes, photo preview, and points to award.
- Actions:
  - **Approve**: Awards points, updates streak, spawns next recurrence instance if recurring.
  - **Needs Re-work / Reject**: Opens input for rejection reason; returns chore back to `assigned` status with feedback.
```

---

## 📋 Prompt 3: Chore Swapping Marketplace & Gamified Rewards Hub

*Use this prompt to add the peer-to-peer swapping and points redemption features:*

```markdown
Add the Chore Swapping Board and the Rewards & Gamification Hub to ChoreSync:

### 1. Chore Swapping & Trading Marketplace
- From any assigned chore card, a member can click "Request Swap":
  - Modal lets member select: "Offer to specific housemate" or "Post to Household Swap Board".
  - Optional note (e.g. "Working late tonight, can anyone take this?").
- **Swap Requests View**:
  - Displays outgoing and incoming swap proposals.
  - Peers see: "Liam wants to swap 'Vacuum Living Room' (15 pts)".
  - Buttons: **Accept Trade** (instantly reassigns chore to acceptor) or **Decline**.
  - Update state cleanly in `src/services/api.ts`.

### 2. Gamification & Rewards Catalog
- Create a dedicated "Rewards" view/page:
  - Header displays active user's spendable points balance with a coin/star icon.
  - **Household Reward Cards**:
    - "Pick Movie Night Film" (30 pts)
    - "Exemption from Dishes for a Day" (50 pts)
    - "$10 Coffee Treat / Allowance" (100 pts)
    - "Late Bedtime Pass" (40 pts - family mode)
  - **Redeem Button**: Deducts points from member balance and creates a `RewardRedemption` in `pending` state.
  - Admin/Parent view has a "Fulfill" button to clear redeemed perks.
  - Leaderboard widget showing top household contributors this week.

### 3. Activity Feed & Gentle Nudges
- Slide-over Activity Feed showing recent household actions:
  - "Sarah approved Leo's 'Clean Bedroom' (+20 pts) 10m ago"
  - "Alex claimed 'Take out trash' from Open Pool 1h ago"
- "Nudge" Action: Clicking "Nudge" on an overdue or upcoming chore triggers a polite in-app toast ("Friendly reminder sent to Liam!").
```

---

## 📋 Prompt 4: Finishing Touches & Mode Customization

*Use this prompt to polish and add the household mode presets:*

```markdown
Polish ChoreSync to ensure full fidelity across all three household personas:

1. **Household Preset Quick-Filter**:
   - In the settings or header mode switcher, ensure the UI adapts appropriately:
     - **Flatmate Mode**: Highlights Round-Robin rotations, the Swap Board, and the Activity Feed. Hides parental approval locks.
     - **Family Mode**: Highlights Parent Approval queue, Kid-friendly colorful task icons, and the Reward Catalog.
     - **Casual / Couples Mode**: Simplifies the dashboard into a streamlined checklist with Open Pool claiming and 1-tap completions.

2. **Verify API Contract Structure in `src/services/api.ts`**:
   - Ensure the following methods are clearly defined, typed, and exported:
     - `getHouseholds()`, `getMembers(householdId)`
     - `getChores(filterOptions)`, `createChore(payload)`, `updateChore(id, payload)`, `deleteChore(id)`
     - `claimChore(choreId, memberId)`
     - `completeChore(choreId, proofPayload)`
     - `approveChore(completionId)`, `rejectChore(completionId, reason)`
     - `requestSwap(payload)`, `acceptSwap(swapId)`, `rejectSwap(swapId)`
     - `getRewards()`, `redeemReward(rewardId, memberId)`, `fulfillReward(redemptionId)`
     - `getActivities()`, `sendNudge(choreId)`

3. **Check for Zero Console Errors**:
   - Verify all mock state transitions function seamlessly with mock data.
```

---

## Next Steps Once the Prototype is Built

1. Export or push your project from Lovable / v0 / Bolt to a GitHub repository.
2. Provide the repository URL (`[repo_url]`).
3. We will ingest it into the `frontend/` directory and proceed to **Step 3: Manual Test Scenario** (`_docs/manual-test.md`).
