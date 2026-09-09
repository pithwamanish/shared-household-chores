import React from 'react';
import { Household, Member, HouseholdMode } from '../types';
import {
  Sparkles,
  Users,
  Flame,
  Plus,
  Coins,
  Bell,
  ChevronDown,
  ShieldCheck,
  Smile,
  Home,
} from 'lucide-react';

interface NavbarProps {
  households: Household[];
  activeHousehold: Household | null;
  onSelectHousehold: (householdId: string) => void;
  members: Member[];
  activeMember: Member | null;
  onSelectMember: (memberId: string) => void;
  onOpenCreateChore: () => void;
  onOpenActivities: () => void;
  unreadActivitiesCount: number;
  pendingApprovalsCount: number;
  onModeChange?: (mode: HouseholdMode) => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  households,
  activeHousehold,
  onSelectHousehold,
  members,
  activeMember,
  onSelectMember,
  onOpenCreateChore,
  onOpenActivities,
  unreadActivitiesCount,
  pendingApprovalsCount,
  onModeChange,
}) => {
  const mode = activeHousehold?.settings.default_mode || 'flatmate';

  const modeBadgeText = {
    flatmate: 'Flatmate Mode',
    family: 'Family Mode',
    casual: 'Couple Mode',
  }[mode];

  const modeBadgeColor = {
    flatmate: 'bg-indigo-50 text-indigo-700 border-indigo-200',
    family: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    casual: 'bg-rose-50 text-rose-700 border-rose-200',
  }[mode];

  return (
    <header className="sticky top-0 z-30 bg-white/95 backdrop-blur border-b border-zinc-200 px-4 lg:px-6 py-3">
      <div className="max-w-7xl mx-auto flex items-center justify-between gap-3">
        {/* Brand & Household Switcher */}
        <div className="flex items-center gap-3 md:gap-5 min-w-0">
          <div className="flex items-center gap-2 shrink-0">
            <div className="w-9 h-9 rounded-xl bg-zinc-900 text-white flex items-center justify-center font-bold tracking-tight shadow-sm">
              <Sparkles className="w-5 h-5 text-amber-400" />
            </div>
            <div className="hidden sm:block">
              <div className="font-bold text-zinc-900 leading-none tracking-tight text-base">ChoreSync</div>
              <div className="text-[11px] text-zinc-500 font-medium leading-tight">Shared Home Harmony</div>
            </div>
          </div>

          {/* Household selector */}
          <div className="relative flex items-center">
            <div className="relative inline-flex items-center bg-zinc-100 hover:bg-zinc-200/70 transition-colors rounded-xl px-2.5 py-1.5 border border-zinc-200/80">
              <Home className="w-4 h-4 text-zinc-500 mr-2 shrink-0" />
              <select
                id="household-selector"
                value={activeHousehold?.id || ''}
                onChange={(e) => onSelectHousehold(e.target.value)}
                aria-label="Select household"
                className="bg-transparent text-sm font-semibold text-zinc-800 pr-6 focus:outline-none cursor-pointer appearance-none"
              >
                {households.map((h) => (
                  <option key={h.id} value={h.id}>
                    {h.name} ({h.settings.default_mode})
                  </option>
                ))}
              </select>
              <ChevronDown className="w-3.5 h-3.5 text-zinc-400 absolute right-2 pointer-events-none" />
            </div>

            {/* Household mode tag */}
            <span
              className={`hidden md:inline-flex ml-2 items-center px-2 py-0.5 rounded-full text-xs font-medium border ${modeBadgeColor}`}
            >
              {modeBadgeText}
            </span>
          </div>
        </div>

        {/* User Persona & Action Bar */}
        <div className="flex items-center gap-2.5 shrink-0">
          {/* Active Persona Switcher */}
          <div className="flex items-center bg-zinc-100/90 rounded-xl px-2.5 py-1 border border-zinc-200">
            <div className="flex items-center gap-2">
              <div className="relative">
                {activeMember?.avatar_url ? (
                  <img
                    src={activeMember.avatar_url}
                    alt={activeMember.name}
                    className="w-7 h-7 rounded-full object-cover border border-white shadow-xs"
                  />
                ) : (
                  <div className="w-7 h-7 rounded-full bg-zinc-300 flex items-center justify-center text-xs font-semibold">
                    {activeMember?.name[0] || 'U'}
                  </div>
                )}
                {activeMember?.role === 'admin' && (
                  <span
                    className="absolute -top-1 -right-1 w-3.5 h-3.5 bg-indigo-600 rounded-full flex items-center justify-center text-[8px] text-white font-bold"
                    title="Admin / Parent Role"
                  >
                    ★
                  </span>
                )}
                {activeMember?.role === 'child' && (
                  <span
                    className="absolute -top-1 -right-1 w-3.5 h-3.5 bg-amber-500 rounded-full flex items-center justify-center text-[8px] text-white font-bold"
                    title="Child Persona"
                  >
                    ●
                  </span>
                )}
              </div>

              <div className="flex flex-col text-left">
                <div className="flex items-center gap-1">
                  <span className="text-xs text-zinc-500 font-medium hidden lg:inline">Role persona:</span>
                  <div className="relative inline-flex items-center">
                    <select
                      id="persona-selector"
                      value={activeMember?.id || ''}
                      onChange={(e) => onSelectMember(e.target.value)}
                      aria-label="Active member persona"
                      className="bg-transparent text-xs font-semibold text-zinc-900 pr-4 focus:outline-none cursor-pointer appearance-none"
                    >
                      {members.map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.name} ({m.role.toUpperCase()})
                        </option>
                      ))}
                    </select>
                    <ChevronDown className="w-3 h-3 text-zinc-400 absolute right-0 pointer-events-none" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Spendable Points Badge */}
          {activeMember && (
            <div
              className="hidden sm:flex items-center gap-2 bg-amber-50/80 border border-amber-200/80 px-3 py-1.5 rounded-xl text-amber-900"
              title="Spendable points & current completion streak"
            >
              <div className="flex items-center gap-1 font-semibold text-xs">
                <Coins className="w-3.5 h-3.5 text-amber-600" />
                <span>{activeMember.points_balance} pts</span>
              </div>
              {activeMember.streak > 0 && (
                <div className="flex items-center gap-0.5 text-[11px] font-medium text-orange-600 pl-1.5 border-l border-amber-200">
                  <Flame className="w-3 h-3 text-orange-500" />
                  <span>{activeMember.streak}d</span>
                </div>
              )}
            </div>
          )}

          {/* Activity / Notification Bell */}
          <button
            id="notifications-button"
            onClick={onOpenActivities}
            className="relative p-2 text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 rounded-xl transition-colors cursor-pointer"
            aria-label="Activity logs and notifications"
          >
            <Bell className="w-5 h-5" />
            {unreadActivitiesCount > 0 && (
              <span className="absolute top-1 right-1 w-2.5 h-2.5 bg-rose-500 rounded-full ring-2 ring-white animate-pulse" />
            )}
          </button>

          {/* New Chore Button */}
          <button
            id="add-chore-button"
            onClick={onOpenCreateChore}
            className="flex items-center gap-1.5 bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 text-white text-xs sm:text-sm font-semibold px-3.5 py-2 rounded-xl transition-all shadow-xs cursor-pointer"
          >
            <Plus className="w-4 h-4 stroke-[2.5]" />
            <span className="hidden sm:inline">New Chore</span>
            <span className="sm:hidden">Add</span>
          </button>
        </div>
      </div>
    </header>
  );
};
