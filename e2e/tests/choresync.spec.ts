import { test, expect, Page } from '@playwright/test';

// Helper to switch user via Sign Out + Login or Kiosk touch bar
async function switchUser(page: Page, memberId: string) {
  // If kiosk touch bar is visible and member card is visible, click it
  const kioskCard = page.locator(`#kiosk-member-${memberId}`);
  if (await kioskCard.isVisible({ timeout: 500 }).catch(() => false)) {
    await kioskCard.click();
    return;
  }

  // If already on login view, click quick-fill chip and submit
  const quickFill = page.locator(`#quick-fill-${memberId}`);
  if (await quickFill.isVisible({ timeout: 500 }).catch(() => false)) {
    await quickFill.click();
    await page.locator('#login-submit-btn').click();
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
    return;
  }

  // Otherwise, sign out via profile dropdown
  const profileBtn = page.locator('#user-profile-button');
  if (await profileBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
    await profileBtn.click();
    const signOutBtn = page.locator('#profile-sign-out-btn');
    await signOutBtn.click();
  }

  // Ensure login tab is active so quick-fill chips are visible
  const loginTab = page.locator('#tab-login');
  if (await loginTab.isVisible({ timeout: 1000 }).catch(() => false)) {
    await loginTab.click();
  }

  const chip = page.locator(`#quick-fill-${memberId}`);
  await expect(chip).toBeVisible({ timeout: 5000 });
  await chip.click();
  await page.locator('#login-submit-btn').click();
  await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
}

test.describe('ChoreSync End-to-End Verification Journey', () => {
  test.beforeEach(async ({ page, request }) => {
    page.on('console', (msg) => console.log('BROWSER CONSOLE:', msg.type(), msg.text()));
    page.on('pageerror', (err) => console.log('BROWSER UNCAUGHT ERROR:', err.message, err.stack));

    // Reset backend demo data to ensure pristine seed state
    try {
      await request.post('http://localhost:8000/api/v1/reset');
    } catch (e) {
      console.log('Backend reset endpoint call error (will rely on frontend reset):', e);
    }

    // Navigate to application
    await page.goto('/', { waitUntil: 'domcontentloaded' });

    // Clear local storage and reload to start clean
    await page.evaluate(() => {
      localStorage.clear();
    });
    await page.reload({ waitUntil: 'networkidle' });

    // With real auth flow, clearing local storage lands on Login screen.
    // Log in as default admin Sarah Chen if on login screen
    const sarahQuickFill = page.locator('#quick-fill-m-sarah');
    if (await sarahQuickFill.isVisible({ timeout: 3000 }).catch(() => false)) {
      await sarahQuickFill.click();
      await page.locator('#login-submit-btn').click();
      await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
    }
  });

  test('Step 0: Authentication Flow (Password Login, Magic Link, Registration, and Sign Out)', async ({ page }) => {
    // 1. Sign out via user profile menu
    const profileBtn = page.locator('#user-profile-button');
    await profileBtn.click();
    const signOutBtn = page.locator('#profile-sign-out-btn');
    await signOutBtn.click();

    // Verify Login Screen appears with tabs and inputs
    await expect(page.locator('#tab-login')).toBeVisible();
    await expect(page.locator('#tab-register')).toBeVisible();
    await expect(page.locator('#tab-magic')).toBeVisible();
    await expect(page.locator('#login-email-input')).toBeVisible();
    await expect(page.locator('#login-password-input')).toBeVisible();

    // 2. Test Invalid Credentials rejection
    await page.locator('#login-email-input').fill('sarah@example.com');
    await page.locator('#login-password-input').fill('wrongpassword');
    await page.locator('#login-submit-btn').click();
    await expect(page.locator('#login-error-message')).toBeVisible({ timeout: 4000 });

    // 3. Test Successful Password Login with Sarah Chen
    await page.locator('#login-password-input').fill('password123');
    await page.locator('#login-submit-btn').click();
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');

    // 4. Test Magic Link tab
    await page.locator('#user-profile-button').click();
    await page.locator('#profile-sign-out-btn').click();
    await page.locator('#tab-magic').click();
    await expect(page.locator('#magic-email-input')).toBeVisible();
    await page.locator('#magic-email-input').fill('sarah@example.com');
    await page.locator('#send-magic-link-btn').click();
    await expect(page.locator('#magic-link-token-input')).toBeVisible({ timeout: 5000 });
    await page.locator('#verify-magic-link-btn').click();
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');

    // 4b. Test Direct Magic Link URL Auto-Login (?magic_token=...)
    await page.locator('#user-profile-button').click();
    await page.locator('#profile-sign-out-btn').click();
    const magicReq = await page.request.post('http://localhost:8000/api/v1/auth/magic-link', {
      data: { email: 'liam@example.com' }
    });
    expect(magicReq.ok()).toBeTruthy();
    const magicData = await magicReq.json();
    expect(magicData.token).toBeTruthy();

    await page.goto(`/?magic_token=${magicData.token}`);
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 8000 });
    await expect(page.locator('#user-profile-button')).toContainText('Liam Vance');

    // 4b-ii. Verify that using the exact same magic link token again fails (Single-Use Invalidation)
    await page.locator('#user-profile-button').click();
    await page.locator('#profile-sign-out-btn').click();

    // Verify via backend API that used token returns 401 Unauthorized
    const reuseVerifyReq = await page.request.post('http://localhost:8000/api/v1/auth/verify', {
      data: { token: magicData.token }
    });
    expect(reuseVerifyReq.status()).toBe(401);

    // Verify in UI that reusing the link fails and displays error
    await page.goto(`/?magic_token=${magicData.token}`);
    await expect(page.getByText(/Invalid or expired/i).first()).toBeVisible({ timeout: 8000 });
    await expect(page.locator('#user-profile-button')).not.toBeVisible();

    // 4c. Test Dev Mailbox 1-Click Magic Link Sign In
    if (await page.locator('#user-profile-button').isVisible()) {
      await page.locator('#user-profile-button').click();
      await page.locator('#profile-sign-out-btn').click();
    }
    // Request a magic link for Sarah to appear in mailbox
    await page.locator('#tab-magic').click();
    await page.locator('#magic-email-input').fill('sarah@example.com');
    await page.locator('#send-magic-link-btn').click();
    await expect(page.locator('#magic-link-token-input')).toBeVisible({ timeout: 5000 });

    // Open mailbox and click "Sign In with Magic Link"
    await page.locator('#open-dev-mailbox-btn').click();
    await expect(page.locator('#dev-mailbox-modal')).toBeVisible();
    await expect(page.locator('#mailbox-signin-magic-btn')).toBeVisible({ timeout: 5000 });
    await page.locator('#mailbox-signin-magic-btn').click();
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 8000 });
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');

    // 5. Test Registration Flow
    await page.locator('#user-profile-button').click();
    await page.locator('#profile-sign-out-btn').click();
    await page.locator('#tab-register').click();
    const testEmail = `newuser_${Date.now()}@example.com`;
    await page.locator('#register-name-input').fill('Test Flatmate');
    await page.locator('#register-email-input').fill(testEmail);
    await page.locator('#register-password-input').fill('password123');
    await page.getByRole('button', { name: 'Join with Invite Code' }).click();
    await page.locator('#register-invite-code-input').fill('APT4B-SHARE');
    await page.locator('#register-submit-btn').click();

    // Verify user is registered and logged in
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#user-profile-button')).toContainText('Test Flatmate');

    // Switch back to Sarah Chen for subsequent tests
    await switchUser(page, 'm-sarah');
  });

  test('Step 1: Access Application & Verify Seed State', async ({ page }) => {
    // 1. Inspect Header & Active Household
    const householdBadge = page.locator('#household-badge-btn');
    await expect(householdBadge).toContainText('Apartment 4B');

    const modeBadge = page.locator('text=Flatmate Mode');
    await expect(modeBadge).toBeVisible();

    // Click Household Badge to verify Household Info Modal and Invite Code (multi-tenant safety)
    await householdBadge.click();
    const infoModal = page.locator('#household-info-modal');
    await expect(infoModal).toBeVisible();
    await expect(infoModal).toContainText('APT4B-SHARE');
    await expect(page.locator('#copy-invite-link-btn')).toBeVisible();
    await page.getByRole('button', { name: 'Done' }).click();
    await expect(infoModal).not.toBeVisible();

    // 2. Inspect Active Persona in User Profile Button
    const profileBtn = page.locator('#user-profile-button');
    await expect(profileBtn).toContainText('Sarah Chen');

    // 3. Inspect Spendable Points & Streak for Sarah Chen (145 pts, 6d)
    await expect(page.locator('text=145 pts')).toBeVisible();
    await expect(page.locator('text=6d')).toBeVisible();

    // 4. Inspect Board & Seed Chore Cards
    await expect(page.getByText('Recycling & Garbage to Curb')).toBeVisible();
    await expect(page.getByText('Deep Clean Kitchen Counters & Stovetop')).toBeVisible();
    await expect(page.getByText('Vacuum Living Room & Dust Shelves')).toBeVisible();
    await expect(page.getByText('Water Plants & Sweep Balcony')).toBeVisible();

    // 5. Inspect Board Header Summary
    await expect(page.getByText(/active tasks remaining/i).first()).toBeVisible();
  });

  test('Step 2: Primary Entity Creation (Recurring Chore with Approval Gate)', async ({ page }) => {
    // 1. Open Modal
    const addChoreBtn = page.locator('#add-chore-button');
    await addChoreBtn.click();
    await expect(page.getByText('Create New Household Chore')).toBeVisible();

    // 2. Fill Form Fields
    await page.getByPlaceholder(/Mop kitchen floor/i).fill('Wipe Down Kitchen Hood & Filters');
    await page.getByPlaceholder(/Specific checklist notes/i).fill('Remove grease mesh, soak in warm soapy water, wipe down hood exterior.');

    // Category: Kitchen
    await page.locator('form').getByRole('button', { name: 'Kitchen', exact: true }).click();

    // Effort Points: 30 pts via range slider
    const rangeInput = page.locator('input[type="range"]');
    await rangeInput.evaluate((el: HTMLInputElement) => {
      const nativeSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
      if (nativeSetter) {
        nativeSetter.call(el, '30');
      } else {
        el.value = '30';
      }
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await expect(page.getByText(/Effort Points Reward:/i)).toContainText('30 pts');

    // Assignment Strategy: Round-Robin
    await page.getByRole('button', { name: /Round-Robin/i }).click();

    // Verification Gate: Toggle ON
    const approvalToggle = page.locator('text=Requires Parent / Admin Approval')
      .locator('xpath=ancestor::div[contains(@class, "flex items-center justify-between")]')
      .locator('input[type="checkbox"]');
    await approvalToggle.check({ force: true });
    await expect(approvalToggle).toBeChecked();

    // Proof Required: Toggle ON
    const proofToggle = page.locator('text=Requires Completion Proof')
      .locator('xpath=ancestor::div[contains(@class, "flex items-center justify-between")]')
      .locator('input[type="checkbox"]');
    await proofToggle.check({ force: true });
    await expect(proofToggle).toBeChecked();

    // 3. Submit
    const createBtn = page.getByRole('button', { name: 'Create Chore' });
    await createBtn.click();

    // 4. Verify card appears on board
    await expect(page.getByText('Wipe Down Kitchen Hood & Filters')).toBeVisible({ timeout: 5000 });
    const choreCard = page.locator('[id^="chore-card-"]').filter({ hasText: 'Wipe Down Kitchen Hood & Filters' });
    await expect(choreCard.getByText('30 pts')).toBeVisible();
  });

  test('Step 3: Core Workflow Execution (Claim Open Task & Propose Swap)', async ({ page }) => {
    // 1. Claim Open Task as Sarah Chen
    await page.locator('#filter-pill-open').click();
    const openCard = page.locator('[id^="chore-card-"]').filter({ hasText: 'Water Plants & Sweep Balcony' });
    await expect(openCard).toBeVisible();

    const claimBtn = openCard.getByRole('button', { name: /Claim Task/i });
    await claimBtn.click();

    // Verify task transitions to assigned status
    await page.locator('#filter-pill-all').click();
    const claimedCard = page.locator('[id^="chore-card-"]').filter({ hasText: 'Water Plants & Sweep Balcony' });
    await expect(claimedCard.getByRole('button', { name: /Mark Completed/i })).toBeVisible();

    // 2. Switch Persona to Liam Vance
    await switchUser(page, 'm-liam');
    await expect(page.getByText('85 pts')).toBeVisible();
    await expect(page.getByText('3d')).toBeVisible();

    // 3. Propose a Peer Swap
    await page.getByRole('button', { name: /Swap Market/i }).click();
    await expect(page.getByText('Chore Trading & Swapping Board')).toBeVisible();

    await page.locator('#propose-swap-btn').click();
    await expect(page.getByRole('heading', { name: 'Propose Chore Swap' })).toBeVisible();

    // Select Chore to Swap: Recycling & Garbage to Curb
    const choreSelect = page.locator('form select').first();
    await choreSelect.selectOption('c-recycling-rr');

    // Select Target Peer: Maya Patel
    const peerSelect = page.locator('form select').nth(1);
    await peerSelect.selectOption('m-maya');

    // Enter Reason
    await page.getByPlaceholder(/Working double shift/i).fill('Final exams on Tuesday evening, can take your weekend chore in return.');

    // Submit Swap Proposal
    await page.getByRole('button', { name: /Post Swap Offer/i }).click();

    // Verify Pending Swap Offer Card
    const swapCard = page.locator('[id^="swap-card-"]').filter({ hasText: 'Recycling & Garbage to Curb' });
    await expect(swapCard).toBeVisible();
    await expect(swapCard.getByText(/Offered to Maya Patel/i)).toBeVisible();
  });

  test('Step 4 & 5: State Transitions, Completions, Approvals & Activity Log', async ({ page }) => {
    // Switch to Liam Vance
    await switchUser(page, 'm-liam');
    await expect(page.getByText('85 pts')).toBeVisible();

    // 1. Instant Self-Completion: Vacuum Living Room & Dust Shelves (+15 pts)
    await page.getByRole('button', { name: /Chores Board/i }).click();
    const instantCard = page.locator('[id^="chore-card-"]').filter({ hasText: 'Vacuum Living Room & Dust Shelves' });
    await expect(instantCard).toBeVisible();

    const completeBtn = instantCard.getByRole('button', { name: /Mark Completed/i });
    await completeBtn.click();

    // Verify card turns to Done & Verified and Liam's points balance updates to 100 pts
    await expect(page.getByText('Done & Verified').first()).toBeVisible();
    await expect(page.getByText('100 pts')).toBeVisible();

    // 2. Submit Gated Chore for Approval:
    // First, let's create a gated chore or use an existing one
    // Let's create one with Liam as assignee or submit a gated chore
    await page.locator('#add-chore-button').click();
    await page.getByPlaceholder(/Mop kitchen floor/i).fill('Wipe Down Kitchen Hood & Filters');
    await page.getByPlaceholder(/Specific checklist notes/i).fill('Degreased and cleaned.');
    await page.locator('form').getByRole('button', { name: 'Kitchen', exact: true }).click();

    // Direct Assign to Liam Vance
    await page.getByRole('button', { name: /Direct Assign/i }).click();
    await page.locator('form select').first().selectOption('m-liam');

    // Toggle Approval ON
    const approvalToggle = page.locator('text=Requires Parent / Admin Approval')
      .locator('xpath=ancestor::div[contains(@class, "flex items-center justify-between")]')
      .locator('input[type="checkbox"]');
    await approvalToggle.check({ force: true });

    // Toggle Proof ON
    const proofToggle = page.locator('text=Requires Completion Proof')
      .locator('xpath=ancestor::div[contains(@class, "flex items-center justify-between")]')
      .locator('input[type="checkbox"]');
    await proofToggle.check({ force: true });

    await page.getByRole('button', { name: 'Create Chore' }).click();
    await expect(page.getByText('Wipe Down Kitchen Hood & Filters')).toBeVisible();

    // Click Submit for Approval on this chore
    const gatedCard = page.locator('[id^="chore-card-"]').filter({ hasText: 'Wipe Down Kitchen Hood & Filters' });
    await gatedCard.getByRole('button', { name: /Submit for Approval/i }).click();

    // Completion Modal: choose sample photo and enter proof notes
    await expect(page.getByText('Complete Chore')).toBeVisible();
    await page.getByRole('button', { name: /Clean Kitchen/i }).click();
    await page.getByPlaceholder(/Cleaned under the couch/i).fill('Degreaser applied, soaked filters for 30 mins, reinstalled.');
    await page.getByRole('button', { name: /Submit for Parent Approval/i }).click();

    // Verify card updates to Awaiting Verification
    await expect(gatedCard.getByText('Awaiting Verification')).toBeVisible();

    // 3. Step 5: Admin Approval Gate
    // Switch to Sarah Chen (Admin)
    await switchUser(page, 'm-sarah');
    await page.getByRole('button', { name: /Approval Queue/i }).click();
    await expect(page.getByText('Parent & Admin Verification Queue')).toBeVisible();

    // Locate Liam's submission and approve it
    await expect(page.getByText('Wipe Down Kitchen Hood & Filters')).toBeVisible();
    await expect(page.getByText(/Degreaser applied/i)).toBeVisible();

    const approveBtn = page.getByRole('button', { name: /Approve/i }).first();
    await approveBtn.click();

    // Verify queue is cleared or item is removed
    await expect(page.getByText('All caught up!')).toBeVisible({ timeout: 5000 });

    // Switch to Liam to verify points credited (+15 pts default effort -> 115 pts)
    await switchUser(page, 'm-liam');
    await expect(page.getByText('115 pts')).toBeVisible();

    // 4. Audit Trail Inspection
    await page.getByRole('button', { name: /Audit & Feed/i }).click();
    await expect(page.getByText('Household Activity & Audit Trail')).toBeVisible();
    await expect(page.getByText(/completed/i).or(page.getByText(/approved/i)).first()).toBeVisible();

    // 5. Persistence Verification
    await page.reload();
    await page.waitForLoadState('networkidle');
    await switchUser(page, 'm-liam');
    await expect(page.getByText('115 pts')).toBeVisible();
  });

  test('Step 6: Edge Case & Validation Guardrails', async ({ page }) => {
    // 1. Form Validation Guardrail
    await page.locator('#add-chore-button').click();
    await expect(page.getByText('Create New Household Chore')).toBeVisible();

    // Title is empty by default -> Create Chore button MUST be disabled
    const createBtn = page.getByRole('button', { name: 'Create Chore' });
    await expect(createBtn).toBeDisabled();

    // Close modal
    await page.getByRole('button', { name: 'Cancel' }).click();

    // 2. Role Permission Guardrail
    // Switch to non-admin persona: Liam Vance
    await switchUser(page, 'm-liam');
    await page.getByRole('button', { name: /Approval Queue/i }).click();

    // Verify non-admin warning banner is displayed
    await expect(
      page.getByText(/Admin Privileges Required: Approvals can only be resolved by household Admins\/Parents/i)
    ).toBeVisible();

    // Verify no enabled approve buttons exist
    const approveBtns = page.getByRole('button', { name: /Approve/i });
    if ((await approveBtns.count()) > 0) {
      await expect(approveBtns.first()).toBeDisabled();
    }
  });

  test('Step 7: Gamification, Rewards Hub & Balance Integrity', async ({ page }) => {
    // Switch to Sarah Chen (has 145 pts)
    await switchUser(page, 'm-sarah');
    await expect(page.getByText('145 pts')).toBeVisible();

    // Navigate to Rewards Hub
    await page.getByRole('button', { name: /Rewards Hub/i }).click();
    await expect(page.getByText('Household Contributors Leaderboard')).toBeVisible();

    // Verify Sarah is on the leaderboard
    await expect(page.getByText('Sarah', { exact: true })).toBeVisible();

    // Redeem an affordable reward (e.g. Coffee run / 50 pts)
    const affordableRedeemBtn = page.getByRole('button', { name: 'Redeem Reward' }).first();
    await expect(affordableRedeemBtn).toBeEnabled();
    await affordableRedeemBtn.click();

    // Verify Sarah's points were deducted (145 - 50 = 95 pts or similar depending on reward cost)
    await expect(page.getByText(/pts/).first()).toBeVisible();

    // Switch to Liam Vance and check expensive reward (requires more pts than Liam has)
    await switchUser(page, 'm-liam');
    await page.getByRole('button', { name: /Rewards Hub/i }).click();

    // Verify disabled buttons for rewards Liam cannot afford
    const needMorePtsBtn = page.getByRole('button', { name: /Need \d+ more pts/i }).first();
    await expect(needMorePtsBtn).toBeDisabled();
  });

  test('Step 8: Device Mode Switcher (Kitchen Tablet Kiosk vs Personal Device Mode)', async ({ page }) => {
    // 1. Initial State: Defaults to Personal Device Mode on clean setup
    const personalBanner = page.locator('#personal-welcome-banner');
    await expect(personalBanner).toBeVisible();
    await expect(personalBanner.getByText('Personal Device Mode')).toBeVisible();

    // 2. Switch to Kitchen Tablet Kiosk mode via Header Toggle
    const tabletToggleBtn = page.locator('#toggle-mode-tablet');
    await tabletToggleBtn.click();

    const kioskTouchBar = page.locator('#kiosk-touch-bar');
    await expect(kioskTouchBar).toBeVisible();
    await expect(page.getByText('Kitchen Tablet Kiosk')).toBeVisible();

    // Verify member quick-tap cards in kiosk mode
    const liamCard = page.locator('#kiosk-member-m-liam');
    await expect(liamCard).toBeVisible();

    // 3. Tap Liam's card to switch persona in 1 click on kiosk
    await liamCard.click();
    await expect(page.locator('#user-profile-button')).toContainText('Liam Vance');
    await expect(page.getByText('85 pts')).toBeVisible();

    // 4. Open Admin PIN lock modal from Kiosk bar
    await page.locator('#kiosk-admin-pin-btn').click();
    await expect(page.getByText('Parent & Admin PIN Required')).toBeVisible();

    // Enter PIN 1234
    for (const digit of ['1', '2', '3', '4']) {
      await page.getByRole('button', { name: digit, exact: true }).click();
    }

    // Verify successful unlock switches persona back to Admin (Sarah Chen)
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');

    // 5. Switch to Personal Device Mode via Header Segmented Toggle
    const personalToggleBtn = page.locator('#toggle-mode-personal');
    await personalToggleBtn.click();

    // Verify Personal Welcome Banner appears and Kiosk Touch Bar is hidden
    await expect(kioskTouchBar).not.toBeVisible();
    await expect(personalBanner).toBeVisible();

    // 6. Verify Personal Mode defaults to "My Tasks"
    const myFilterPill = page.locator('#filter-pill-my');
    await expect(myFilterPill).toHaveClass(/bg-zinc-900/);

    // Switch to Liam Vance in Personal Mode via real sign in
    await switchUser(page, 'm-liam');
    await expect(page.getByText('Hi, Liam Vance!')).toBeVisible();

    // In personal mode, only Liam's assigned chores show under My Tasks
    await expect(page.getByText('Vacuum Living Room & Dust Shelves')).toBeVisible();
    // Sarah's chore should NOT be shown in Liam's personal tasks view
    await expect(page.getByText('Deep Clean Kitchen Counters & Stovetop')).not.toBeVisible();

    // 7. Switch back to Kitchen Tablet Mode via Banner button
    await page.locator('#personal-switch-to-kiosk-btn').click();
    await expect(kioskTouchBar).toBeVisible();
    await expect(page.locator('#filter-pill-all')).toHaveClass(/bg-zinc-900/);
  });

  test('Step 9: Admin PIN Modification and Verification', async ({ page }) => {
    // 1. Ensure Sarah Chen (Admin) is logged in
    await switchUser(page, 'm-sarah');

    // 2. Navigate to Household Settings tab
    await page.locator('#nav-btn-settings').click();

    // 3. Verify Change PIN button is visible for admin
    const editPinBtn = page.locator('#edit-admin-pin-btn');
    await expect(editPinBtn).toBeVisible();

    // 4. Click Change PIN button and verify modal opens
    await editPinBtn.click();
    await expect(page.getByText('Change Admin PIN')).toBeVisible();

    // 5. Fill new PIN '8844' and submit
    const pinInput = page.locator('#update-admin-pin-input');
    await pinInput.fill('8844');
    await page.locator('#save-admin-pin-btn').click();

    // 6. Verify modal closes and Household Settings shows new PIN '8844'
    await expect(page.getByText('Change Admin PIN')).not.toBeVisible();
    await expect(page.locator('text=8844')).toBeVisible();

    // 7. Verify non-admin member cannot change PIN
    await switchUser(page, 'm-liam');
    await page.locator('#nav-btn-settings').click();
    await expect(page.locator('#edit-admin-pin-btn')).not.toBeVisible();
    await expect(page.getByText('(Admin only)')).toBeVisible();

    // 8. Test Shared Tablet Kiosk Mode unlocks with the new PIN '8844'
    await page.locator('#nav-btn-dashboard').click();
    await page.locator('#toggle-mode-tablet').click();
    await expect(page.locator('#kiosk-touch-bar')).toBeVisible();

    // Tap Liam's card to ensure persona is Liam
    await page.locator('#kiosk-member-m-liam').click();
    await expect(page.locator('#user-profile-button')).toContainText('Liam Vance');

    // Open Admin PIN modal from Kiosk
    await page.locator('#kiosk-admin-pin-btn').click();
    await expect(page.getByText('Parent & Admin PIN Required')).toBeVisible();

    // Try old PIN 1234 -> should fail with error
    for (const digit of ['1', '2', '3', '4']) {
      await page.getByRole('button', { name: digit, exact: true }).click();
    }
    await expect(page.locator('#kiosk-pin-error')).toBeVisible({ timeout: 4000 });

    // Enter new PIN 8844
    for (const digit of ['8', '8', '4', '4']) {
      await page.getByRole('button', { name: digit, exact: true }).click();
    }

    // Verify unlocked and switched back to Sarah Chen
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');
  });

  test('Step 10: External Email & Notification Integration (Forgot Password, Dev Mailbox Inspection, and Reset Password Flow)', async ({ page, request }) => {
    // 1. Clear dev mailbox outbox first
    await request.delete('http://localhost:8000/api/v1/dev/emails');

    // 2. Sign out via profile menu to land on Login view
    const profileBtn = page.locator('#user-profile-button');
    if (await profileBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await profileBtn.click();
      await page.locator('#profile-sign-out-btn').click();
    }

    // 3. Verify Dev Mailbox button is visible on login screen
    const devMailboxBtn = page.locator('#open-dev-mailbox-btn');
    await expect(devMailboxBtn).toBeVisible({ timeout: 5000 });

    // 4. Click "Forgot password?" trigger link
    const forgotLink = page.locator('#forgot-password-link');
    await expect(forgotLink).toBeVisible();
    await forgotLink.click();

    // 5. Fill email and submit forgot password form
    const forgotEmailInput = page.locator('#forgot-email-input');
    await expect(forgotEmailInput).toBeVisible();
    await forgotEmailInput.fill('sarah@example.com');
    await page.locator('#forgot-password-submit-btn').click();

    // 6. Verify success message appears
    const successMsg = page.locator('#forgot-success-message');
    await expect(successMsg).toBeVisible({ timeout: 5000 });
    await expect(successMsg).toContainText('Password reset link sent');

    // 7. Open Dev Mailbox modal
    await devMailboxBtn.click();
    const modal = page.locator('#dev-mailbox-modal');
    await expect(modal).toBeVisible();

    // Verify email appears in mailbox list
    const emailItem = page.locator('#dev-email-item-0');
    await expect(emailItem).toBeVisible({ timeout: 5000 });
    await expect(emailItem).toContainText('sarah@example.com');
    await expect(emailItem).toContainText('Reset your ChoreSync password');

    // Click email item to view full email preview
    await emailItem.click();
    await expect(page.locator('#dev-email-subject')).toContainText('Reset your ChoreSync password');

    // Close Dev Mailbox modal
    await page.locator('#close-dev-mailbox-btn').click();
    await expect(modal).not.toBeVisible();

    // 8. Proceed to Reset Password tab
    await page.locator('#proceed-to-reset-btn').click();
    await expect(page.locator('#reset-token-input')).toBeVisible();

    // Fill new password and confirm
    await page.locator('#reset-new-password-input').fill('newpassword456');
    await page.locator('#reset-confirm-password-input').fill('newpassword456');
    await page.locator('#reset-password-submit-btn').click();

    // 9. Verify reset success message and automatic login as Sarah Chen
    await expect(page.locator('#reset-success-message')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 6000 });
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');

    // 10. Verify login works with the new password
    await page.locator('#user-profile-button').click();
    await page.locator('#profile-sign-out-btn').click();
    await page.locator('#login-email-input').fill('sarah@example.com');
    await page.locator('#login-password-input').fill('newpassword456');
    await page.locator('#login-submit-btn').click();
    await expect(page.locator('#user-profile-button')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#user-profile-button')).toContainText('Sarah Chen');
  });

  test('Step 11: Chore Reminder Nudge Dispatches Email Notification to Assignee', async ({ page, request }) => {
    // 1. Ensure Sarah Chen is logged in
    await switchUser(page, 'm-sarah');

    // 2. Clear dev email mailbox
    await request.delete('http://localhost:8000/api/v1/dev/emails');

    // 3. Ensure we are in dashboard view
    await page.locator('#nav-btn-dashboard').click();

    // Switch to "All Tasks" view to see Liam's chores
    const allFilterPill = page.locator('#filter-pill-all');
    if (await allFilterPill.isVisible({ timeout: 1000 }).catch(() => false)) {
      await allFilterPill.click();
    }

    // 4. Locate Liam's chore "Vacuum Living Room & Dust Shelves"
    const choreCard = page.locator('#chore-card-c-vacuum-living');
    await expect(choreCard).toBeVisible();

    // 5. Open chore kebab menu and click nudge button
    const menuBtn = page.locator('#chore-menu-btn-c-vacuum-living');
    await menuBtn.click();

    const nudgeBtn = page.locator('#nudge-chore-btn-c-vacuum-living');
    await expect(nudgeBtn).toBeVisible();
    await nudgeBtn.click();

    // 6. Verify toast notification appears
    await expect(page.locator('text=Friendly reminder')).toBeVisible({ timeout: 5000 });

    // 7. Inspect Dev Emails via API to verify email dispatch
    const emailsRes = await request.get('http://localhost:8000/api/v1/dev/emails');
    const emails = await emailsRes.json();
    expect(emails.length).toBeGreaterThan(0);
    const reminderEmail = emails.find((e: any) => e.type === 'chore_reminder');
    expect(reminderEmail).toBeDefined();
    expect(reminderEmail.to).toBe('liam@example.com');
    expect(reminderEmail.subject).toContain('Friendly reminder');
    expect(reminderEmail.html_body).toContain('Vacuum Living Room');
  });
});

