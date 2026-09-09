import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { api } from './services/api';
import {
  Household,
  Member,
  Chore,
  ChoreCompletion,
  ChoreSwapRequest,
  RewardItem,
  RewardRedemption,
  ActivityLog,
  HouseholdMode,
  MemberRole,
  ChoreCategory,
} from './types';
import { ToastProvider, useToast } from './components/Toast';
import { Navbar } from './components/Navbar';
import { Sidebar, NavTab } from './components/Sidebar';
import { ChoreCard } from './components/ChoreCard';
import { ChoreCreationModal } from './components/ChoreCreationModal';
import { CompletionModal } from './components/CompletionModal';
import { ApprovalQueueView } from './components/ApprovalQueueView';
import { SwapMarketplaceView } from './components/SwapMarketplaceView';
import { RewardsHubView } from './components/RewardsHubView';
import { ActivityFeedView } from './components/ActivityFeedView';
import { HouseholdView } from './components/HouseholdView';
import { ChoreDetailsModal } from './components/ChoreDetailsModal';
import {
  Search,
  CheckSquare,
  Sparkles,
  Layers,
  Plus,
  Flame,
  Coins,
  CheckCircle2,
  Clock,
  AlertTriangle,
  RotateCw,
} from 'lucide-react';

function ChoreSyncApp() {
  const { showToast } = useToast();

  // Primary State
  const [households, setHouseholds] = useState<Household[]>([]);
  const [activeHouseholdId, setActiveHouseholdId] = useState<string>('h-roommates');
  const [members, setMembers] = useState<Member[]>([]);
  const [activeMemberId, setActiveMemberId] = useState<string>('m-sarah');

  // Navigation State
  const [currentTab, setCurrentTab] = useState<NavTab>('dashboard');

  // Chores State & Filters
  const [chores, setChores] = useState<Chore[]>([]);
  const [activeViewFilter, setActiveViewFilter] = useState<'all' | 'my' | 'open' | 'approval' | 'overdue'>('all');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Queue & Sub-modules State
  const [pendingApprovals, setPendingApprovals] = useState<
    Array<{ completion: ChoreCompletion; chore: Chore; member: Member }>
  >([]);
  const [swaps, setSwaps] = useState<
    Array<{ swap: ChoreSwapRequest; chore: Chore; requester: Member; targetMember?: Member | null }>
  >([]);
  const [rewards, setRewards] = useState<RewardItem[]>([]);
  const [redemptions, setRedemptions] = useState<
    Array<{ redemption: RewardRedemption; reward: RewardItem; member: Member }>
  >([]);
  const [activities, setActivities] = useState<ActivityLog[]>([]);

  // Modals State
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingChore, setEditingChore] = useState<Chore | null>(null);
  const [completionTargetChore, setCompletionTargetChore] = useState<Chore | null>(null);
  const [detailsTargetChore, setDetailsTargetChore] = useState<Chore | null>(null);

  // Loading State
  const [isLoading, setIsLoading] = useState(true);

  // Active household object
  const activeHousehold = useMemo(() => {
    return households.find((h) => h.id === activeHouseholdId) || null;
  }, [households, activeHouseholdId]);

  // Active member object
  const activeMember = useMemo(() => {
    return members.find((m) => m.id === activeMemberId) || null;
  }, [members, activeMemberId]);

  // Load all data
  const refreshData = useCallback(async () => {
    try {
      const allHouseholds = await api.getHouseholds();
      setHouseholds(allHouseholds);

      if (allHouseholds.length > 0) {
        const currentHouseId = activeHouseholdId || allHouseholds[0].id;
        const currentMembers = await api.getMembers(currentHouseId);
        setMembers(currentMembers);

        // Ensure active member exists in this household
        if (!currentMembers.some((m) => m.id === activeMemberId)) {
          if (currentMembers.length > 0) {
            setActiveMemberId(currentMembers[0].id);
          }
        }

        const [allChores, approvalsList, swapsList, rewardsList, redemptionsList, activitiesList] =
          await Promise.all([
            api.getChores(currentHouseId),
            api.getPendingApprovals(currentHouseId),
            api.getSwaps(currentHouseId),
            api.getRewards(currentHouseId),
            api.getRedemptions(currentHouseId),
            api.getActivities(currentHouseId),
          ]);

        setChores(allChores);
        setPendingApprovals(approvalsList);
        setSwaps(swapsList);
        setRewards(rewardsList);
        setRedemptions(redemptionsList);
        setActivities(activitiesList);
      }
    } catch (err) {
      console.error('Error loading ChoreSync data', err);
      showToast('Error syncing data. Retrying...', 'warning');
    } finally {
      setIsLoading(false);
    }
  }, [activeHouseholdId, activeMemberId, showToast]);

  useEffect(() => {
    refreshData();
  }, [refreshData]);

  // Household switch handler
  const handleSelectHousehold = (newHouseId: string) => {
    setActiveHouseholdId(newHouseId);
    // Switch default persona based on household
    if (newHouseId === 'h-roommates') {
      setActiveMemberId('m-sarah');
    } else if (newHouseId === 'h-family') {
      setActiveMemberId('m-david-fam');
    } else if (newHouseId === 'h-couple') {
      setActiveMemberId('m-alex');
    }
  };

  // Member persona switch handler
  const handleSelectMember = (newMemberId: string) => {
    setActiveMemberId(newMemberId);
    const m = members.find((x) => x.id === newMemberId);
    if (m) {
      showToast(`Switched persona to ${m.name} (${m.role.toUpperCase()})`, 'info');
    }
  };

  // Quick switch to admin helper
  const handleSwitchToAdmin = () => {
    const admin = members.find((m) => m.role === 'admin');
    if (admin) {
      handleSelectMember(admin.id);
    }
  };

  // Filtered Chores for the dashboard
  const filteredChores = useMemo(() => {
    return chores.filter((c) => {
      // Search query
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const matchTitle = c.title.toLowerCase().includes(q);
        const matchDesc = c.description ? c.description.toLowerCase().includes(q) : false;
        if (!matchTitle && !matchDesc) return false;
      }

      // Category filter
      if (selectedCategory !== 'all' && c.category !== selectedCategory) {
        return false;
      }

      // View Filter
      if (activeViewFilter === 'my') {
        return activeMember && c.current_assignee_id === activeMember.id;
      }
      if (activeViewFilter === 'open') {
        return c.assignment_type === 'open_pool' && !c.current_assignee_id;
      }
      if (activeViewFilter === 'approval') {
        return c.status === 'pending_approval';
      }
      if (activeViewFilter === 'overdue') {
        return c.status === 'overdue';
      }

      return true;
    });
  }, [chores, searchQuery, selectedCategory, activeViewFilter, activeMember]);

  // Counts
  const completedCount = useMemo(() => chores.filter((c) => c.status === 'completed').length, [chores]);
  const activeTasksCount = useMemo(() => chores.filter((c) => c.status !== 'completed').length, [chores]);
  const myAssignedChores = useMemo(() => {
    return activeMember ? chores.filter((c) => c.current_assignee_id === activeMember.id && c.status !== 'completed') : [];
  }, [chores, activeMember]);

  // Handlers
  const handleStartComplete = (chore: Chore) => {
    if (chore.requires_approval || chore.requires_proof) {
      setCompletionTargetChore(chore);
    } else {
      handleInstantComplete(chore);
    }
  };

  const handleInstantComplete = async (chore: Chore) => {
    if (!activeMember) return;
    try {
      const res = await api.completeChore(chore.id, activeMember.id);
      showToast(`Chore completed! You earned +${res.pointsAwarded} pts.`, 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error completing chore', 'warning');
    }
  };

  const handleSubmitProofCompletion = async (proof: { proof_notes?: string; proof_photo_url?: string }) => {
    if (!completionTargetChore || !activeMember) return;
    try {
      const res = await api.completeChore(completionTargetChore.id, activeMember.id, proof);
      if (completionTargetChore.requires_approval) {
        showToast('Submitted for Parent/Admin verification!', 'info');
      } else {
        showToast(`Chore completed! You earned +${res.pointsAwarded} pts.`, 'success');
      }
      setCompletionTargetChore(null);
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error submitting chore', 'warning');
    }
  };

  const handleClaimChore = async (choreId: string) => {
    if (!activeMember) return;
    try {
      await api.claimChore(choreId, activeMember.id);
      showToast(`You claimed this task!`, 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error claiming chore', 'warning');
    }
  };

  const handleSendNudge = async (chore: Chore) => {
    if (!activeMember) return;
    try {
      const res = await api.sendNudge(chore.id, activeMember.id);
      showToast(res.message, 'nudge', 'Reminder Dispatched');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error sending nudge', 'warning');
    }
  };

  const handleApproveCompletion = async (completionId: string) => {
    if (!activeMember) return;
    try {
      await api.approveChore(completionId, activeMember.id);
      showToast('Chore approved and points awarded to member!', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error approving chore', 'warning');
    }
  };

  const handleRejectCompletion = async (completionId: string, reason: string) => {
    if (!activeMember) return;
    try {
      await api.rejectChore(completionId, activeMember.id, reason);
      showToast('Re-work requested with feedback.', 'info');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error returning chore', 'warning');
    }
  };

  const handleAcceptSwap = async (swapId: string) => {
    if (!activeMember) return;
    try {
      await api.acceptSwap(swapId, activeMember.id);
      showToast('Chore trade accepted! Task reassigned to you.', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error accepting swap', 'warning');
    }
  };

  const handleRejectSwap = async (swapId: string) => {
    try {
      await api.rejectSwap(swapId);
      showToast('Swap declined.', 'info');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error rejecting swap', 'warning');
    }
  };

  const handleCancelSwap = async (swapId: string) => {
    try {
      await api.cancelSwap(swapId);
      showToast('Swap proposal cancelled.', 'info');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error cancelling swap', 'warning');
    }
  };

  const handleRequestSwap = async (payload: { chore_id: string; requester_id: string; target_member_id?: string | null; reason?: string }) => {
    try {
      await api.requestSwap(payload);
      showToast('Swap request posted to housemates!', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error proposing swap', 'warning');
    }
  };

  const handleRedeemReward = async (rewardId: string) => {
    if (!activeMember) return;
    try {
      await api.redeemReward(rewardId, activeMember.id);
      showToast('Reward claimed! Awaiting parent/admin fulfillment.', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Could not redeem reward', 'warning');
    }
  };

  const handleFulfillReward = async (redemptionId: string) => {
    try {
      await api.fulfillReward(redemptionId);
      showToast('Perk marked fulfilled!', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error fulfilling reward', 'warning');
    }
  };

  const handleCreateReward = async (payload: { title: string; description: string; points_cost: number; icon?: string }) => {
    try {
      await api.createReward(activeHouseholdId, payload);
      showToast('New reward added to household catalog!', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error creating reward', 'warning');
    }
  };

  const handleSaveChore = async (payload: any) => {
    try {
      if (editingChore) {
        await api.updateChore(editingChore.id, payload);
        showToast('Chore updated successfully.', 'success');
      } else {
        await api.createChore(payload);
        showToast('New chore added to household.', 'success');
      }
      setEditingChore(null);
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error saving chore', 'warning');
    }
  };

  const handleDeleteChore = async (choreId: string) => {
    try {
      await api.deleteChore(choreId);
      showToast('Chore deleted.', 'info');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error deleting chore', 'warning');
    }
  };

  const handleUpdateHouseholdMode = async (mode: HouseholdMode) => {
    try {
      await api.updateHouseholdSettings(activeHouseholdId, { default_mode: mode });
      showToast(`Household dynamic updated to ${mode.toUpperCase()} mode.`, 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error updating settings', 'warning');
    }
  };

  const handleAddMember = async (payload: { name: string; role: MemberRole }) => {
    try {
      await api.addMember(activeHouseholdId, payload);
      showToast(`Added ${payload.name} to household roster.`, 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error adding member', 'warning');
    }
  };

  const handleResetDemo = async () => {
    try {
      await api.resetToDefaultData();
      showToast('Reset to default seed data.', 'info');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error resetting store', 'warning');
    }
  };

  return (
    <div className="min-h-screen bg-zinc-100/70 text-zinc-900 flex flex-col font-sans selection:bg-zinc-900 selection:text-white pb-16 md:pb-0">
      {/* Top Navbar */}
      <Navbar
        households={households}
        activeHousehold={activeHousehold}
        onSelectHousehold={handleSelectHousehold}
        members={members}
        activeMember={activeMember}
        onSelectMember={handleSelectMember}
        onOpenCreateChore={() => {
          setEditingChore(null);
          setIsCreateModalOpen(true);
        }}
        onOpenActivities={() => setCurrentTab('activity')}
        unreadActivitiesCount={activities.filter((a) => !a.read).length}
        pendingApprovalsCount={pendingApprovals.length}
        onModeChange={handleUpdateHouseholdMode}
      />

      {/* Main Body */}
      <div className="flex-1 flex max-w-7xl w-full mx-auto">
        {/* Sidebar */}
        <Sidebar
          currentTab={currentTab}
          onSelectTab={setCurrentTab}
          pendingApprovalsCount={pendingApprovals.length}
          pendingSwapsCount={swaps.filter((s) => s.swap.status === 'pending').length}
          householdMode={activeHousehold?.settings.default_mode || 'flatmate'}
          userRole={activeMember?.role || 'member'}
          choresCount={activeTasksCount}
          completedChoresCount={completedCount}
        />

        {/* Tab Content Canvas */}
        <main className="flex-1 p-4 sm:p-6 lg:p-8 min-w-0">
          {/* TAB 1: CHORES DASHBOARD */}
          {currentTab === 'dashboard' && (
            <div className="space-y-6">
              {/* Dashboard Header & Search Controls */}
              <div className="bg-white rounded-2xl border border-zinc-200 p-5 space-y-4 shadow-xs">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div>
                    <h1 className="text-xl font-black text-zinc-900 tracking-tight">Household Chores Board</h1>
                    <p className="text-xs text-zinc-500 mt-0.5">
                      {activeTasksCount} active tasks remaining • {completedCount} verified completed
                    </p>
                  </div>

                  {/* Search bar */}
                  <div className="relative max-w-xs w-full">
                    <Search className="w-4 h-4 text-zinc-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                    <input
                      type="text"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder="Search chores or instructions..."
                      className="w-full text-xs font-medium pl-9 pr-3 py-2 rounded-xl bg-zinc-50 border border-zinc-200 focus:outline-none focus:ring-2 focus:ring-zinc-900/10 focus:bg-white transition-all"
                    />
                  </div>
                </div>

                {/* Filter Pills */}
                <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0">
                  <button
                    id="filter-pill-all"
                    onClick={() => setActiveViewFilter('all')}
                    className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer shrink-0 ${
                      activeViewFilter === 'all'
                        ? 'bg-zinc-900 text-white shadow-xs'
                        : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                    }`}
                  >
                    All Chores ({chores.length})
                  </button>

                  <button
                    id="filter-pill-my"
                    onClick={() => setActiveViewFilter('my')}
                    className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer shrink-0 ${
                      activeViewFilter === 'my'
                        ? 'bg-zinc-900 text-white shadow-xs'
                        : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                    }`}
                  >
                    My Tasks ({myAssignedChores.length})
                  </button>

                  <button
                    id="filter-pill-open"
                    onClick={() => setActiveViewFilter('open')}
                    className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer shrink-0 ${
                      activeViewFilter === 'open'
                        ? 'bg-zinc-900 text-white shadow-xs'
                        : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                    }`}
                  >
                    Open Pool ({chores.filter((c) => c.assignment_type === 'open_pool' && !c.current_assignee_id).length})
                  </button>

                  <button
                    id="filter-pill-approval"
                    onClick={() => setActiveViewFilter('approval')}
                    className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer shrink-0 ${
                      activeViewFilter === 'approval'
                        ? 'bg-zinc-900 text-white shadow-xs'
                        : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                    }`}
                  >
                    Needs Review ({pendingApprovals.length})
                  </button>

                  <button
                    id="filter-pill-overdue"
                    onClick={() => setActiveViewFilter('overdue')}
                    className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer shrink-0 ${
                      activeViewFilter === 'overdue'
                        ? 'bg-zinc-900 text-white shadow-xs'
                        : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                    }`}
                  >
                    Overdue ({chores.filter((c) => c.status === 'overdue').length})
                  </button>
                </div>

                {/* Category Secondary Filter */}
                <div className="flex items-center gap-1.5 overflow-x-auto text-xs text-zinc-600 pt-2 border-t border-zinc-100">
                  <span className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider shrink-0 mr-1">
                    Category:
                  </span>
                  {['all', 'cleaning', 'kitchen', 'yard', 'pets', 'maintenance', 'other'].map((cat) => (
                    <button
                      key={cat}
                      onClick={() => setSelectedCategory(cat)}
                      className={`capitalize px-2.5 py-1 rounded-lg text-xs font-semibold cursor-pointer transition-colors shrink-0 ${
                        selectedCategory === cat
                          ? 'bg-zinc-200 text-zinc-900'
                          : 'hover:bg-zinc-100 text-zinc-500'
                      }`}
                    >
                      {cat}
                    </button>
                  ))}
                </div>
              </div>

              {/* Chores Grid */}
              {filteredChores.length === 0 ? (
                <div className="bg-white rounded-2xl border border-zinc-200 p-12 text-center">
                  <CheckSquare className="w-10 h-10 text-zinc-300 mx-auto mb-3" />
                  <h3 className="font-bold text-zinc-900 text-sm">No matching chores found</h3>
                  <p className="text-xs text-zinc-500 max-w-sm mx-auto mt-1 mb-4">
                    There are no chores in this view. Try relaxing your filters or create a new household task!
                  </p>
                  <button
                    onClick={() => {
                      setEditingChore(null);
                      setIsCreateModalOpen(true);
                    }}
                    className="inline-flex items-center gap-1.5 bg-zinc-900 hover:bg-zinc-800 text-white text-xs font-bold px-4 py-2 rounded-xl transition-colors cursor-pointer"
                  >
                    <Plus className="w-4 h-4" />
                    Add Chore
                  </button>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {filteredChores.map((chore) => {
                    const assignee = members.find((m) => m.id === chore.current_assignee_id);
                    return (
                      <ChoreCard
                        key={chore.id}
                        chore={chore}
                        assignee={assignee}
                        currentMember={activeMember}
                        onComplete={handleStartComplete}
                        onClaim={handleClaimChore}
                        onRequestSwap={(c) => {
                          setCurrentTab('swaps');
                          // Small timeout to allow tab switch
                          setTimeout(() => {
                            showToast(`Propose swap for "${c.title}" below`, 'info');
                          }, 100);
                        }}
                        onSendNudge={handleSendNudge}
                        onOpenDetails={(c) => setDetailsTargetChore(c)}
                        onEdit={(c) => {
                          setEditingChore(c);
                          setIsCreateModalOpen(true);
                        }}
                        onDelete={handleDeleteChore}
                      />
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {/* TAB 2: APPROVAL QUEUE */}
          {currentTab === 'approvals' && (
            <ApprovalQueueView
              approvals={pendingApprovals}
              currentMember={activeMember}
              onApprove={handleApproveCompletion}
              onReject={handleRejectCompletion}
              onSwitchToAdminPersona={handleSwitchToAdmin}
            />
          )}

          {/* TAB 3: SWAP MARKETPLACE */}
          {currentTab === 'swaps' && (
            <SwapMarketplaceView
              swaps={swaps}
              currentMember={activeMember}
              myAssignedChores={myAssignedChores}
              householdMembers={members}
              onAcceptSwap={handleAcceptSwap}
              onRejectSwap={handleRejectSwap}
              onCancelSwap={handleCancelSwap}
              onRequestSwap={handleRequestSwap}
            />
          )}

          {/* TAB 4: REWARDS HUB */}
          {currentTab === 'rewards' && (
            <RewardsHubView
              rewards={rewards}
              redemptions={redemptions}
              currentMember={activeMember}
              householdMembers={members}
              onRedeem={handleRedeemReward}
              onFulfill={handleFulfillReward}
              onCreateReward={handleCreateReward}
            />
          )}

          {/* TAB 5: AUDIT & FEED */}
          {currentTab === 'activity' && (
            <ActivityFeedView activities={activities} members={members} />
          )}

          {/* TAB 6: HOUSEHOLD & MEMBERS */}
          {currentTab === 'settings' && (
            <HouseholdView
              household={activeHousehold}
              members={members}
              onUpdateMode={handleUpdateHouseholdMode}
              onAddMember={handleAddMember}
              onResetData={handleResetDemo}
            />
          )}
        </main>
      </div>

      {/* MODALS */}
      {/* Create / Edit Chore Modal */}
      <ChoreCreationModal
        isOpen={isCreateModalOpen}
        onClose={() => {
          setIsCreateModalOpen(false);
          setEditingChore(null);
        }}
        householdId={activeHouseholdId}
        members={members}
        currentMember={activeMember}
        editingChore={editingChore}
        onSave={handleSaveChore}
      />

      {/* Completion & Proof Modal */}
      <CompletionModal
        isOpen={Boolean(completionTargetChore)}
        onClose={() => setCompletionTargetChore(null)}
        chore={completionTargetChore}
        currentMember={activeMember}
        onSubmit={handleSubmitProofCompletion}
      />

      {/* Chore Details & Discussion Modal */}
      <ChoreDetailsModal
        isOpen={Boolean(detailsTargetChore)}
        onClose={() => setDetailsTargetChore(null)}
        chore={detailsTargetChore}
        assignee={members.find((m) => m.id === detailsTargetChore?.current_assignee_id)}
        currentMember={activeMember}
        onSendNudge={handleSendNudge}
        onRequestSwap={(c) => {
          setDetailsTargetChore(null);
          setCurrentTab('swaps');
        }}
      />
    </div>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <ChoreSyncApp />
    </ToastProvider>
  );
}
