export type HouseholdMode = 'flatmate' | 'family' | 'casual';

export type MemberRole = 'admin' | 'member' | 'child';

export type ChoreCategory =
  | 'cleaning'
  | 'kitchen'
  | 'yard'
  | 'pets'
  | 'maintenance'
  | 'other';

export type ChoreAssignmentType = 'direct' | 'round_robin' | 'open_pool';

export type ChoreRecurrenceType =
  | 'none'
  | 'daily'
  | 'weekly'
  | 'monthly'
  | 'custom_days';

export interface ChoreRecurrenceRule {
  days?: string[]; // e.g. ['mon', 'thu']
}

export type ChoreStatus =
  | 'unassigned'
  | 'assigned'
  | 'in_progress'
  | 'pending_approval'
  | 'completed'
  | 'overdue';

export interface HouseholdSettings {
  default_mode: HouseholdMode;
  allow_swaps: boolean;
  timezone: string;
}

export interface Household {
  id: string;
  name: string;
  invite_code: string;
  settings: HouseholdSettings;
  created_at: string;
  updated_at: string;
}

export interface Member {
  id: string;
  household_id: string;
  name: string;
  avatar_url?: string;
  role: MemberRole;
  points_balance: number;
  total_points_earned: number;
  streak: number;
  created_at: string;
  updated_at: string;
}

export interface Chore {
  id: string;
  household_id: string;
  title: string;
  description?: string;
  category: ChoreCategory;
  effort_points: number;
  assignment_type: ChoreAssignmentType;
  current_assignee_id?: string | null;
  rotation_member_ids?: string[];
  current_rotation_index?: number;
  recurrence_type: ChoreRecurrenceType;
  recurrence_rule?: ChoreRecurrenceRule;
  due_date?: string | null;
  requires_approval: boolean;
  requires_proof: boolean;
  status: ChoreStatus;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export type CompletionStatus = 'approved' | 'rejected' | 'pending_approval';

export interface ChoreCompletion {
  id: string;
  chore_id: string;
  household_id: string;
  member_id: string;
  status: CompletionStatus;
  proof_notes?: string;
  proof_photo_url?: string;
  approved_by?: string | null;
  rejection_reason?: string;
  points_awarded: number;
  completed_at: string;
  verified_at?: string | null;
}

export type SwapStatus = 'pending' | 'accepted' | 'rejected' | 'cancelled';

export interface ChoreSwapRequest {
  id: string;
  household_id: string;
  chore_id: string;
  requester_id: string;
  target_member_id?: string | null; // null means open to entire household
  status: SwapStatus;
  reason?: string;
  created_at: string;
  resolved_at?: string | null;
}

export interface RewardItem {
  id: string;
  household_id: string;
  title: string;
  description: string;
  points_cost: number;
  icon?: string;
  is_active: boolean;
}

export type RedemptionStatus = 'requested' | 'fulfilled' | 'cancelled';

export interface RewardRedemption {
  id: string;
  household_id: string;
  member_id: string;
  reward_item_id: string;
  points_spent: number;
  status: RedemptionStatus;
  requested_at: string;
  fulfilled_at?: string | null;
}

export type ActivityEventType =
  | 'chore_created'
  | 'chore_assigned'
  | 'chore_completed'
  | 'approval_requested'
  | 'chore_approved'
  | 'chore_rejected'
  | 'swap_proposed'
  | 'swap_accepted'
  | 'chore_nudge'
  | 'reward_redeemed'
  | 'reward_fulfilled';

export interface ActivityLog {
  id: string;
  household_id: string;
  recipient_id?: string | null;
  actor_id: string;
  event_type: ActivityEventType;
  entity_type: 'chore' | 'swap' | 'reward';
  entity_id: string;
  message: string;
  read: boolean;
  created_at: string;
}

export interface ChoreComment {
  id: string;
  chore_id: string;
  member_id: string;
  message: string;
  created_at: string;
}

export interface ChoreFilterOptions {
  status?: string;
  assignee_id?: string;
  category?: ChoreCategory;
  search?: string;
  view?: 'all' | 'my' | 'open' | 'approval' | 'overdue';
}
