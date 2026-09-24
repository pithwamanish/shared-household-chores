import {
  Household,
  Member,
  Chore,
  ChoreCompletion,
  ChoreSwapRequest,
  RewardItem,
  RewardRedemption,
  ActivityLog,
  ChoreComment,
  ChoreFilterOptions,
  HouseholdSettings,
  HouseholdMode,
  AuthTokenResponse,
  ForgotPasswordResponse,
  ChoreNudgeResponse,
  DevEmail,
} from '../types';
import { httpClient, isNetworkError, ApiError } from './httpClient';

const STORAGE_KEY = 'choresync_storage_v1';

// Seed initial data
const INITIAL_HOUSEHOLDS: Household[] = [
  {
    id: 'h-roommates',
    name: 'Apartment 4B',
    invite_code: 'APT4B-SHARE',
    settings: {
      default_mode: 'flatmate',
      allow_swaps: true,
      timezone: 'America/New_York',
    },
    created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'h-family',
    name: 'The Miller Family',
    invite_code: 'MILLER-HOME',
    settings: {
      default_mode: 'family',
      allow_swaps: true,
      timezone: 'America/New_York',
    },
    created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'h-couple',
    name: 'Alex & Sam',
    invite_code: 'COUPLE-NEST',
    settings: {
      default_mode: 'casual',
      allow_swaps: false,
      timezone: 'America/New_York',
    },
    created_at: new Date(Date.now() - 15 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const INITIAL_MEMBERS: Member[] = [
  // Roommates
  {
    id: 'm-sarah',
    household_id: 'h-roommates',
    name: 'Sarah Chen',
    email: 'sarah@example.com',
    avatar_url: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=120&auto=format&fit=crop&q=80',
    role: 'admin',
    points_balance: 145,
    total_points_earned: 420,
    streak: 6,
    created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-liam',
    household_id: 'h-roommates',
    name: 'Liam Vance',
    email: 'liam@example.com',
    avatar_url: 'https://images.unsplash.com/photo-1539571696357-5a69c17a67c6?w=120&auto=format&fit=crop&q=80',
    role: 'member',
    points_balance: 85,
    total_points_earned: 310,
    streak: 3,
    created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-maya',
    household_id: 'h-roommates',
    name: 'Maya Patel',
    email: 'maya@example.com',
    avatar_url: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?w=120&auto=format&fit=crop&q=80',
    role: 'member',
    points_balance: 190,
    total_points_earned: 490,
    streak: 9,
    created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-noah',
    household_id: 'h-roommates',
    name: 'Noah Brooks',
    email: 'noah@example.com',
    avatar_url: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=120&auto=format&fit=crop&q=80',
    role: 'member',
    points_balance: 60,
    total_points_earned: 220,
    streak: 1,
    created_at: new Date(Date.now() - 20 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },

  // Family
  {
    id: 'm-david-fam',
    household_id: 'h-family',
    name: 'David (Dad)',
    email: 'david@miller.family',
    avatar_url: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=120&auto=format&fit=crop&q=80',
    role: 'admin',
    points_balance: 320,
    total_points_earned: 890,
    streak: 14,
    created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-sarah-fam',
    household_id: 'h-family',
    name: 'Sarah (Mom)',
    email: 'sarah@miller.family',
    avatar_url: 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=120&auto=format&fit=crop&q=80',
    role: 'admin',
    points_balance: 290,
    total_points_earned: 760,
    streak: 12,
    created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-leo-kid',
    household_id: 'h-family',
    name: 'Leo (10y)',
    email: 'leo@miller.family',
    avatar_url: 'https://images.unsplash.com/photo-1485546246426-74dc88dec4d9?w=120&auto=format&fit=crop&q=80',
    role: 'child',
    points_balance: 75,
    total_points_earned: 230,
    streak: 5,
    created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-emma-kid',
    household_id: 'h-family',
    name: 'Emma (8y)',
    email: 'emma@miller.family',
    avatar_url: 'https://images.unsplash.com/photo-1519456264917-42d0aa2e0625?w=120&auto=format&fit=crop&q=80',
    role: 'child',
    points_balance: 110,
    total_points_earned: 280,
    streak: 8,
    created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },

  // Couple
  {
    id: 'm-alex',
    household_id: 'h-couple',
    name: 'Alex Rivera',
    email: 'alex@example.com',
    avatar_url: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=120&auto=format&fit=crop&q=80',
    role: 'member',
    points_balance: 150,
    total_points_earned: 450,
    streak: 5,
    created_at: new Date(Date.now() - 15 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'm-sam',
    household_id: 'h-couple',
    name: 'Sam Taylor',
    email: 'sam@example.com',
    avatar_url: 'https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?w=120&auto=format&fit=crop&q=80',
    role: 'member',
    points_balance: 180,
    total_points_earned: 490,
    streak: 7,
    created_at: new Date(Date.now() - 15 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const INITIAL_CHORES: Chore[] = [
  // Roommates chores
  {
    id: 'c-recycling-rr',
    household_id: 'h-roommates',
    title: 'Recycling & Garbage to Curb',
    description: 'Separate glass, plastic, and cardboard. Roll bins to curb by 8 PM Tuesday.',
    category: 'maintenance',
    effort_points: 20,
    assignment_type: 'round_robin',
    current_assignee_id: 'm-liam',
    rotation_member_ids: ['m-sarah', 'm-liam', 'm-maya', 'm-noah'],
    current_rotation_index: 1,
    recurrence_type: 'weekly',
    recurrence_rule: { days: ['tue'] },
    due_date: new Date(Date.now() + 18 * 3600000).toISOString(), // upcoming today
    requires_approval: false,
    requires_proof: false,
    status: 'assigned',
    created_by: 'm-sarah',
    created_at: new Date(Date.now() - 7 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-kitchen-deep',
    household_id: 'h-roommates',
    title: 'Deep Clean Kitchen Counters & Stovetop',
    description: 'Wipe all counters with disinfectant, degrease stovetop burners, polish sink.',
    category: 'kitchen',
    effort_points: 25,
    assignment_type: 'direct',
    current_assignee_id: 'm-sarah',
    recurrence_type: 'weekly',
    due_date: new Date(Date.now() + 2 * 86400000).toISOString(),
    requires_approval: false,
    requires_proof: false,
    status: 'assigned',
    created_by: 'm-sarah',
    created_at: new Date(Date.now() - 5 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-vacuum-living',
    household_id: 'h-roommates',
    title: 'Vacuum Living Room & Dust Shelves',
    description: 'Move chairs, vacuum under sofa edge, dust TV console and bookshelves.',
    category: 'cleaning',
    effort_points: 15,
    assignment_type: 'direct',
    current_assignee_id: 'm-liam',
    recurrence_type: 'weekly',
    due_date: new Date(Date.now() - 14 * 3600000).toISOString(), // Overdue!
    requires_approval: false,
    requires_proof: false,
    status: 'overdue',
    created_by: 'm-sarah',
    created_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-bathroom-mop',
    household_id: 'h-roommates',
    title: 'Scrub Shower & Mop Bathroom Floor',
    description: 'Scrub tiles, empty shower drain, disinfect toilet, and mop floor.',
    category: 'cleaning',
    effort_points: 30,
    assignment_type: 'round_robin',
    current_assignee_id: 'm-maya',
    rotation_member_ids: ['m-maya', 'm-noah', 'm-sarah', 'm-liam'],
    current_rotation_index: 0,
    recurrence_type: 'weekly',
    due_date: new Date(Date.now() + 4 * 86400000).toISOString(),
    requires_approval: false,
    requires_proof: false,
    status: 'assigned',
    created_by: 'm-sarah',
    created_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-balcony-sweep',
    household_id: 'h-roommates',
    title: 'Water Plants & Sweep Balcony',
    description: 'Water fern and herb garden, sweep fallen leaves and tidy outdoor chairs.',
    category: 'yard',
    effort_points: 10,
    assignment_type: 'open_pool',
    current_assignee_id: null,
    recurrence_type: 'daily',
    due_date: new Date(Date.now() + 6 * 3600000).toISOString(),
    requires_approval: false,
    requires_proof: false,
    status: 'unassigned',
    created_by: 'm-maya',
    created_at: new Date(Date.now() - 1 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },

  // Family chores
  {
    id: 'c-fam-bedroom-leo',
    household_id: 'h-family',
    title: 'Clean Up Bedroom & Put Away Toys',
    description: 'Make bed, put Lego boxes back in closet, hang up clean clothes.',
    category: 'cleaning',
    effort_points: 15,
    assignment_type: 'direct',
    current_assignee_id: 'm-leo-kid',
    recurrence_type: 'daily',
    due_date: new Date(Date.now() - 2 * 3600000).toISOString(),
    requires_approval: true,
    requires_proof: true,
    status: 'pending_approval',
    created_by: 'm-sarah-fam',
    created_at: new Date(Date.now() - 1 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-fam-dog-walk',
    household_id: 'h-family',
    title: 'Walk Buster & Refill Dog Bowls',
    description: '20-minute walk around the park. Wash food bowl and add clean cold water.',
    category: 'pets',
    effort_points: 20,
    assignment_type: 'round_robin',
    current_assignee_id: 'm-emma-kid',
    rotation_member_ids: ['m-leo-kid', 'm-emma-kid'],
    current_rotation_index: 1,
    recurrence_type: 'daily',
    due_date: new Date(Date.now() + 5 * 3600000).toISOString(),
    requires_approval: true,
    requires_proof: true,
    status: 'assigned',
    created_by: 'm-david-fam',
    created_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-fam-lawn',
    household_id: 'h-family',
    title: 'Mow Front Lawn & Rake Grass',
    description: 'Use the electric mower on level 3, edge sidewalk, and put clippings in compost bin.',
    category: 'yard',
    effort_points: 35,
    assignment_type: 'direct',
    current_assignee_id: 'm-david-fam',
    recurrence_type: 'weekly',
    due_date: new Date(Date.now() + 3 * 86400000).toISOString(),
    requires_approval: false,
    requires_proof: false,
    status: 'assigned',
    created_by: 'm-sarah-fam',
    created_at: new Date(Date.now() - 2 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-fam-dishwasher',
    household_id: 'h-family',
    title: 'Unload Clean Dishwasher',
    description: 'Put plates, bowls, cups, and cutlery into their respective cupboards.',
    category: 'kitchen',
    effort_points: 10,
    assignment_type: 'open_pool',
    current_assignee_id: null,
    recurrence_type: 'daily',
    due_date: new Date(Date.now() + 8 * 3600000).toISOString(),
    requires_approval: true,
    requires_proof: false,
    status: 'unassigned',
    created_by: 'm-sarah-fam',
    created_at: new Date(Date.now() - 1 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },

  // Couple chores
  {
    id: 'c-couple-groceries',
    household_id: 'h-couple',
    title: 'Weekend Farmers Market Run',
    description: 'Pick up eggs, sourdough bread, organic greens, and oat milk.',
    category: 'other',
    effort_points: 20,
    assignment_type: 'open_pool',
    current_assignee_id: null,
    recurrence_type: 'weekly',
    due_date: new Date(Date.now() + 24 * 3600000).toISOString(),
    requires_approval: false,
    requires_proof: false,
    status: 'unassigned',
    created_by: 'm-alex',
    created_at: new Date(Date.now() - 1 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: 'c-couple-sheets',
    household_id: 'h-couple',
    title: 'Wash & Change Bed Linens',
    description: 'Fitted sheet, duvet cover, and pillowcases. Wash on warm, hang to dry.',
    category: 'cleaning',
    effort_points: 15,
    assignment_type: 'direct',
    current_assignee_id: 'm-sam',
    recurrence_type: 'weekly',
    due_date: new Date(Date.now() + 12 * 3600000).toISOString(),
    requires_approval: false,
    requires_proof: false,
    status: 'assigned',
    created_by: 'm-alex',
    created_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const INITIAL_COMPLETIONS: ChoreCompletion[] = [
  {
    id: 'comp-seed-leo-1',
    chore_id: 'c-fam-bedroom-leo',
    household_id: 'h-family',
    member_id: 'm-leo-kid',
    status: 'pending_approval',
    proof_notes: 'I made the bed neat, organized my Lego sets, and put all the shoes away!',
    proof_photo_url: 'https://images.unsplash.com/photo-1513694203232-719a280e022f?w=600&auto=format&fit=crop&q=80',
    points_awarded: 15,
    completed_at: new Date(Date.now() - 45 * 60000).toISOString(),
  },
];

const INITIAL_SWAPS: ChoreSwapRequest[] = [
  {
    id: 'swap-seed-1',
    household_id: 'h-roommates',
    chore_id: 'c-vacuum-living',
    requester_id: 'm-liam',
    target_member_id: 'm-maya',
    status: 'pending',
    reason: 'I have a big mid-term exam tomorrow night; could anyone take this? I will cover recycling next week!',
    created_at: new Date(Date.now() - 5 * 3600000).toISOString(),
  },
];

const INITIAL_REWARDS: RewardItem[] = [
  // Roommates
  {
    id: 'rew-room-coffee',
    household_id: 'h-roommates',
    title: 'Roommate Buys Your Fancy Coffee',
    description: 'Next coffee run is on the house! Order whatever oat latte or pastry you like.',
    points_cost: 60,
    icon: 'Coffee',
    is_active: true,
  },
  {
    id: 'rew-room-dishes-pass',
    household_id: 'h-roommates',
    title: 'Exemption from Dishes for a Day',
    description: 'Get a full 24-hour pass where your housemates clean up all cooking dishes.',
    points_cost: 90,
    icon: 'Sparkles',
    is_active: true,
  },
  {
    id: 'rew-room-movie',
    household_id: 'h-roommates',
    title: 'Sole Pick for Apartment Movie Night',
    description: 'Zero debates or vetoes: you pick the movie and Friday snack.',
    points_cost: 50,
    icon: 'Tv',
    is_active: true,
  },

  // Family
  {
    id: 'rew-fam-bedtime',
    household_id: 'h-family',
    title: '30-Minute Late Bedtime Pass',
    description: 'Stay up an extra 30 minutes reading or watching your favorite cartoon.',
    points_cost: 40,
    icon: 'Moon',
    is_active: true,
  },
  {
    id: 'rew-fam-icecream',
    household_id: 'h-family',
    title: 'Ice Cream Sundae Trip',
    description: 'Trip to the local creamery with double scoops and two toppings of your choice!',
    points_cost: 75,
    icon: 'IceCream',
    is_active: true,
  },
  {
    id: 'rew-fam-game',
    household_id: 'h-family',
    title: 'Choose Sunday Family Board Game',
    description: 'Pick the game we all play together after Sunday dinner.',
    points_cost: 30,
    icon: 'Gamepad2',
    is_active: true,
  },

  // Couple
  {
    id: 'rew-couple-dinner',
    household_id: 'h-couple',
    title: 'Partner Cooks Your Favorite Dinner',
    description: 'Sit back with a beverage while your partner cooks any meal you crave.',
    points_cost: 80,
    icon: 'UtensilsCrossed',
    is_active: true,
  },
  {
    id: 'rew-couple-massage',
    household_id: 'h-couple',
    title: '20-Minute Shoulder & Back Massage',
    description: 'Relaxing massage with soothing music and essential oils.',
    points_cost: 70,
    icon: 'HeartHandshake',
    is_active: true,
  },
];

const INITIAL_REDEMPTIONS: RewardRedemption[] = [
  {
    id: 'red-seed-1',
    household_id: 'h-family',
    member_id: 'm-emma-kid',
    reward_item_id: 'rew-fam-icecream',
    points_spent: 75,
    status: 'requested',
    requested_at: new Date(Date.now() - 3 * 3600000).toISOString(),
  },
];

const INITIAL_ACTIVITIES: ActivityLog[] = [
  {
    id: 'act-1',
    household_id: 'h-roommates',
    actor_id: 'm-sarah',
    event_type: 'chore_created',
    entity_type: 'chore',
    entity_id: 'c-recycling-rr',
    message: 'Sarah created weekly round-robin chore "Recycling & Garbage to Curb"',
    read: true,
    created_at: new Date(Date.now() - 24 * 3600000).toISOString(),
  },
  {
    id: 'act-2',
    household_id: 'h-roommates',
    actor_id: 'm-liam',
    event_type: 'swap_proposed',
    entity_type: 'swap',
    entity_id: 'swap-seed-1',
    message: 'Liam requested to swap "Vacuum Living Room" with Maya',
    read: false,
    created_at: new Date(Date.now() - 5 * 3600000).toISOString(),
  },
  {
    id: 'act-3',
    household_id: 'h-family',
    actor_id: 'm-leo-kid',
    event_type: 'approval_requested',
    entity_type: 'chore',
    entity_id: 'c-fam-bedroom-leo',
    message: 'Leo submitted "Clean Up Bedroom & Put Away Toys" for Parent verification',
    read: false,
    created_at: new Date(Date.now() - 45 * 60000).toISOString(),
  },
  {
    id: 'act-4',
    household_id: 'h-family',
    actor_id: 'm-emma-kid',
    event_type: 'reward_redeemed',
    entity_type: 'reward',
    entity_id: 'rew-fam-icecream',
    message: 'Emma redeemed 75 pts for "Ice Cream Sundae Trip"',
    read: false,
    created_at: new Date(Date.now() - 3 * 3600000).toISOString(),
  },
];

const INITIAL_COMMENTS: ChoreComment[] = [
  {
    id: 'comm-1',
    chore_id: 'c-vacuum-living',
    member_id: 'm-liam',
    message: 'Will definitely vacuum under the sofa cushions this time as requested!',
    created_at: new Date(Date.now() - 2 * 3600000).toISOString(),
  },
];

interface ChoreSyncStore {
  households: Household[];
  members: Member[];
  chores: Chore[];
  completions: ChoreCompletion[];
  swaps: ChoreSwapRequest[];
  rewards: RewardItem[];
  redemptions: RewardRedemption[];
  activities: ActivityLog[];
  comments: ChoreComment[];
}

function loadStore(): ChoreSyncStore {
  if (typeof window === 'undefined') {
    return {
      households: INITIAL_HOUSEHOLDS,
      members: INITIAL_MEMBERS,
      chores: INITIAL_CHORES,
      completions: INITIAL_COMPLETIONS,
      swaps: INITIAL_SWAPS,
      rewards: INITIAL_REWARDS,
      redemptions: INITIAL_REDEMPTIONS,
      activities: INITIAL_ACTIVITIES,
      comments: INITIAL_COMMENTS,
    };
  }
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      if (parsed && Array.isArray(parsed.households) && parsed.households.length > 0) {
        return parsed;
      }
    }
  } catch (e) {
    console.warn('Failed to parse ChoreSync store from localStorage, initializing default.', e);
  }

  const initialStore: ChoreSyncStore = {
    households: INITIAL_HOUSEHOLDS,
    members: INITIAL_MEMBERS,
    chores: INITIAL_CHORES,
    completions: INITIAL_COMPLETIONS,
    swaps: INITIAL_SWAPS,
    rewards: INITIAL_REWARDS,
    redemptions: INITIAL_REDEMPTIONS,
    activities: INITIAL_ACTIVITIES,
    comments: INITIAL_COMMENTS,
  };
  saveStore(initialStore);
  return initialStore;
}

function saveStore(store: ChoreSyncStore): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(store));
  } catch (e) {
    console.error('Failed to save ChoreSync store to localStorage', e);
  }
}

// Simulated network delay helper
function delay<T>(value: T, ms = 80): Promise<T> {
  return new Promise((resolve) => setTimeout(() => resolve(value), ms));
}

// Generate unique ID helper
function generateId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`;
}

export const localApi = {
  // Household operations
  async getHouseholds(): Promise<Household[]> {
    const store = loadStore();
    return delay([...store.households]);
  },

  async createHousehold(payload: {
    name: string;
    default_mode?: HouseholdMode;
    allow_swaps?: boolean;
    timezone?: string;
  }): Promise<Household> {
    const store = loadStore();
    const newHousehold: Household = {
      id: generateId('h'),
      name: payload.name.trim(),
      invite_code: `${payload.name.toUpperCase().replace(/[^A-Z0-9]/g, '').substring(0, 5) || 'HOUSE'}-${Math.random().toString(36).substring(2, 6).toUpperCase()}`,
      settings: {
        default_mode: payload.default_mode || 'flatmate',
        allow_swaps: payload.allow_swaps ?? true,
        timezone: payload.timezone || 'America/New_York',
      },
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    store.households.push(newHousehold);
    saveStore(store);
    return delay({ ...newHousehold });
  },

  async joinHousehold(inviteCode: string): Promise<Household> {
    const store = loadStore();
    const code = inviteCode.trim().toUpperCase();
    const found = store.households.find((h) => h.invite_code.toUpperCase() === code);
    if (!found) throw new Error(`Household with invite code "${inviteCode}" not found`);
    return delay({ ...found });
  },

  async getHousehold(id: string): Promise<Household | null> {
    const store = loadStore();
    const h = store.households.find((x) => x.id === id) || null;
    return delay(h ? { ...h } : null);
  },

  async updateHouseholdSettings(id: string, settings: Partial<HouseholdSettings>): Promise<Household> {
    const store = loadStore();
    const index = store.households.findIndex((x) => x.id === id);
    if (index === -1) throw new Error('Household not found');
    store.households[index] = {
      ...store.households[index],
      settings: { ...store.households[index].settings, ...settings },
      updated_at: new Date().toISOString(),
    };
    saveStore(store);
    return delay({ ...store.households[index] });
  },

  // Member operations
  async getMembers(householdId: string): Promise<Member[]> {
    const store = loadStore();
    const members = store.members.filter((m) => m.household_id === householdId);
    return delay([...members]);
  },

  async addMember(
    householdId: string,
    payload: { name: string; role: Member['role']; avatar_url?: string }
  ): Promise<Member> {
    const store = loadStore();
    const newMember: Member = {
      id: generateId('m'),
      household_id: householdId,
      name: payload.name.trim(),
      role: payload.role,
      avatar_url: payload.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${encodeURIComponent(payload.name)}`,
      points_balance: 0,
      total_points_earned: 0,
      streak: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    store.members.push(newMember);
    saveStore(store);
    return delay({ ...newMember });
  },

  async login(email: string, _password: string): Promise<AuthTokenResponse> {
    const store = loadStore();
    const cleanEmail = email.trim().toLowerCase();
    const member = store.members.find((m) => (m.email || '').toLowerCase() === cleanEmail);
    if (!member) throw new Error('Invalid email or password');
    const household = store.households.find((h) => h.id === member.household_id);
    if (!household) throw new Error('Household not found');
    const token = `local-token-${member.id}`;
    httpClient.setAuthToken(token);
    return delay({
      access_token: token,
      token_type: 'Bearer',
      expires_in: 86400,
      member: { ...member },
      household: { ...household },
    });
  },

  async register(payload: {
    name: string;
    email: string;
    password: string;
    household_name?: string;
    invite_code?: string;
    role?: string;
  }): Promise<AuthTokenResponse> {
    const store = loadStore();
    let house: Household | undefined;
    if (payload.invite_code) {
      const code = payload.invite_code.trim().toUpperCase();
      house = store.households.find((h) => h.invite_code.toUpperCase() === code);
      if (!house) throw new Error('Invalid invite code');
    } else {
      house = {
        id: generateId('h'),
        name: (payload.household_name || 'My Household').trim(),
        invite_code: `${(payload.household_name || 'HOUSE').toUpperCase().replace(/[^A-Z0-9]/g, '').substring(0, 5) || 'HOUSE'}-${Math.random().toString(36).substring(2, 6).toUpperCase()}`,
        settings: {
          default_mode: 'flatmate',
          allow_swaps: true,
          timezone: 'America/New_York',
        },
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      store.households.push(house);
    }

    const member: Member = {
      id: generateId('m'),
      household_id: house.id,
      name: payload.name.trim(),
      email: payload.email.trim(),
      role: (payload.role as any) || 'member',
      points_balance: 0,
      total_points_earned: 0,
      streak: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    store.members.push(member);
    saveStore(store);

    const token = `local-token-${member.id}`;
    httpClient.setAuthToken(token);
    return delay({
      access_token: token,
      token_type: 'Bearer',
      expires_in: 86400,
      member: { ...member },
      household: { ...house },
    });
  },

  // Chore operations
  async getChores(householdId: string, filterOptions?: ChoreFilterOptions): Promise<Chore[]> {
    const store = loadStore();
    let chores = store.chores.filter((c) => c.household_id === householdId);

    // Dynamic overdue status check based on current local time
    const now = new Date();
    chores = chores.map((c) => {
      if (c.due_date && c.status !== 'completed' && c.status !== 'pending_approval') {
        const isPastDue = new Date(c.due_date).getTime() < now.getTime();
        if (isPastDue && c.status !== 'overdue') {
          return { ...c, status: 'overdue' as const };
        }
      }
      return c;
    });

    if (filterOptions) {
      if (filterOptions.category) {
        chores = chores.filter((c) => c.category === filterOptions.category);
      }
      if (filterOptions.status) {
        chores = chores.filter((c) => c.status === filterOptions.status);
      }
      if (filterOptions.search) {
        const q = filterOptions.search.toLowerCase();
        chores = chores.filter(
          (c) =>
            c.title.toLowerCase().includes(q) ||
            (c.description && c.description.toLowerCase().includes(q))
        );
      }
      if (filterOptions.view === 'my' && filterOptions.assignee_id) {
        chores = chores.filter((c) => c.current_assignee_id === filterOptions.assignee_id);
      } else if (filterOptions.view === 'open') {
        chores = chores.filter((c) => c.assignment_type === 'open_pool' && !c.current_assignee_id);
      } else if (filterOptions.view === 'approval') {
        chores = chores.filter((c) => c.status === 'pending_approval');
      } else if (filterOptions.view === 'overdue') {
        chores = chores.filter((c) => c.status === 'overdue');
      }
    }

    return delay([...chores]);
  },

  async getChore(choreId: string): Promise<Chore | null> {
    const store = loadStore();
    const chore = store.chores.find((c) => c.id === choreId);
    return delay(chore ? { ...chore } : null);
  },

  async createChore(payload: Omit<Chore, 'id' | 'created_at' | 'updated_at' | 'status'> & { status?: Chore['status'] }): Promise<Chore> {
    const store = loadStore();
    const newChoreId = generateId('c');

    let initialStatus: Chore['status'] = payload.status || 'assigned';
    let currentAssigneeId = payload.current_assignee_id;

    if (payload.assignment_type === 'open_pool') {
      initialStatus = 'unassigned';
      currentAssigneeId = null;
    } else if (payload.assignment_type === 'round_robin') {
      const rotationList = payload.rotation_member_ids || [];
      const idx = payload.current_rotation_index || 0;
      currentAssigneeId = rotationList.length > 0 ? rotationList[idx % rotationList.length] : null;
      initialStatus = currentAssigneeId ? 'assigned' : 'unassigned';
    }

    const newChore: Chore = {
      ...payload,
      id: newChoreId,
      current_assignee_id: currentAssigneeId,
      status: initialStatus,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    store.chores.unshift(newChore);

    // Log activity
    store.activities.unshift({
      id: generateId('act'),
      household_id: payload.household_id,
      actor_id: payload.created_by,
      event_type: 'chore_created',
      entity_type: 'chore',
      entity_id: newChoreId,
      message: `Created new chore "${payload.title}" (${payload.effort_points} pts)`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay({ ...newChore });
  },

  async updateChore(id: string, payload: Partial<Chore>): Promise<Chore> {
    const store = loadStore();
    const index = store.chores.findIndex((c) => c.id === id);
    if (index === -1) throw new Error('Chore not found');

    const updated = {
      ...store.chores[index],
      ...payload,
      updated_at: new Date().toISOString(),
    };
    store.chores[index] = updated;
    saveStore(store);
    return delay({ ...updated });
  },

  async deleteChore(id: string): Promise<boolean> {
    const store = loadStore();
    const index = store.chores.findIndex((c) => c.id === id);
    if (index === -1) return delay(false);
    store.chores.splice(index, 1);
    saveStore(store);
    return delay(true);
  },

  // Claim chore from open pool
  async claimChore(choreId: string, memberId: string): Promise<Chore> {
    const store = loadStore();
    const chore = store.chores.find((c) => c.id === choreId);
    if (!chore) throw new Error('Chore not found');
    const member = store.members.find((m) => m.id === memberId);

    chore.current_assignee_id = memberId;
    chore.status = 'assigned';
    chore.updated_at = new Date().toISOString();

    store.activities.unshift({
      id: generateId('act'),
      household_id: chore.household_id,
      actor_id: memberId,
      event_type: 'chore_assigned',
      entity_type: 'chore',
      entity_id: chore.id,
      message: `${member?.name || 'A member'} claimed "${chore.title}" from the Open Pool`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay({ ...chore });
  },

  // Complete chore
  async completeChore(
    choreId: string,
    memberId: string,
    proofPayload?: { proof_notes?: string; proof_photo_url?: string }
  ): Promise<{ chore: Chore; completion?: ChoreCompletion; pointsAwarded: number }> {
    const store = loadStore();
    const chore = store.chores.find((c) => c.id === choreId);
    if (!chore) throw new Error('Chore not found');

    const member = store.members.find((m) => m.id === memberId);
    const memberName = member?.name || 'A member';

    if (chore.requires_approval) {
      // Moves to pending approval
      chore.status = 'pending_approval';
      chore.updated_at = new Date().toISOString();

      const newCompletion: ChoreCompletion = {
        id: generateId('comp'),
        chore_id: chore.id,
        household_id: chore.household_id,
        member_id: memberId,
        status: 'pending_approval',
        proof_notes: proofPayload?.proof_notes,
        proof_photo_url: proofPayload?.proof_photo_url,
        points_awarded: chore.effort_points,
        completed_at: new Date().toISOString(),
      };
      store.completions.unshift(newCompletion);

      store.activities.unshift({
        id: generateId('act'),
        household_id: chore.household_id,
        actor_id: memberId,
        event_type: 'approval_requested',
        entity_type: 'chore',
        entity_id: chore.id,
        message: `${memberName} submitted "${chore.title}" for verification`,
        read: false,
        created_at: new Date().toISOString(),
      });

      saveStore(store);
      return delay({ chore: { ...chore }, completion: newCompletion, pointsAwarded: 0 });
    } else {
      // Instant completion
      chore.status = 'completed';
      chore.updated_at = new Date().toISOString();

      if (member) {
        member.points_balance += chore.effort_points;
        member.total_points_earned += chore.effort_points;
        member.streak += 1;
        member.updated_at = new Date().toISOString();
      }

      const newCompletion: ChoreCompletion = {
        id: generateId('comp'),
        chore_id: chore.id,
        household_id: chore.household_id,
        member_id: memberId,
        status: 'approved',
        proof_notes: proofPayload?.proof_notes,
        proof_photo_url: proofPayload?.proof_photo_url,
        points_awarded: chore.effort_points,
        completed_at: new Date().toISOString(),
        verified_at: new Date().toISOString(),
      };
      store.completions.unshift(newCompletion);

      store.activities.unshift({
        id: generateId('act'),
        household_id: chore.household_id,
        actor_id: memberId,
        event_type: 'chore_completed',
        entity_type: 'chore',
        entity_id: chore.id,
        message: `${memberName} completed "${chore.title}" (+${chore.effort_points} pts)`,
        read: false,
        created_at: new Date().toISOString(),
      });

      // Advance recurrence if applicable
      advanceRecurrence(store, chore);

      saveStore(store);
      return delay({ chore: { ...chore }, completion: newCompletion, pointsAwarded: chore.effort_points });
    }
  },

  async uploadPhoto(file: File): Promise<{ success: boolean; url: string; s3_key: string }> {
    return new Promise((resolve) => {
      const reader = new FileReader();
      reader.onloadend = () => {
        resolve({
          success: true,
          url: reader.result as string,
          s3_key: `proofs/local-${Date.now()}-${file.name}`,
        });
      };
      reader.readAsDataURL(file);
    });
  },

  async rotateChore(choreId: string): Promise<Chore> {
    const store = loadStore();
    const chore = store.chores.find((c) => c.id === choreId);
    if (!chore) throw new Error('Chore not found');

    const rotationList = chore.rotation_member_ids || [];
    if (rotationList.length > 0) {
      const nextIdx = ((chore.current_rotation_index || 0) + 1) % rotationList.length;
      chore.current_rotation_index = nextIdx;
      chore.current_assignee_id = rotationList[nextIdx];
      chore.status = 'assigned';
      chore.updated_at = new Date().toISOString();
      saveStore(store);
    }
    return delay({ ...chore });
  },

  // Approvals
  async getPendingApprovals(householdId: string): Promise<
    Array<{ completion: ChoreCompletion; chore: Chore; member: Member }>
  > {
    const store = loadStore();
    const list: Array<{ completion: ChoreCompletion; chore: Chore; member: Member }> = [];

    for (const comp of store.completions) {
      if (comp.household_id === householdId && comp.status === 'pending_approval') {
        const chore = store.chores.find((c) => c.id === comp.chore_id);
        const member = store.members.find((m) => m.id === comp.member_id);
        if (chore && member) {
          list.push({ completion: { ...comp }, chore: { ...chore }, member: { ...member } });
        }
      }
    }
    return delay(list);
  },

  async approveChore(completionId: string, adminMemberId: string): Promise<boolean> {
    const store = loadStore();
    const completion = store.completions.find((c) => c.id === completionId);
    if (!completion) throw new Error('Completion record not found');

    const chore = store.chores.find((c) => c.id === completion.chore_id);
    const member = store.members.find((m) => m.id === completion.member_id);
    const admin = store.members.find((m) => m.id === adminMemberId);

    completion.status = 'approved';
    completion.approved_by = adminMemberId;
    completion.verified_at = new Date().toISOString();

    if (chore) {
      chore.status = 'completed';
      chore.updated_at = new Date().toISOString();
    }

    if (member && chore) {
      member.points_balance += chore.effort_points;
      member.total_points_earned += chore.effort_points;
      member.streak += 1;
      member.updated_at = new Date().toISOString();
    }

    store.activities.unshift({
      id: generateId('act'),
      household_id: completion.household_id,
      actor_id: adminMemberId,
      event_type: 'chore_approved',
      entity_type: 'chore',
      entity_id: completion.chore_id,
      message: `${admin?.name || 'Admin'} approved ${member?.name || 'Member'}'s "${chore?.title || 'chore'}" (+${chore?.effort_points || 0} pts)`,
      read: false,
      created_at: new Date().toISOString(),
    });

    if (chore) {
      advanceRecurrence(store, chore);
    }

    saveStore(store);
    return delay(true);
  },

  async rejectChore(completionId: string, adminMemberId: string, reason: string): Promise<boolean> {
    const store = loadStore();
    const completion = store.completions.find((c) => c.id === completionId);
    if (!completion) throw new Error('Completion record not found');

    const chore = store.chores.find((c) => c.id === completion.chore_id);
    const member = store.members.find((m) => m.id === completion.member_id);
    const admin = store.members.find((m) => m.id === adminMemberId);

    completion.status = 'rejected';
    completion.approved_by = adminMemberId;
    completion.rejection_reason = reason;
    completion.verified_at = new Date().toISOString();

    if (chore) {
      // Revert status to assigned for rework
      chore.status = 'assigned';
      chore.updated_at = new Date().toISOString();
    }

    store.activities.unshift({
      id: generateId('act'),
      household_id: completion.household_id,
      actor_id: adminMemberId,
      event_type: 'chore_rejected',
      entity_type: 'chore',
      entity_id: completion.chore_id,
      message: `${admin?.name || 'Admin'} requested rework for ${member?.name || 'Member'}'s "${chore?.title || 'chore'}": ${reason}`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay(true);
  },

  // Swapping operations
  async getSwaps(householdId: string): Promise<
    Array<{ swap: ChoreSwapRequest; chore: Chore; requester: Member; targetMember?: Member | null }>
  > {
    const store = loadStore();
    const result: Array<{ swap: ChoreSwapRequest; chore: Chore; requester: Member; targetMember?: Member | null }> = [];

    for (const swap of store.swaps) {
      if (swap.household_id === householdId) {
        const chore = store.chores.find((c) => c.id === swap.chore_id);
        const requester = store.members.find((m) => m.id === swap.requester_id);
        const target = swap.target_member_id ? store.members.find((m) => m.id === swap.target_member_id) : null;
        if (chore && requester) {
          result.push({
            swap: { ...swap },
            chore: { ...chore },
            requester: { ...requester },
            targetMember: target ? { ...target } : null,
          });
        }
      }
    }
    return delay(result);
  },

  async requestSwap(payload: {
    chore_id: string;
    requester_id: string;
    target_member_id?: string | null;
    reason?: string;
  }): Promise<ChoreSwapRequest> {
    const store = loadStore();
    const chore = store.chores.find((c) => c.id === payload.chore_id);
    if (!chore) throw new Error('Chore not found');

    const requester = store.members.find((m) => m.id === payload.requester_id);
    const target = payload.target_member_id ? store.members.find((m) => m.id === payload.target_member_id) : null;

    const newSwap: ChoreSwapRequest = {
      id: generateId('swap'),
      household_id: chore.household_id,
      chore_id: chore.id,
      requester_id: payload.requester_id,
      target_member_id: payload.target_member_id || null,
      status: 'pending',
      reason: payload.reason,
      created_at: new Date().toISOString(),
    };

    store.swaps.unshift(newSwap);

    const targetDesc = target ? `with ${target.name}` : 'on the Household Swap Board';
    store.activities.unshift({
      id: generateId('act'),
      household_id: chore.household_id,
      actor_id: payload.requester_id,
      event_type: 'swap_proposed',
      entity_type: 'swap',
      entity_id: newSwap.id,
      message: `${requester?.name || 'A member'} offered to swap "${chore.title}" ${targetDesc}`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay({ ...newSwap });
  },

  async acceptSwap(swapId: string, acceptorMemberId: string): Promise<boolean> {
    const store = loadStore();
    const swap = store.swaps.find((s) => s.id === swapId);
    if (!swap || swap.status !== 'pending') throw new Error('Swap request not available');

    const chore = store.chores.find((c) => c.id === swap.chore_id);
    if (!chore) throw new Error('Chore not found');

    const acceptor = store.members.find((m) => m.id === acceptorMemberId);

    // Transfer chore assignee
    chore.current_assignee_id = acceptorMemberId;
    chore.status = 'assigned';
    chore.updated_at = new Date().toISOString();

    swap.status = 'accepted';
    swap.resolved_at = new Date().toISOString();

    store.activities.unshift({
      id: generateId('act'),
      household_id: swap.household_id,
      actor_id: acceptorMemberId,
      event_type: 'swap_accepted',
      entity_type: 'swap',
      entity_id: swap.id,
      message: `${acceptor?.name || 'Member'} accepted the swap for "${chore.title}"`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay(true);
  },

  async rejectSwap(swapId: string): Promise<boolean> {
    const store = loadStore();
    const swap = store.swaps.find((s) => s.id === swapId);
    if (!swap) return delay(false);
    swap.status = 'rejected';
    swap.resolved_at = new Date().toISOString();
    saveStore(store);
    return delay(true);
  },

  async cancelSwap(swapId: string): Promise<boolean> {
    const store = loadStore();
    const swap = store.swaps.find((s) => s.id === swapId);
    if (!swap) return delay(false);
    swap.status = 'cancelled';
    swap.resolved_at = new Date().toISOString();
    saveStore(store);
    return delay(true);
  },

  // Rewards & Gamification
  async getRewards(householdId: string): Promise<RewardItem[]> {
    const store = loadStore();
    const list = store.rewards.filter((r) => r.household_id === householdId && r.is_active);
    return delay([...list]);
  },

  async createReward(
    householdId: string,
    payload: { title: string; description: string; points_cost: number; icon?: string }
  ): Promise<RewardItem> {
    const store = loadStore();
    const newReward: RewardItem = {
      id: generateId('rew'),
      household_id: householdId,
      title: payload.title,
      description: payload.description,
      points_cost: payload.points_cost,
      icon: payload.icon || 'Gift',
      is_active: true,
    };
    store.rewards.push(newReward);
    saveStore(store);
    return delay({ ...newReward });
  },

  async redeemReward(rewardId: string, memberId: string): Promise<RewardRedemption> {
    const store = loadStore();
    const reward = store.rewards.find((r) => r.id === rewardId);
    if (!reward) throw new Error('Reward not found');

    const member = store.members.find((m) => m.id === memberId);
    if (!member) throw new Error('Member not found');

    if (member.points_balance < reward.points_cost) {
      throw new Error(`Insufficient points: You need ${reward.points_cost} pts, but have ${member.points_balance} pts.`);
    }

    // Deduct points
    member.points_balance -= reward.points_cost;
    member.updated_at = new Date().toISOString();

    const redemption: RewardRedemption = {
      id: generateId('red'),
      household_id: reward.household_id,
      member_id: memberId,
      reward_item_id: rewardId,
      points_spent: reward.points_cost,
      status: 'requested',
      requested_at: new Date().toISOString(),
    };

    store.redemptions.unshift(redemption);

    store.activities.unshift({
      id: generateId('act'),
      household_id: reward.household_id,
      actor_id: memberId,
      event_type: 'reward_redeemed',
      entity_type: 'reward',
      entity_id: reward.id,
      message: `${member.name} redeemed "${reward.title}" for ${reward.points_cost} pts`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay({ ...redemption });
  },

  async fulfillReward(redemptionId: string): Promise<boolean> {
    const store = loadStore();
    const red = store.redemptions.find((r) => r.id === redemptionId);
    if (!red) throw new Error('Redemption not found');

    red.status = 'fulfilled';
    red.fulfilled_at = new Date().toISOString();

    const reward = store.rewards.find((r) => r.id === red.reward_item_id);
    const member = store.members.find((m) => m.id === red.member_id);

    store.activities.unshift({
      id: generateId('act'),
      household_id: red.household_id,
      actor_id: member?.id || 'admin',
      event_type: 'reward_fulfilled',
      entity_type: 'reward',
      entity_id: red.id,
      message: `Reward "${reward?.title || 'Perk'}" was fulfilled for ${member?.name || 'Member'}!`,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay(true);
  },

  async getRedemptions(householdId: string): Promise<
    Array<{ redemption: RewardRedemption; reward: RewardItem; member: Member }>
  > {
    const store = loadStore();
    const results: Array<{ redemption: RewardRedemption; reward: RewardItem; member: Member }> = [];

    for (const red of store.redemptions) {
      if (red.household_id === householdId) {
        const reward = store.rewards.find((r) => r.id === red.reward_item_id);
        const member = store.members.find((m) => m.id === red.member_id);
        if (reward && member) {
          results.push({
            redemption: { ...red },
            reward: { ...reward },
            member: { ...member },
          });
        }
      }
    }
    return delay(results);
  },

  // Activities & Nudges
  async getActivities(householdId: string): Promise<ActivityLog[]> {
    const store = loadStore();
    const list = store.activities.filter((a) => a.household_id === householdId);
    return delay([...list]);
  },

  async sendNudge(choreId: string, senderMemberId: string): Promise<ChoreNudgeResponse> {
    const store = loadStore();
    const chore = store.chores.find((c) => c.id === choreId);
    if (!chore) throw new Error('Chore not found');

    const sender = store.members.find((m) => m.id === senderMemberId);
    const assignee = chore.current_assignee_id ? store.members.find((m) => m.id === chore.current_assignee_id) : null;

    const nudgeMessage = assignee
      ? `Friendly nudge from ${sender?.name || 'a housemate'} to ${assignee.name}: "${chore.title}" is due soon!`
      : `Friendly reminder: Open chore "${chore.title}" is ready for someone to claim!`;

    store.activities.unshift({
      id: generateId('act'),
      household_id: chore.household_id,
      actor_id: senderMemberId,
      recipient_id: chore.current_assignee_id || null,
      event_type: 'chore_nudge',
      entity_type: 'chore',
      entity_id: chore.id,
      message: nudgeMessage,
      read: false,
      created_at: new Date().toISOString(),
    });

    saveStore(store);
    return delay({ success: true, message: nudgeMessage, email_dispatched: true });
  },

  // Chore comments
  async getComments(choreId: string): Promise<Array<{ comment: ChoreComment; member?: Member }>> {
    const store = loadStore();
    const comments = store.comments.filter((c) => c.chore_id === choreId);
    const results = comments.map((comm) => ({
      comment: { ...comm },
      member: store.members.find((m) => m.id === comm.member_id),
    }));
    return delay(results);
  },

  async addComment(choreId: string, memberId: string, message: string): Promise<ChoreComment> {
    const store = loadStore();
    const newComment: ChoreComment = {
      id: generateId('comm'),
      chore_id: choreId,
      member_id: memberId,
      message: message.trim(),
      created_at: new Date().toISOString(),
    };
    store.comments.push(newComment);
    saveStore(store);
    return delay({ ...newComment });
  },

  // Reset demo
  async resetToDefaultData(): Promise<void> {
    const initialStore: ChoreSyncStore = {
      households: INITIAL_HOUSEHOLDS,
      members: INITIAL_MEMBERS,
      chores: INITIAL_CHORES,
      completions: INITIAL_COMPLETIONS,
      swaps: INITIAL_SWAPS,
      rewards: INITIAL_REWARDS,
      redemptions: INITIAL_REDEMPTIONS,
      activities: INITIAL_ACTIVITIES,
      comments: INITIAL_COMMENTS,
    };
    saveStore(initialStore);
    return delay(undefined);
  },
};

/**
 * Executes a live backend call via httpClient with fallback to local storage
 * if the backend is offline or network connectivity fails.
 */
async function withFallback<T>(
  actionName: string,
  liveFn: () => Promise<T>,
  fallbackFn: () => Promise<T>
): Promise<T> {
  try {
    return await liveFn();
  } catch (error) {
    if (isNetworkError(error)) {
      console.warn(
        `[ChoreSync API] Live backend unreachable for "${actionName}". Falling back to local storage:`,
        error instanceof Error ? error.message : error
      );
      return await fallbackFn();
    }
    throw error;
  }
}

/**
 * ChoreSync Centralized API Service Layer
 * Connects to Go backend API (http://localhost:8000/api/v1) conforming to contracts/openapi.yaml,
 * with automatic fallback to localStorage/mock data when offline.
 */
export const api = {
  // Auth operations
  async login(email: string, password: string): Promise<AuthTokenResponse> {
    return withFallback(
      'login',
      async () => {
        const res = await httpClient.post<AuthTokenResponse>('/auth/login', { email, password });
        if (res?.access_token) {
          httpClient.setAuthToken(res.access_token);
        }
        return res;
      },
      () => localApi.login(email, password)
    );
  },

  async register(payload: {
    name: string;
    email: string;
    password: string;
    household_name?: string;
    invite_code?: string;
    role?: string;
  }): Promise<AuthTokenResponse> {
    return withFallback(
      'register',
      async () => {
        const res = await httpClient.post<AuthTokenResponse>('/auth/register', payload);
        if (res?.access_token) {
          httpClient.setAuthToken(res.access_token);
        }
        return res;
      },
      () => localApi.register(payload)
    );
  },

  async demoLogin(memberId: string): Promise<AuthTokenResponse> {
    const res = await httpClient.post<AuthTokenResponse>('/auth/demo-login', { member_id: memberId });
    if (res?.access_token) {
      httpClient.setAuthToken(res.access_token);
    }
    return res;
  },

  async requestMagicLink(email: string): Promise<{ message: string; token?: string }> {
    return httpClient.post<{ message: string; token?: string }>('/auth/magic-link', { email });
  },

  async verifyMagicLink(token: string): Promise<AuthTokenResponse> {
    const res = await httpClient.post<AuthTokenResponse>('/auth/verify', { token });
    if (res?.access_token) {
      httpClient.setAuthToken(res.access_token);
    }
    return res;
  },

  async verifyPIN(householdId: string, pin: string): Promise<{ valid: boolean; message: string }> {
    return httpClient.post<{ valid: boolean; message: string }>(`/households/${householdId}/verify-pin`, { pin });
  },

  async requestPasswordReset(email: string): Promise<ForgotPasswordResponse> {
    return httpClient.post<ForgotPasswordResponse>('/auth/forgot-password', { email });
  },

  async resetPassword(token: string, new_password: string): Promise<AuthTokenResponse> {
    const res = await httpClient.post<AuthTokenResponse>('/auth/reset-password', { token, new_password });
    if (res?.access_token) {
      httpClient.setAuthToken(res.access_token);
    }
    return res;
  },

  async getDevEmails(): Promise<DevEmail[]> {
    try {
      const res = await httpClient.get<DevEmail[]>('/dev/emails');
      return res || [];
    } catch {
      return [];
    }
  },

  async clearDevEmails(): Promise<{ success: boolean }> {
    return httpClient.delete<{ success: boolean }>('/dev/emails');
  },

  // Household operations
  async getHouseholds(): Promise<Household[]> {
    return withFallback(
      'getHouseholds',
      () => httpClient.get<Household[]>('/households'),
      () => localApi.getHouseholds()
    );
  },

  async getHousehold(id: string): Promise<Household | null> {
    return withFallback(
      `getHousehold(${id})`,
      async () => {
        try {
          return await httpClient.get<Household>(`/households/${id}`);
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) {
            return null;
          }
          throw err;
        }
      },
      () => localApi.getHousehold(id)
    );
  },

  async createHousehold(payload: {
    name: string;
    default_mode?: HouseholdMode;
    allow_swaps?: boolean;
    timezone?: string;
  }): Promise<Household> {
    return withFallback(
      'createHousehold',
      () => httpClient.post<Household>('/households', payload),
      () => localApi.createHousehold(payload)
    );
  },

  async joinHousehold(inviteCode: string): Promise<Household> {
    return withFallback(
      `joinHousehold(${inviteCode})`,
      () => httpClient.post<Household>('/households/join', { invite_code: inviteCode }),
      () => localApi.joinHousehold(inviteCode)
    );
  },

  async updateHouseholdSettings(id: string, settings: Partial<HouseholdSettings>): Promise<Household> {
    return withFallback(
      `updateHouseholdSettings(${id})`,
      () => httpClient.patch<Household>(`/households/${id}/settings`, settings),
      () => localApi.updateHouseholdSettings(id, settings)
    );
  },

  // Member operations
  async getMembers(householdId: string): Promise<Member[]> {
    return withFallback(
      `getMembers(${householdId})`,
      () => httpClient.get<Member[]>(`/households/${householdId}/members`),
      () => localApi.getMembers(householdId)
    );
  },

  async addMember(
    householdId: string,
    payload: { name: string; role: Member['role']; avatar_url?: string }
  ): Promise<Member> {
    return withFallback(
      `addMember(${householdId})`,
      () => httpClient.post<Member>(`/households/${householdId}/members`, payload),
      () => localApi.addMember(householdId, payload)
    );
  },

  // Chore operations
  async getChores(householdId: string, filterOptions?: ChoreFilterOptions): Promise<Chore[]> {
    return withFallback(
      `getChores(${householdId})`,
      () => {
        const params: Record<string, any> = {};
        if (filterOptions?.category) params.category = filterOptions.category;
        if (filterOptions?.status) params.status = filterOptions.status;
        if (filterOptions?.search) params.search = filterOptions.search;
        if (filterOptions?.view) params.view = filterOptions.view;
        if (filterOptions?.assignee_id) params.assignee_id = filterOptions.assignee_id;
        return httpClient.get<Chore[]>(`/households/${householdId}/chores`, params);
      },
      () => localApi.getChores(householdId, filterOptions)
    );
  },

  async getChore(choreId: string): Promise<Chore | null> {
    return withFallback(
      `getChore(${choreId})`,
      async () => {
        try {
          return await httpClient.get<Chore>(`/chores/${choreId}`);
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) {
            return null;
          }
          throw err;
        }
      },
      () => localApi.getChore(choreId)
    );
  },

  async createChore(
    payload: Omit<Chore, 'id' | 'created_at' | 'updated_at' | 'status'> & { status?: Chore['status'] }
  ): Promise<Chore> {
    return withFallback(
      'createChore',
      () => httpClient.post<Chore>(`/households/${payload.household_id}/chores`, payload),
      () => localApi.createChore(payload)
    );
  },

  async updateChore(id: string, payload: Partial<Chore>): Promise<Chore> {
    return withFallback(
      `updateChore(${id})`,
      () => httpClient.patch<Chore>(`/chores/${id}`, payload),
      () => localApi.updateChore(id, payload)
    );
  },

  async deleteChore(id: string): Promise<boolean> {
    return withFallback(
      `deleteChore(${id})`,
      async () => {
        try {
          await httpClient.delete(`/chores/${id}`);
          return true;
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) {
            return false;
          }
          throw err;
        }
      },
      () => localApi.deleteChore(id)
    );
  },

  // Claim chore from open pool
  async claimChore(choreId: string, memberId: string): Promise<Chore> {
    return withFallback(
      `claimChore(${choreId})`,
      () => httpClient.post<Chore>(`/chores/${choreId}/claim`, { member_id: memberId }),
      () => localApi.claimChore(choreId, memberId)
    );
  },

  // Complete chore
  async completeChore(
    choreId: string,
    memberId: string,
    proofPayload?: { proof_notes?: string; proof_photo_url?: string }
  ): Promise<{ chore: Chore; completion?: ChoreCompletion; pointsAwarded: number }> {
    return withFallback(
      `completeChore(${choreId})`,
      () =>
        httpClient.post<{ chore: Chore; completion?: ChoreCompletion; pointsAwarded: number }>(
          `/chores/${choreId}/complete`,
          {
            member_id: memberId,
            proof_notes: proofPayload?.proof_notes,
            proof_photo_url: proofPayload?.proof_photo_url,
          }
        ),
      () => localApi.completeChore(choreId, memberId, proofPayload)
    );
  },

  // Cloud Object Storage (S3 / Floci) - Photo proof upload
  async uploadPhoto(file: File): Promise<{ success: boolean; url: string; s3_key: string }> {
    const formData = new FormData();
    formData.append('file', file);
    return withFallback(
      'uploadPhoto',
      () => httpClient.post<{ success: boolean; url: string; s3_key: string }>('/uploads/photo', formData),
      () => localApi.uploadPhoto(file)
    );
  },

  // Rotate chore (auto-rotation engine)
  async rotateChore(choreId: string): Promise<Chore> {
    return withFallback(
      `rotateChore(${choreId})`,
      () => httpClient.post<Chore>(`/chores/${choreId}/rotate`),
      () => localApi.rotateChore(choreId)
    );
  },

  // Approvals
  async getPendingApprovals(
    householdId: string
  ): Promise<Array<{ completion: ChoreCompletion; chore: Chore; member: Member }>> {
    return withFallback(
      `getPendingApprovals(${householdId})`,
      () =>
        httpClient.get<Array<{ completion: ChoreCompletion; chore: Chore; member: Member }>>(
          `/households/${householdId}/approvals`
        ),
      () => localApi.getPendingApprovals(householdId)
    );
  },

  async approveChore(completionId: string, adminMemberId: string): Promise<boolean> {
    return withFallback(
      `approveChore(${completionId})`,
      async () => {
        await httpClient.post(`/completions/${completionId}/approve`, {
          admin_member_id: adminMemberId,
        });
        return true;
      },
      () => localApi.approveChore(completionId, adminMemberId)
    );
  },

  async rejectChore(completionId: string, adminMemberId: string, reason: string): Promise<boolean> {
    return withFallback(
      `rejectChore(${completionId})`,
      async () => {
        await httpClient.post(`/completions/${completionId}/reject`, {
          admin_member_id: adminMemberId,
          reason,
        });
        return true;
      },
      () => localApi.rejectChore(completionId, adminMemberId, reason)
    );
  },

  // Swapping operations
  async getSwaps(
    householdId: string
  ): Promise<
    Array<{ swap: ChoreSwapRequest; chore: Chore; requester: Member; targetMember?: Member | null }>
  > {
    return withFallback(
      `getSwaps(${householdId})`,
      async () => {
        const res = await httpClient.get<
          Array<{ swap: ChoreSwapRequest; chore: Chore; requester: Member; targetMember?: Member | null }>
        >(`/households/${householdId}/swaps`);
        return res || [];
      },
      () => localApi.getSwaps(householdId)
    );
  },

  async requestSwap(payload: {
    chore_id: string;
    requester_id: string;
    target_member_id?: string | null;
    reason?: string;
  }): Promise<ChoreSwapRequest> {
    return withFallback(
      `requestSwap(${payload.chore_id})`,
      () =>
        httpClient.post<ChoreSwapRequest>(`/chores/${payload.chore_id}/swap`, {
          requester_id: payload.requester_id,
          target_member_id: payload.target_member_id || null,
          reason: payload.reason,
        }),
      () => localApi.requestSwap(payload)
    );
  },

  async acceptSwap(swapId: string, acceptorMemberId: string): Promise<boolean> {
    return withFallback(
      `acceptSwap(${swapId})`,
      async () => {
        await httpClient.post(`/swaps/${swapId}/accept`, {
          acceptor_member_id: acceptorMemberId,
        });
        return true;
      },
      () => localApi.acceptSwap(swapId, acceptorMemberId)
    );
  },

  async rejectSwap(swapId: string): Promise<boolean> {
    return withFallback(
      `rejectSwap(${swapId})`,
      async () => {
        try {
          await httpClient.post(`/swaps/${swapId}/reject`);
          return true;
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) return false;
          throw err;
        }
      },
      () => localApi.rejectSwap(swapId)
    );
  },

  async cancelSwap(swapId: string): Promise<boolean> {
    return withFallback(
      `cancelSwap(${swapId})`,
      async () => {
        try {
          await httpClient.delete(`/swaps/${swapId}`);
          return true;
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) return false;
          throw err;
        }
      },
      () => localApi.cancelSwap(swapId)
    );
  },

  // Rewards & Gamification
  async getRewards(householdId: string): Promise<RewardItem[]> {
    return withFallback(
      `getRewards(${householdId})`,
      async () => {
        const res = await httpClient.get<RewardItem[]>(`/households/${householdId}/rewards`);
        return res || [];
      },
      () => localApi.getRewards(householdId)
    );
  },

  async createReward(
    householdId: string,
    payload: { title: string; description: string; points_cost: number; icon?: string }
  ): Promise<RewardItem> {
    return withFallback(
      `createReward(${householdId})`,
      () => httpClient.post<RewardItem>(`/households/${householdId}/rewards`, payload),
      () => localApi.createReward(householdId, payload)
    );
  },

  async redeemReward(rewardId: string, memberId: string): Promise<RewardRedemption> {
    return withFallback(
      `redeemReward(${rewardId})`,
      () => httpClient.post<RewardRedemption>(`/rewards/${rewardId}/redeem`, { member_id: memberId }),
      () => localApi.redeemReward(rewardId, memberId)
    );
  },

  async fulfillReward(redemptionId: string): Promise<boolean> {
    return withFallback(
      `fulfillReward(${redemptionId})`,
      async () => {
        await httpClient.post(`/redemptions/${redemptionId}/fulfill`);
        return true;
      },
      () => localApi.fulfillReward(redemptionId)
    );
  },

  async getRedemptions(
    householdId: string
  ): Promise<Array<{ redemption: RewardRedemption; reward: RewardItem; member: Member }>> {
    return withFallback(
      `getRedemptions(${householdId})`,
      async () => {
        const res = await httpClient.get<Array<{ redemption: RewardRedemption; reward: RewardItem; member: Member }>>(
          `/households/${householdId}/redemptions`
        );
        return res || [];
      },
      () => localApi.getRedemptions(householdId)
    );
  },

  // Activities & Nudges
  async getActivities(householdId: string): Promise<ActivityLog[]> {
    return withFallback(
      `getActivities(${householdId})`,
      async () => {
        const res = await httpClient.get<ActivityLog[]>(`/households/${householdId}/activity`);
        return res || [];
      },
      () => localApi.getActivities(householdId)
    );
  },

  async sendNudge(choreId: string, senderMemberId: string): Promise<ChoreNudgeResponse> {
    return withFallback(
      `sendNudge(${choreId})`,
      () =>
        httpClient.post<ChoreNudgeResponse>(`/chores/${choreId}/nudge`, {
          sender_member_id: senderMemberId,
        }),
      () => localApi.sendNudge(choreId, senderMemberId)
    );
  },

  // Chore comments
  async getComments(choreId: string): Promise<Array<{ comment: ChoreComment; member?: Member }>> {
    return withFallback(
      `getComments(${choreId})`,
      () => httpClient.get<Array<{ comment: ChoreComment; member?: Member }>>(`/chores/${choreId}/comments`),
      () => localApi.getComments(choreId)
    );
  },

  async addComment(choreId: string, memberId: string, message: string): Promise<ChoreComment> {
    return withFallback(
      `addComment(${choreId})`,
      () =>
        httpClient.post<ChoreComment>(`/chores/${choreId}/comments`, {
          member_id: memberId,
          message,
        }),
      () => localApi.addComment(choreId, memberId, message)
    );
  },

  // Reset demo
  async resetToDefaultData(): Promise<void> {
    httpClient.resetOfflineState();
    await localApi.resetToDefaultData();
    try {
      await httpClient.post('/reset');
    } catch (error) {
      if (!isNetworkError(error)) {
        throw error;
      }
    }
  },
};

// Helper: Advance recurrence and round-robin
function advanceRecurrence(store: ChoreSyncStore, chore: Chore) {
  if (chore.recurrence_type === 'none') return;

  // Calculate next due date
  const baseDate = chore.due_date ? new Date(chore.due_date) : new Date();
  const nextDate = new Date(baseDate);

  if (chore.recurrence_type === 'daily') {
    nextDate.setDate(nextDate.getDate() + 1);
  } else if (chore.recurrence_type === 'weekly' || chore.recurrence_type === 'custom_days') {
    nextDate.setDate(nextDate.getDate() + 7);
  } else if (chore.recurrence_type === 'monthly') {
    nextDate.setMonth(nextDate.getMonth() + 1);
  }

  // Handle round-robin member rotation
  let nextAssigneeId = chore.current_assignee_id;
  let nextIndex = chore.current_rotation_index || 0;

  if (chore.assignment_type === 'round_robin' && chore.rotation_member_ids && chore.rotation_member_ids.length > 0) {
    nextIndex = (nextIndex + 1) % chore.rotation_member_ids.length;
    nextAssigneeId = chore.rotation_member_ids[nextIndex];
  } else if (chore.assignment_type === 'open_pool') {
    nextAssigneeId = null;
  }

  const nextChoreInstance: Chore = {
    ...chore,
    id: generateId('c'),
    due_date: nextDate.toISOString(),
    current_assignee_id: nextAssigneeId,
    current_rotation_index: nextIndex,
    status: chore.assignment_type === 'open_pool' ? 'unassigned' : 'assigned',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  store.chores.unshift(nextChoreInstance);

  const nextAssignee = nextAssigneeId ? store.members.find((m) => m.id === nextAssigneeId) : null;
  const targetDesc = nextAssignee ? `Assigned to ${nextAssignee.name}` : 'Placed in Open Pool';

  store.activities.unshift({
    id: generateId('act'),
    household_id: chore.household_id,
    actor_id: chore.created_by,
    event_type: 'chore_created',
    entity_type: 'chore',
    entity_id: nextChoreInstance.id,
    message: `Scheduled next instance for "${chore.title}". ${targetDesc}`,
    read: false,
    created_at: new Date().toISOString(),
  });
}
