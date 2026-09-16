# Manual Verification Scenario (ChoreSync)

This scenario defines the end-to-end validation journey for the **ChoreSync** household chore coordination application. It verifies the interactive frontend prototype with mock services (Step 3) and establishes the test steps that will be automated in Playwright E2E testing (Step 9).

---

## Prerequisites
- Frontend development server running (typically at `http://localhost:5173/`).
- Local storage reset or running with default seed data (`choresync_storage_v1`).

---

## User Journey: End-to-End Verification

### Step 0: Dedicated Authentication & Account Flow
1. **Navigate**: Open `http://localhost:3000/`. When unauthenticated or after sign-out, the dedicated **Login Screen** is displayed.
2. **Inspect Authentication Tabs**:
   - **Sign In**: Email & Password inputs with `Sign In` submit button, rejection on invalid credentials, and a Demo Credentials bar to 1-click prefill demo accounts (Sarah Chen, Liam Vance, David Miller, Alex).
   - **Create Account**: Register new users with Full Name, Email, Password, and either create a new Household or join an existing one using an invite code (`APT4B-SHARE`).
   - **Magic Link**: Passwordless email authentication with a verification token.
   - **Kitchen Fridge Tablet**: Direct one-tap launch to shared touch display.
3. **Sign In**: Enter `sarah@example.com` / `password123` (or click Sarah's demo prefill card) and click **"Sign In"**. Verify redirect to the household dashboard with an active JWT session.

---

### Step 1: Access Application & Verify Seed State
1. **Inspect Header & Household Info**:
   - Verify Active Household is set to **"Apartment 4B"** (`#household-badge-btn`) with the **"Flatmate Mode"** badge.
   - Click `#household-badge-btn`: verify the **Household Details & Invite Modal** opens showing household mode description, member roster, and the active invite code `APT4B-SHARE` with a 1-click "Copy Link" button (preventing arbitrary cross-tenant switching).
   - Verify Active Persona in the User Profile Menu (`#user-profile-button`) is set to **"Sarah Chen"** (Role: `Admin`, Points: `145`, Streak: `6`).
2. **Inspect Board**:
   - Verify the Chores Board renders seed chore cards:
     - *"Recycling & Garbage to Curb"* (Round-robin, assigned to Liam)
     - *"Deep Clean Kitchen Counters & Stovetop"* (Assigned to Sarah)
     - *"Vacuum Living Room & Dust Shelves"* (Overdue alert banner, assigned to Liam)
     - *"Water Plants & Sweep Balcony"* (Open Pool tag, unassigned)
   - Verify household summary stats at the top: total active chores, completion count, and streak metrics.

---

### Step 2: Primary Entity Creation (Recurring Chore with Approval Gate)
1. **Open Modal**: Click the **"+ New Chore"** button in the header navbar or sidebar.
2. **Fill Form Fields**:
   - **Title**: `Wipe Down Kitchen Hood & Filters`
   - **Description**: `Remove grease mesh, soak in warm soapy water, wipe down hood exterior.`
   - **Category**: Select `Kitchen`
   - **Effort Points**: Set to `30`
   - **Assignment Strategy**: Select `Round-Robin (Rotate)`
   - **Recurrence**: Select `Weekly` (Day: `Sa`)
   - **Verification Gate**: Toggle **"Requires Parent / Admin Approval"** to `ON`.
   - **Proof Required**: Toggle **"Photo Proof Required"** to `ON`.
3. **Submit**: Click **"Create Chore"**.
4. **Verify**:
   - Modal dismisses and a success toast notification appears: *"Chore 'Wipe Down Kitchen Hood & Filters' created!"*.
   - New card appears on the board displaying the `30 pts` badge, round-robin rotation indicator, and `ShieldCheck` (Requires Approval) icon.

---

### Step 3: Core Workflow Execution (Claim Open Task & Propose Swap)
1. **Claim Open Task**:
   - On the Chores Board, filter by **"Open Pool"** or locate *"Water Plants & Sweep Balcony"*.
   - Click the blue **"Claim Task"** button.
   - **Verify**: The card transitions immediately to assigned status under Sarah Chen with the updated 1-tap completion button.
2. **Switch Persona for Swap Proposal**:
   - In the user profile menu (`#user-profile-button`) or Kiosk touch bar, switch the active user from **"Sarah Chen"** to **"Liam Vance"** (`m-liam`).
   - Notice the UI adapts to Liam's perspective (Points: `85`, Streak: `3`).
3. **Propose a Peer Swap**:
   - Click **"Swap Market"** in the sidebar.
   - Click **"+ Propose a Swap"**.
   - Select Chore: *"Recycling & Garbage to Curb"*.
   - Select Target Peer: *"Maya Patel"*.
   - Enter Reason: `"Final exams on Tuesday evening, can take your weekend chore in return."`
   - Click **"Send Swap Proposal"**.
   - **Verify**: A pending swap request card appears under **"Active Swap Offers"** targeted to Maya Patel with status `Pending Maya Patel`.

---

### Step 4: State Transition & Completion (Instant vs. Approval Gate)
1. **Instant Self-Completion (No Approval Gate)**:
   - As **Liam Vance**, navigate back to **"Chores Board"**.
   - Locate the overdue chore: *"Vacuum Living Room & Dust Shelves"*.
   - Click **"Mark Completed"**.
   - **Verify**:
     - Confetti animation triggers on screen.
     - Card updates to **"Done & Verified"** with emerald styling.
     - Liam's points balance increases by `+15` (from `85` to `100`), and streak increments.
2. **Completion with Verification Gate**:
   - Locate *"Wipe Down Kitchen Hood & Filters"* or an assigned gated chore.
   - Click **"Submit for Approval"**.
   - In the `CompletionModal`:
     - Select a sample proof photo (e.g. *"Clean Kitchen"*).
     - Add proof notes: `"Degreaser applied, soaked filters for 30 mins, reinstalled."`
     - Click **"Submit for Verification"**.
   - **Verify**:
     - Card updates state to **"Awaiting Verification"** with amber spinner / hourglass icon.
     - The **"Approval Queue"** badge in the sidebar displays an alert badge (`1`).

---

### Step 5: Verification, Persistence & Activity Hub
1. **Admin Approval Gate**:
   - Switch active persona in the navbar back to Admin **"Sarah Chen"**.
   - Click **"Approval Queue"** in the sidebar.
   - Locate Liam's submission for *"Wipe Down Kitchen Hood & Filters"*.
   - Inspect the attached photo thumbnail and Liam's proof notes.
   - Click **"Approve Completion"**.
   - **Verify**: The item disappears from the queue with an approval toast notification, and Liam is credited with `+30` points.
2. **Audit Trail Inspection**:
   - Click **"Activity Feed"** in the sidebar (or bell icon).
   - Verify chronological logs reflect the full journey:
     - Chore creation event.
     - Open pool chore claimed by Sarah.
     - Swap requested by Liam to Maya.
     - Chore completion approved by Sarah for Liam.
3. **Persistence Verification**:
   - Hard refresh the browser (`F5` or `Ctrl+R`).
   - **Verify**: All state updates (Liam's updated points balance, completed chores, and activity logs) remain intact from `localStorage`.

---

### Step 6: Edge Case & Validation Guardrails
1. **Form Validation Guardrail**:
   - Click **"+ New Chore"**.
   - Leave the **Title** input completely blank.
   - Click **"Create Chore"**.
   - **Verify**: The form blocks submission, retaining input focus without corrupting the chore list or sending empty payloads.
2. **Role Permission Guardrail (Approval Protection)**:
   - Switch persona to a non-admin member (e.g. **"Liam Vance"** or **"Leo (10y)"** in Family Mode).
   - Navigate to **"Approval Queue"**.
   - **Verify**: Non-admin warning banner displays:
     > *"Admin Privileges Required: Approvals can only be resolved by household Admins/Parents."*
   - Approval/Rejection buttons are disabled or hidden for non-admin personas.

---

### Step 8: Shared Tablet / Kiosk Mode vs Personal Device Mode
1. **Kitchen Tablet / Kiosk Mode (Shared Wall Display)**:
   - Verify header shows `[📱 Kitchen Tablet] [👤 Personal Device]` segmented toggle with **Kitchen Tablet** selected by default.
   - Inspect the **Kitchen Tablet Kiosk Touch Bar** above the chores board:
     - Prominent touch cards for all household members with avatars, points, and pending task counts.
     - Tap any member card (e.g. **"Liam Vance"**) to switch user persona instantly in 1 tap.
     - Click **"Admin PIN"** to open the 4-digit PIN verification modal (`PinVerificationModal`). Entering `1234` unlocks Admin role.
2. **Personal Device Mode (Personal Phone / Laptop)**:
   - Click **"Personal Device"** in the header toggle or touch bar.
   - Verify the **Personal Device Welcome Banner** appears:
     - Warm personalized greeting: *"Hi, Liam Vance! 👋"*
     - Summary of personal tasks due today.
     - Board automatically defaults to **"My Tasks"**, displaying only chores assigned to that member.
     - Chores assigned to other household members (e.g. Sarah's cleaning tasks) are filtered out, preventing UX confusion.
   - Click **"Whole House"** in the welcome banner to view all household chores when desired.
   - Click **"Tablet Mode"** to return to the shared household kiosk whiteboard.

---

### Step 9: Admin PIN Modification and Shared Kiosk Verification
1. **Access Settings as Admin**:
   - Log in as **Sarah Chen** (Admin).
   - Navigate to **"Household Settings"** (`#nav-btn-settings`).
2. **Change PIN**:
   - Click **"Change PIN"** (`#edit-admin-pin-btn`).
   - Enter new PIN `8844` and save.
   - Verify modal closes and updated PIN is reflected.
3. **Non-Admin Protection**:
   - Switch user to **Liam Vance** (Member).
   - Verify Change PIN button is hidden and shows `(Admin only)`.
4. **Kiosk Unlock Verification**:
   - Switch to **Kitchen Tablet Kiosk** mode.
   - Tap **Liam Vance** to activate member persona.
   - Open Admin PIN unlock modal; enter old PIN `1234` and verify rejection with error indicator.
   - Enter new PIN `8844` and verify successful unlock and persona switch back to **Sarah Chen** (Admin).

---

### Step 10: External Email & Notification Integration (Password Reset & Mailbox Inspector)
1. **Request Password Reset**:
   - Sign out from active session to land on the dedicated **Login View**.
   - Click **"Forgot password?"** next to the Password input field (`#forgot-password-link`).
   - Enter registered email `sarah@example.com` and submit (`#forgot-password-submit-btn`).
   - Verify green confirmation banner appears: *"Password reset link sent! Check your inbox."*
2. **Inspect Transactional Mailbox**:
   - Click **"📬 Transactional Mailbox"** button in top bar (`#open-dev-mailbox-btn`).
   - Inspect modal showing recent transactional emails captured in dev buffer.
   - Click the latest email item (`Reset your ChoreSync password 🔐`).
   - Verify recipient is `sarah@example.com`, sender is `ChoreSync <notifications@choresync.app>`, and HTML email body includes the branded template, expiration warning, and 1-click reset link with token `RST-...`.
   - Close modal (`#close-dev-mailbox-btn`).
3. **Complete Password Reset & Log In**:
   - Click **"Enter Reset Token"** (`#proceed-to-reset-btn`).
   - Enter reset token and choose new password `newpassword456` (confirming matching password).
   - Submit form (`#reset-password-submit-btn`).
   - Verify success banner and automatic login into Sarah Chen's household account.
   - Sign out and verify logging in with new password `newpassword456` works seamlessly.

---

### Step 11: Chore Reminder Nudge Email Notification
1. **Send Nudge**:
   - Log in as **Sarah Chen** (Admin) in **Apartment 4B**.
   - Navigate to Dashboard and locate Liam's chore *"Vacuum Living Room & Dust Shelves"*.
   - Open chore kebab menu (`#chore-menu-btn-c-vacuum-living`) and click **"Send Friendly Nudge"** (`#nudge-chore-btn-c-vacuum-living`).
   - Verify toast alert appears: *"Friendly reminder sent!"*.
2. **Verify Email Dispatch**:
   - Inspect dev emails buffer via `/api/v1/dev/emails` or open the Transactional Mailbox.
   - Verify a new `chore_reminder` email is captured with recipient `liam@example.com`.
   - Verify email content includes chore title, due date, sender name (`Sarah Chen`), and 1-click action link back to ChoreSync.
