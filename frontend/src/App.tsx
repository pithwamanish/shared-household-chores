import React, { useState, useEffect, useMemo, useCallback, useRef } from 'react';
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
  DeviceMode,
  AuthTokenResponse,
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
import { InviteOnboardingModal } from './components/InviteOnboardingModal';
import { PinVerificationModal } from './components/PinVerificationModal';
import { LoginView } from './components/LoginView';
import { HouseholdInfoModal } from './components/HouseholdInfoModal';
import { httpClient } from './services/httpClient';
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
  X,
  ChevronLeft,
  ChevronRight,
  Tablet,
  Smartphone,
  Lock,
} from 'lucide-react';

function ChoreSyncApp() {
  const { showToast } = useToast();

  // Primary State
  const [households, setHouseholds] = useState<Household[]>([]);
  const [activeHouseholdId, setActiveHouseholdId] = useState<string>(() => {
    return localStorage.getItem('choresync_active_household') || 'h-roommates';
  });
  const [members, setMembers] = useState<Member[]>([]);
  const [activeMemberId, setActiveMemberId] = useState<string>(() => {
    return localStorage.getItem('choresync_active_member') || 'm-sarah';
  });

  // Authentication & Session State
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    const token = localStorage.getItem('choresync_auth_token');
    if (!token) return false;
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      if (payload.exp && payload.exp * 1000 < Date.now()) {
        localStorage.removeItem('choresync_auth_token');
        return false;
      }
      return true;
    } catch {
      localStorage.removeItem('choresync_auth_token');
      return false;
    }
  });
  const [isHouseholdInfoModalOpen, setIsHouseholdInfoModalOpen] = useState(false);
  const [isVerifyingMagic, setIsVerifyingMagic] = useState<boolean>(() => {
    return typeof window !== 'undefined' && new URLSearchParams(window.location.search).has('magic_token');
  });

  // Sync token with httpClient on load
  useEffect(() => {
    const token = localStorage.getItem('choresync_auth_token');
    if (token) {
      httpClient.setAuthToken(token);
    }
  }, []);

  // Device Mode State ('tablet_kiosk' vs 'personal')
  const [deviceMode, setDeviceMode] = useState<DeviceMode>(() => {
    return (localStorage.getItem('choresync_device_mode') as DeviceMode) || 'personal';
  });
  const [isPinModalOpen, setIsPinModalOpen] = useState(false);

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
  const [inviteHousehold, setInviteHousehold] = useState<Household | null>(null);
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);

  // Horizontal Scroll navigation refs & state for filter pills
  const filterScrollRef = useRef<HTMLDivElement>(null);
  const [canScrollFiltersLeft, setCanScrollFiltersLeft] = useState(false);
  const [canScrollFiltersRight, setCanScrollFiltersRight] = useState(false);

  const checkFilterScroll = useCallback(() => {
    const el = filterScrollRef.current;
    if (el) {
      setCanScrollFiltersLeft(el.scrollLeft > 2);
      setCanScrollFiltersRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 2);
    }
  }, []);

  useEffect(() => {
    const el = filterScrollRef.current;
    if (el) {
      checkFilterScroll();
      el.addEventListener('scroll', checkFilterScroll);
      window.addEventListener('resize', checkFilterScroll);
      return () => {
        el.removeEventListener('scroll', checkFilterScroll);
        window.removeEventListener('resize', checkFilterScroll);
      };
    }
  }, [checkFilterScroll, chores]);

  const scrollFilters = (direction: 'left' | 'right') => {
    if (filterScrollRef.current) {
      filterScrollRef.current.scrollBy({
        left: direction === 'left' ? -220 : 220,
        behavior: 'smooth',
      });
    }
  };

  // Horizontal Scroll navigation refs & state for category filters
  const categoryScrollRef = useRef<HTMLDivElement>(null);
  const [canScrollCategoriesLeft, setCanScrollCategoriesLeft] = useState(false);
  const [canScrollCategoriesRight, setCanScrollCategoriesRight] = useState(false);

  const checkCategoryScroll = useCallback(() => {
    const el = categoryScrollRef.current;
    if (el) {
      setCanScrollCategoriesLeft(el.scrollLeft > 2);
      setCanScrollCategoriesRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 2);
    }
  }, []);

  useEffect(() => {
    const el = categoryScrollRef.current;
    if (el) {
      checkCategoryScroll();
      el.addEventListener('scroll', checkCategoryScroll);
      window.addEventListener('resize', checkCategoryScroll);
      return () => {
        el.removeEventListener('scroll', checkCategoryScroll);
        window.removeEventListener('resize', checkCategoryScroll);
      };
    }
  }, [checkCategoryScroll]);

  const scrollCategories = (direction: 'left' | 'right') => {
    if (categoryScrollRef.current) {
      categoryScrollRef.current.scrollBy({
        left: direction === 'left' ? -180 : 180,
        behavior: 'smooth',
      });
    }
  };
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

      // Verify and auto-heal active session against available households
      const token = localStorage.getItem('choresync_auth_token');
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const isExpired = payload.exp && payload.exp * 1000 < Date.now();
          const householdExists = payload.household_id && allHouseholds.some((h) => h.id === payload.household_id);
          if (isExpired || (allHouseholds.length > 0 && !householdExists)) {
            console.warn(`[ChoreSync] Auth session invalid (expired=${isExpired}, householdExists=${householdExists}). Signing out.`);
            localStorage.removeItem('choresync_auth_token');
            httpClient.setAuthToken(null);
            setIsAuthenticated(false);
          }
        } catch {
          localStorage.removeItem('choresync_auth_token');
          httpClient.setAuthToken(null);
          setIsAuthenticated(false);
        }
      }

      if (allHouseholds.length > 0) {
        // Validate activeHouseholdId against available households
        const existingHousehold = allHouseholds.find((h) => h.id === activeHouseholdId);
        const targetHouse = existingHousehold || allHouseholds[0];
        if (!existingHousehold) {
          console.warn(`[ChoreSync] Active household '${activeHouseholdId}' no longer exists in backend. Auto-healing to '${targetHouse.id}'.`);
          setActiveHouseholdId(targetHouse.id);
          localStorage.setItem('choresync_active_household', targetHouse.id);
        }

        const currentHouseId = targetHouse.id;
        const currentMembers = await api.getMembers(currentHouseId);
        setMembers(currentMembers);

        // Ensure active member exists in this household
        if (!currentMembers.some((m) => m.id === activeMemberId)) {
          if (currentMembers.length > 0) {
            setActiveMemberId(currentMembers[0].id);
            localStorage.setItem('choresync_active_member', currentMembers[0].id);
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
    } catch (err: any) {
      console.error('Error loading ChoreSync data', err);
      // Handle deleted household / stale token / unauthorized cleanly
      if (err?.status === 401 || err?.code === 401 || err?.status === 403 || err?.code === 403) {
        localStorage.removeItem('choresync_auth_token');
        httpClient.setAuthToken(null);
        setIsAuthenticated(false);
        showToast('Session expired or invalidated. Please sign in again.', 'info');
      } else if (err?.status === 404 || err?.code === 404) {
        localStorage.removeItem('choresync_active_household');
        localStorage.removeItem('choresync_active_member');
        setActiveHouseholdId('h-roommates');
        setActiveMemberId('m-sarah');
        showToast('Household not found. Reverting to default Apartment 4B...', 'warning');
      } else {
        showToast('Error syncing data. Retrying...', 'warning');
      }
    } finally {
      setIsLoading(false);
    }
  }, [activeHouseholdId, activeMemberId, showToast]);

  useEffect(() => {
    refreshData();
  }, [refreshData]);

  // Check for invite code in URL query params on page load
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const inviteCode = urlParams.get('invite');
    if (inviteCode) {
      api.joinHousehold(inviteCode).then((house) => {
        if (house) {
          setInviteHousehold(house);
          setIsInviteModalOpen(true);
        }
      }).catch((err) => {
        console.warn('Invite link lookup:', err);
      });
    }
  }, []);

  // Handle member joining from 1-click invite link
  const handleJoinFromInvite = async (name: string, role: MemberRole) => {
    if (!inviteHousehold) return;
    try {
      const newMember = await api.addMember(inviteHousehold.id, { name, role });
      handleSelectHousehold(inviteHousehold.id);
      setActiveMemberId(newMember.id);
      localStorage.setItem('choresync_active_member', newMember.id);
      await api.demoLogin(newMember.id).catch(() => {});
      const url = new URL(window.location.href);
      url.searchParams.delete('invite');
      window.history.replaceState({}, '', url.pathname);
      showToast(`Welcome to ${inviteHousehold.name}, ${newMember.name}!`, 'success');
      await refreshData();
    } catch (err: any) {
      showToast(err?.message || 'Failed to join household', 'warning');
    }
  };

  // Household switch handler
  const handleSelectHousehold = (newHouseId: string) => {
    setActiveHouseholdId(newHouseId);
    localStorage.setItem('choresync_active_household', newHouseId);
    let defaultMember = 'm-sarah';
    if (newHouseId === 'h-family') {
      defaultMember = 'm-david-fam';
    } else if (newHouseId === 'h-couple') {
      defaultMember = 'm-alex';
    }
    setActiveMemberId(defaultMember);
    localStorage.setItem('choresync_active_member', defaultMember);
    api.demoLogin(defaultMember).catch(() => {});
  };

  // Member persona switch handler (mints fresh JWT for demo / testing)
  const handleSelectMember = (newMemberId: string) => {
    setActiveMemberId(newMemberId);
    localStorage.setItem('choresync_active_member', newMemberId);
    api.demoLogin(newMemberId).then((res) => {
      if (res?.access_token) {
        localStorage.setItem('choresync_auth_token', res.access_token);
        httpClient.setAuthToken(res.access_token);
      }
    }).catch(() => {});
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

  // Device Mode toggle handler
  const handleToggleDeviceMode = (newMode: DeviceMode) => {
    setDeviceMode(newMode);
    localStorage.setItem('choresync_device_mode', newMode);
    if (newMode === 'personal') {
      setActiveViewFilter('my');
      showToast('Personal Device Mode: Showing your assigned chores', 'info');
    } else {
      setActiveViewFilter('all');
      showToast('Kitchen Tablet Mode: Showing shared household board', 'info');
    }
  };

  // Auth & Login handlers
  const handleLoginSuccess = useCallback((authData: AuthTokenResponse) => {
    if (authData.access_token) {
      localStorage.setItem('choresync_auth_token', authData.access_token);
      httpClient.setAuthToken(authData.access_token);
    }
    if (authData.household?.id) {
      setActiveHouseholdId(authData.household.id);
      localStorage.setItem('choresync_active_household', authData.household.id);
    }
    if (authData.member?.id) {
      setActiveMemberId(authData.member.id);
      localStorage.setItem('choresync_active_member', authData.member.id);
    }
    setIsAuthenticated(true);
    showToast(`Welcome back, ${authData.member.name}!`, 'success');
  }, [showToast]);

  const verifyingTokenRef = useRef<string | null>(null);

  // Handle direct Magic Link URL navigation (?magic_token=...)
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const magicToken = urlParams.get('magic_token');
    if (magicToken && verifyingTokenRef.current !== magicToken) {
      verifyingTokenRef.current = magicToken;
      setIsVerifyingMagic(true);
      api.verifyMagicLink(magicToken.trim())
        .then((authData) => {
          handleLoginSuccess(authData);
          const url = new URL(window.location.href);
          url.searchParams.delete('magic_token');
          window.history.replaceState({}, '', url.pathname + (url.search ? url.search : ''));
        })
        .catch((err) => {
          console.warn('Magic link verification failed:', err);
          showToast(err?.message || 'Invalid or expired magic link. Magic links are single-use and expire after 15 minutes.', 'warning');
        })
        .finally(() => {
          setIsVerifyingMagic(false);
        });
    }
  }, [handleLoginSuccess, showToast]);

  const handleSignOut = () => {
    localStorage.removeItem('choresync_auth_token');
    httpClient.setAuthToken(null);
    setIsAuthenticated(false);
    showToast('Signed out of ChoreSync.', 'info');
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

  const handleUpdateAdminPin = async (pin: string) => {
    try {
      await api.updateHouseholdSettings(activeHouseholdId, { admin_pin: pin });
      showToast('Admin security PIN updated successfully.', 'success');
      await refreshData();
    } catch (e: any) {
      showToast(e.message || 'Error updating Admin PIN', 'warning');
      throw e;
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

  // If verifying a magic link token from URL, show dedicated verification splash
  if (isVerifyingMagic) {
    return (
      <div id="magic-verifying-splash" className="min-h-screen bg-zinc-950 flex flex-col items-center justify-center p-4 text-center">
        <div className="bg-zinc-900 border border-zinc-800 rounded-2xl p-8 max-w-md w-full shadow-2xl space-y-4 animate-fadeIn">
          <div className="w-14 h-14 rounded-2xl bg-indigo-500/20 text-indigo-400 flex items-center justify-center mx-auto shadow-inner">
            <Sparkles className="w-7 h-7 animate-spin" />
          </div>
          <h2 className="text-xl font-bold text-white">Verifying Magic Link</h2>
          <p className="text-sm text-zinc-400">Authenticating your secure session. Loading your household...</p>
        </div>
      </div>
    );
  }

  // If not authenticated, render the dedicated Login / Onboarding screen
  if (!isAuthenticated) {
    return (
      <LoginView
        onLoginSuccess={handleLoginSuccess}
        onEnterKioskMode={async () => {
          handleToggleDeviceMode('tablet_kiosk');
          try {
            const res = await api.demoLogin(activeMemberId || 'm-sarah');
            handleLoginSuccess(res);
          } catch {
            setIsAuthenticated(true);
          }
        }}
        availableHouseholds={households}
      />
    );
  }

  return (
    <div className="min-h-screen bg-zinc-100/70 text-zinc-900 flex flex-col font-sans selection:bg-zinc-900 selection:text-white pb-16 md:pb-0">
      {/* Top Navbar */}
      <Navbar
        activeHousehold={activeHousehold}
        onOpenHouseholdInfo={() => setIsHouseholdInfoModalOpen(true)}
        members={members}
        activeMember={activeMember}
        onSignOut={handleSignOut}
        onOpenCreateChore={() => {
          setEditingChore(null);
          setIsCreateModalOpen(true);
        }}
        onOpenActivities={() => setCurrentTab('activity')}
        unreadActivitiesCount={activities.filter((a) => !a.read).length}
        pendingApprovalsCount={pendingApprovals.length}
        onModeChange={handleUpdateHouseholdMode}
        deviceMode={deviceMode}
        onToggleDeviceMode={handleToggleDeviceMode}
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
              {/* KITCHEN TABLET KIOSK TOUCH BAR */}
              {deviceMode === 'tablet_kiosk' && (
                <div
                  id="kiosk-touch-bar"
                  className="bg-linear-to-r from-indigo-900 via-zinc-900 to-indigo-950 rounded-2xl text-white p-4 sm:p-5 shadow-lg border border-indigo-800/50"
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
                    <div className="flex items-center gap-2.5">
                      <div className="w-9 h-9 rounded-xl bg-white/10 flex items-center justify-center border border-white/15">
                        <Tablet className="w-5 h-5 text-indigo-300" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h2 className="text-sm sm:text-base font-black tracking-tight text-white">
                            Kitchen Tablet Kiosk
                          </h2>
                          <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-400/30">
                            Shared Touch Screen
                          </span>
                        </div>
                        <p className="text-xs text-indigo-200/80 mt-0.5">
                          Tap your avatar below to switch user, complete tasks, or claim open chores.
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-2 self-start sm:self-auto">
                      <button
                        type="button"
                        id="kiosk-admin-pin-btn"
                        onClick={() => setIsPinModalOpen(true)}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-white/10 hover:bg-white/20 text-xs font-semibold text-white transition-colors cursor-pointer border border-white/15"
                        title="Open Admin PIN Verification (Default PIN: 1234)"
                      >
                        <Lock className="w-3.5 h-3.5 text-amber-300" />
                        <span>Admin PIN</span>
                      </button>
                      <button
                        type="button"
                        id="kiosk-switch-to-personal-btn"
                        onClick={() => handleToggleDeviceMode('personal')}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-xs font-semibold text-white transition-colors cursor-pointer shadow-xs"
                      >
                        <Smartphone className="w-3.5 h-3.5" />
                        <span className="hidden sm:inline">Switch to</span> Personal
                      </button>
                    </div>
                  </div>

                  {/* Quick-Tap Member Cards Grid */}
                  <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2.5">
                    {members.map((m) => {
                      const isCurrent = m.id === activeMemberId;
                      const memberChoresCount = chores.filter(
                        (c) => c.current_assignee_id === m.id && c.status !== 'completed'
                      ).length;

                      return (
                        <button
                          key={m.id}
                          type="button"
                          id={`kiosk-member-${m.id}`}
                          onClick={() => handleSelectMember(m.id)}
                          className={`flex items-center gap-3 p-3 rounded-xl text-left transition-all cursor-pointer border ${
                            isCurrent
                              ? 'bg-white text-zinc-900 border-white ring-2 ring-indigo-400 shadow-md scale-[1.02]'
                              : 'bg-white/10 hover:bg-white/15 text-white border-white/10 hover:border-white/20'
                          }`}
                        >
                          <div className="relative shrink-0">
                            <img
                              src={m.avatar_url}
                              alt={m.name}
                              className={`w-10 h-10 rounded-full object-cover border-2 ${
                                isCurrent ? 'border-indigo-600' : 'border-white/30'
                              }`}
                            />
                            {isCurrent && (
                              <span className="absolute -bottom-1 -right-1 w-4 h-4 bg-emerald-500 text-white rounded-full flex items-center justify-center text-[10px] font-bold ring-2 ring-white">
                                ✓
                              </span>
                            )}
                          </div>

                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-1.5">
                              <span
                                className={`font-bold text-xs truncate ${
                                  isCurrent ? 'text-zinc-900' : 'text-white'
                                }`}
                              >
                                {m.name}
                              </span>
                              {m.role === 'admin' && (
                                <span
                                  className={`text-[9px] font-black uppercase px-1 rounded ${
                                    isCurrent ? 'bg-indigo-100 text-indigo-700' : 'bg-white/20 text-indigo-200'
                                  }`}
                                  title="Admin"
                                >
                                  ★
                                </span>
                              )}
                            </div>

                            <div
                              className={`flex items-center gap-2 text-[11px] mt-0.5 ${
                                isCurrent ? 'text-zinc-500' : 'text-indigo-200/80'
                              }`}
                            >
                              <span>{m.points_balance} points</span>
                              <span>•</span>
                              <span>{memberChoresCount} due</span>
                            </div>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* PERSONAL DEVICE WELCOME BANNER */}
              {deviceMode === 'personal' && activeMember && (
                <div
                  id="personal-welcome-banner"
                  className="bg-white rounded-2xl border border-zinc-200 p-5 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4"
                >
                  <div className="flex items-center gap-3.5">
                    <div className="relative shrink-0">
                      <img
                        src={activeMember.avatar_url}
                        alt={activeMember.name}
                        className="w-12 h-12 rounded-2xl object-cover border-2 border-emerald-500 shadow-xs"
                      />
                      <div className="absolute -bottom-1 -right-1 w-5 h-5 rounded-lg bg-emerald-600 text-white flex items-center justify-center text-[10px] font-bold shadow-xs">
                        👤
                      </div>
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h2 className="text-base font-black text-zinc-900 tracking-tight">
                          Hi, {activeMember.name}! 👋
                        </h2>
                        <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
                          Personal Device Mode
                        </span>
                      </div>
                      <p className="text-xs text-zinc-500 mt-0.5">
                        {myAssignedChores.length > 0 ? (
                          <span>
                            You have <strong className="text-zinc-800">{myAssignedChores.length} active tasks</strong> assigned to you today.
                          </span>
                        ) : (
                          <span className="text-emerald-600 font-medium">
                            You are all caught up! No active chores assigned right now.
                          </span>
                        )}
                        {' '}
                        <span className="hidden sm:inline text-zinc-400">
                          (Board defaults to your tasks. Switch to Kitchen Tablet for shared view.)
                        </span>
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center gap-2">
                    {/* Quick View Filter Switchers */}
                    <button
                      type="button"
                      id="personal-filter-my"
                      onClick={() => setActiveViewFilter('my')}
                      className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                        activeViewFilter === 'my'
                          ? 'bg-zinc-900 text-white shadow-xs'
                          : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                      }`}
                    >
                      My Chores ({myAssignedChores.length})
                    </button>

                    <button
                      type="button"
                      id="personal-filter-all"
                      onClick={() => setActiveViewFilter('all')}
                      className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                        activeViewFilter === 'all'
                          ? 'bg-zinc-900 text-white shadow-xs'
                          : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
                      }`}
                    >
                      Whole House ({chores.length})
                    </button>

                    <button
                      type="button"
                      onClick={() => setCurrentTab('rewards')}
                      className="px-3 py-1.5 rounded-xl text-xs font-bold bg-amber-50 hover:bg-amber-100/80 text-amber-900 border border-amber-200 transition-colors cursor-pointer flex items-center gap-1.5"
                    >
                      <Coins className="w-3.5 h-3.5 text-amber-600" />
                      <span>{activeMember.points_balance} points</span>
                    </button>

                    <button
                      type="button"
                      id="personal-switch-to-kiosk-btn"
                      onClick={() => handleToggleDeviceMode('tablet_kiosk')}
                      className="px-3 py-1.5 rounded-xl text-xs font-semibold bg-zinc-100 hover:bg-zinc-200 text-zinc-700 transition-colors cursor-pointer flex items-center gap-1"
                      title="Switch to Kitchen Tablet Mode"
                    >
                      <Tablet className="w-3.5 h-3.5 text-indigo-600" />
                      <span className="hidden sm:inline">Tablet Mode</span>
                    </button>
                  </div>
                </div>
              )}

              {/* Dashboard Header & Search Controls */}
              <div className="bg-white rounded-2xl border border-zinc-200 p-5 space-y-4 shadow-xs">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div>
                    <h1 className="text-xl font-black text-zinc-900 tracking-tight flex items-center gap-2">
                      <span>Household Chores Board</span>
                      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-bold bg-indigo-100 text-indigo-700 border border-indigo-200">
                        v1.1.0 - Kind Rollout Verified
                      </span>
                    </h1>
                    <p className="text-xs text-zinc-500 mt-0.5">
                      {activeTasksCount} active tasks remaining • {completedCount} verified completed
                    </p>
                  </div>

                  {/* Search bar */}
                  <div className="relative w-full sm:w-72 lg:w-80">
                    <Search className="w-4 h-4 text-zinc-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                    <input
                      id="search-chores-input"
                      type="text"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder="Search chores or instructions..."
                      aria-label="Search chores or instructions"
                      className="w-full text-xs font-medium pl-9 pr-8 py-2 rounded-xl bg-zinc-50 border border-zinc-200 focus:outline-none focus:ring-2 focus:ring-zinc-900/10 focus:bg-white transition-all shadow-2xs"
                    />
                    {searchQuery && (
                      <button
                        id="clear-search-button"
                        type="button"
                        onClick={() => setSearchQuery('')}
                        className="absolute right-2.5 top-1/2 -translate-y-1/2 p-0.5 text-zinc-400 hover:text-zinc-700 rounded-full hover:bg-zinc-200/60 transition-colors cursor-pointer"
                        aria-label="Clear search query"
                        title="Clear search"
                      >
                        <X className="w-3.5 h-3.5" />
                      </button>
                    )}
                  </div>
                </div>

                {/* Filter Pills with Left/Right Arrow Navigation */}
                <div className="relative flex items-center gap-1.5">
                  <button
                    type="button"
                    id="filter-scroll-left"
                    onClick={() => scrollFilters('left')}
                    disabled={!canScrollFiltersLeft}
                    className="p-1.5 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-600 hover:text-zinc-900 disabled:opacity-25 disabled:pointer-events-none transition-all shrink-0 cursor-pointer shadow-2xs"
                    aria-label="Previous filters"
                    title="Previous"
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </button>

                  <div
                    ref={filterScrollRef}
                    className="flex-1 flex items-center gap-1.5 overflow-x-auto no-scrollbar scroll-smooth pb-1 sm:pb-0"
                  >
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

                  <button
                    type="button"
                    id="filter-scroll-right"
                    onClick={() => scrollFilters('right')}
                    disabled={!canScrollFiltersRight}
                    className="p-1.5 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-600 hover:text-zinc-900 disabled:opacity-25 disabled:pointer-events-none transition-all shrink-0 cursor-pointer shadow-2xs"
                    aria-label="Next filters"
                    title="Next"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>

                {/* Category Secondary Filter with Left/Right Arrow Navigation */}
                <div className="relative flex items-center gap-1.5 pt-2 border-t border-zinc-100">
                  <button
                    type="button"
                    id="category-scroll-left"
                    onClick={() => scrollCategories('left')}
                    disabled={!canScrollCategoriesLeft}
                    className="p-1 rounded-lg bg-zinc-100 hover:bg-zinc-200 text-zinc-500 hover:text-zinc-800 disabled:opacity-25 disabled:pointer-events-none transition-all shrink-0 cursor-pointer"
                    aria-label="Previous categories"
                    title="Previous"
                  >
                    <ChevronLeft className="w-3.5 h-3.5" />
                  </button>

                  <span className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider shrink-0 mr-1">
                    Category:
                  </span>

                  <div
                    ref={categoryScrollRef}
                    className="flex-1 flex items-center gap-1.5 overflow-x-auto no-scrollbar scroll-smooth text-xs text-zinc-600"
                  >
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

                  <button
                    type="button"
                    id="category-scroll-right"
                    onClick={() => scrollCategories('right')}
                    disabled={!canScrollCategoriesRight}
                    className="p-1 rounded-lg bg-zinc-100 hover:bg-zinc-200 text-zinc-500 hover:text-zinc-800 disabled:opacity-25 disabled:pointer-events-none transition-all shrink-0 cursor-pointer"
                    aria-label="Next categories"
                    title="Next"
                  >
                    <ChevronRight className="w-3.5 h-3.5" />
                  </button>
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
              currentMember={activeMember}
              onUpdateMode={handleUpdateHouseholdMode}
              onUpdateAdminPin={handleUpdateAdminPin}
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

      {/* Invite Link Onboarding Modal */}
      {inviteHousehold && (
        <InviteOnboardingModal
          isOpen={isInviteModalOpen}
          household={inviteHousehold}
          onJoin={handleJoinFromInvite}
          onClose={() => setIsInviteModalOpen(false)}
        />
      )}

      {/* Admin PIN Verification Modal for Kiosk */}
      <PinVerificationModal
        householdId={activeHouseholdId}
        isOpen={isPinModalOpen}
        onSuccess={() => {
          handleSwitchToAdmin();
          showToast('Admin access verified via PIN! Switched to Admin role.', 'success');
        }}
        onClose={() => setIsPinModalOpen(false)}
      />

      {/* Household Details & Invite Modal */}
      <HouseholdInfoModal
        isOpen={isHouseholdInfoModalOpen}
        onClose={() => setIsHouseholdInfoModalOpen(false)}
        household={activeHousehold}
        members={members}
        currentMember={activeMember}
        onSignOut={handleSignOut}
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
