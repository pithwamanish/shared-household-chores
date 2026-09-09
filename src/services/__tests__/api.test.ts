import { describe, it, expect, beforeEach } from 'vitest';
import { api } from '../api';

describe('ChoreSync Centralized Service Layer', () => {
  beforeEach(async () => {
    await api.resetToDefaultData();
  });

  it('loads seeded households and members properly', async () => {
    const households = await api.getHouseholds();
    expect(households.length).toBe(3);
    expect(households.map((h) => h.name)).toContain('Apartment 4B');
    expect(households.map((h) => h.name)).toContain('The Miller Family');
    expect(households.map((h) => h.name)).toContain('Alex & Sam');

    const roommates = await api.getMembers('h-roommates');
    expect(roommates.length).toBe(4);
    const sarah = roommates.find((m) => m.name.includes('Sarah'));
    expect(sarah).toBeDefined();
    expect(sarah?.role).toBe('admin');
  });

  it('creates an Open Pool chore and allows a member to claim it', async () => {
    const newChore = await api.createChore({
      household_id: 'h-roommates',
      title: 'Wipe Balcony Furniture',
      category: 'yard',
      effort_points: 15,
      assignment_type: 'open_pool',
      recurrence_type: 'none',
      requires_approval: false,
      requires_proof: false,
      created_by: 'm-sarah',
    });

    expect(newChore.id).toBeDefined();
    expect(newChore.status).toBe('unassigned');
    expect(newChore.current_assignee_id).toBeNull();

    // Member claims the chore
    const claimedChore = await api.claimChore(newChore.id, 'm-maya');
    expect(claimedChore.current_assignee_id).toBe('m-maya');
    expect(claimedChore.status).toBe('assigned');
  });

  it('completes an instant chore immediately, awarding points and advancing round-robin rotation', async () => {
    const membersBefore = await api.getMembers('h-roommates');
    const liamBefore = membersBefore.find((m) => m.id === 'm-liam')!;
    const initialPoints = liamBefore.points_balance;
    const initialStreak = liamBefore.streak;

    // c-recycling-rr is assigned to Liam in round-robin sequence ['m-sarah', 'm-liam', 'm-maya', 'm-noah']
    const result = await api.completeChore('c-recycling-rr', 'm-liam');
    expect(result.chore.status).toBe('completed');
    expect(result.pointsAwarded).toBe(20);

    const membersAfter = await api.getMembers('h-roommates');
    const liamAfter = membersAfter.find((m) => m.id === 'm-liam')!;
    expect(liamAfter.points_balance).toBe(initialPoints + 20);
    expect(liamAfter.streak).toBe(initialStreak + 1);

    // Verify next recurring instance was scheduled and rotated to Maya (index 2)
    const allChores = await api.getChores('h-roommates');
    const recyclingInstances = allChores.filter((c) => c.title === 'Recycling & Garbage to Curb' && c.status === 'assigned');
    expect(recyclingInstances.length).toBeGreaterThanOrEqual(1);
    expect(recyclingInstances[0].current_assignee_id).toBe('m-maya');
  });

  it('requires approval when requires_approval is true and credits points only after admin approves', async () => {
    // Create chore requiring parent verification
    const chore = await api.createChore({
      household_id: 'h-family',
      title: 'Practice Piano 20 Minutes',
      category: 'other',
      effort_points: 25,
      assignment_type: 'direct',
      current_assignee_id: 'm-leo-kid',
      recurrence_type: 'daily',
      requires_approval: true,
      requires_proof: true,
      created_by: 'm-sarah-fam',
    });

    const membersBefore = await api.getMembers('h-family');
    const leoBefore = membersBefore.find((m) => m.id === 'm-leo-kid')!;
    const initialPoints = leoBefore.points_balance;

    // Leo submits completion with proof notes
    const submitResult = await api.completeChore(chore.id, 'm-leo-kid', {
      proof_notes: 'Practiced Mozart sonatina and scale drills!',
      proof_photo_url: 'https://example.com/sheet-music.jpg',
    });

    expect(submitResult.chore.status).toBe('pending_approval');
    expect(submitResult.pointsAwarded).toBe(0); // Not awarded yet

    // Leo points should still be unchanged
    const membersMid = await api.getMembers('h-family');
    const leoMid = membersMid.find((m) => m.id === 'm-leo-kid')!;
    expect(leoMid.points_balance).toBe(initialPoints);

    // Parent inspects pending approvals
    const pendingList = await api.getPendingApprovals('h-family');
    const myApproval = pendingList.find((p) => p.chore.id === chore.id);
    expect(myApproval).toBeDefined();
    expect(myApproval?.completion.proof_notes).toBe('Practiced Mozart sonatina and scale drills!');

    // Parent approves
    await api.approveChore(myApproval!.completion.id, 'm-david-fam');

    // Points now credited
    const membersAfter = await api.getMembers('h-family');
    const leoAfter = membersAfter.find((m) => m.id === 'm-leo-kid')!;
    expect(leoAfter.points_balance).toBe(initialPoints + 25);
  });

  it('handles rejection and requests rework', async () => {
    // There is an initial pending completion for Leo in family
    const pending = await api.getPendingApprovals('h-family');
    expect(pending.length).toBeGreaterThan(0);
    const targetComp = pending[0];

    await api.rejectChore(targetComp.completion.id, 'm-david-fam', 'Please make sure closet doors are also closed!');

    const chore = await api.getChore(targetComp.chore.id);
    expect(chore?.status).toBe('assigned');
  });

  it('supports peer chore swap proposals and acceptances', async () => {
    const swap = await api.requestSwap({
      chore_id: 'c-kitchen-deep', // currently assigned to Sarah
      requester_id: 'm-sarah',
      target_member_id: 'm-maya',
      reason: 'Working a double shift on Wednesday',
    });

    expect(swap.status).toBe('pending');

    // Maya accepts the swap
    const accepted = await api.acceptSwap(swap.id, 'm-maya');
    expect(accepted).toBe(true);

    const updatedChore = await api.getChore('c-kitchen-deep');
    expect(updatedChore?.current_assignee_id).toBe('m-maya');
  });

  it('allows members to redeem rewards and admins to fulfill them', async () => {
    const members = await api.getMembers('h-roommates');
    const maya = members.find((m) => m.id === 'm-maya')!;
    const startingBalance = maya.points_balance; // 190

    // Maya redeems 'Pick Apartment Movie Night' (50 pts)
    const redemption = await api.redeemReward('rew-room-movie', 'm-maya');
    expect(redemption.points_spent).toBe(50);
    expect(redemption.status).toBe('requested');

    const updatedMembers = await api.getMembers('h-roommates');
    const mayaAfter = updatedMembers.find((m) => m.id === 'm-maya')!;
    expect(mayaAfter.points_balance).toBe(startingBalance - 50);

    // Admin fulfills
    await api.fulfillReward(redemption.id);
    const redemptions = await api.getRedemptions('h-roommates');
    const fulfilled = redemptions.find((r) => r.redemption.id === redemption.id);
    expect(fulfilled?.redemption.status).toBe('fulfilled');
  });

  it('sends friendly reminder nudges and creates activity log', async () => {
    const nudge = await api.sendNudge('c-vacuum-living', 'm-sarah');
    expect(nudge.success).toBe(true);
    expect(nudge.message).toContain('Liam');

    const activities = await api.getActivities('h-roommates');
    const nudgeAct = activities.find((a) => a.event_type === 'chore_nudge');
    expect(nudgeAct).toBeDefined();
  });
});
