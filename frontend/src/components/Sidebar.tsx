import React from 'react';
import {
  CheckSquare,
  ClipboardCheck,
  Repeat,
  Gift,
  History,
  Settings,
  Flame,
  Users,
  Award,
} from 'lucide-react';
import { HouseholdMode, MemberRole } from '../types';

export type NavTab =
  | 'dashboard'
  | 'approvals'
  | 'swaps'
  | 'rewards'
  | 'activity'
  | 'settings';

interface SidebarProps {
  currentTab: NavTab;
  onSelectTab: (tab: NavTab) => void;
  pendingApprovalsCount: number;
  pendingSwapsCount: number;
  householdMode: HouseholdMode;
  userRole: MemberRole;
  choresCount: number;
  completedChoresCount: number;
}

export const Sidebar: React.FC<SidebarProps> = ({
  currentTab,
  onSelectTab,
  pendingApprovalsCount,
  pendingSwapsCount,
  householdMode,
  userRole,
  choresCount,
  completedChoresCount,
}) => {
  const navItems = [
    {
      id: 'dashboard' as NavTab,
      label: 'Chores Board',
      icon: CheckSquare,
      badge: choresCount > 0 ? choresCount : undefined,
    },
    {
      id: 'approvals' as NavTab,
      label: 'Approval Queue',
      icon: ClipboardCheck,
      badge: pendingApprovalsCount > 0 ? pendingApprovalsCount : undefined,
      badgeColor: 'bg-amber-500 text-white',
      // Highlight in family mode or for admin
      highlight: householdMode === 'family',
    },
    {
      id: 'swaps' as NavTab,
      label: 'Swap Market',
      icon: Repeat,
      badge: pendingSwapsCount > 0 ? pendingSwapsCount : undefined,
      badgeColor: 'bg-indigo-500 text-white',
      // Highlight in flatmate mode
      highlight: householdMode === 'flatmate',
    },
    {
      id: 'rewards' as NavTab,
      label: 'Rewards Hub',
      icon: Gift,
      badge: undefined,
    },
    {
      id: 'activity' as NavTab,
      label: 'Audit & Feed',
      icon: History,
      badge: undefined,
    },
    {
      id: 'settings' as NavTab,
      label: 'House & Roster',
      icon: Settings,
      badge: undefined,
    },
  ];

  return (
    <>
      {/* Desktop Sidebar */}
      <aside className="hidden md:flex flex-col w-64 shrink-0 border-r border-zinc-200 bg-white min-h-[calc(100vh-61px)] p-4">
        {/* Mode context helper banner */}
        <div className="mb-4 p-3 rounded-xl bg-zinc-50 border border-zinc-200/80">
          <div className="flex items-center justify-between text-xs font-semibold text-zinc-900 mb-1">
            <span className="flex items-center gap-1.5">
              <Users className="w-3.5 h-3.5 text-zinc-500" />
              {householdMode === 'flatmate'
                ? 'Flatmate Setup'
                : householdMode === 'family'
                ? 'Family Hierarchy'
                : 'Couple Shared'}
            </span>
          </div>
          <p className="text-[11px] text-zinc-500 leading-relaxed">
            {householdMode === 'flatmate' && 'Rotations, peer chore swaps & fair distribution history.'}
            {householdMode === 'family' && 'Parent approvals, visual checklist, and reward points.'}
            {householdMode === 'casual' && 'Low-overhead open pool and 1-tap quick completions.'}
          </p>
        </div>

        {/* Navigation list */}
        <nav className="flex-1 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = currentTab === item.id;
            return (
              <button
                key={item.id}
                id={`nav-btn-${item.id}`}
                onClick={() => onSelectTab(item.id)}
                className={`w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-sm font-medium transition-colors text-left cursor-pointer ${
                  isActive
                    ? 'bg-zinc-900 text-white font-semibold shadow-xs'
                    : 'text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100'
                }`}
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <Icon className={`w-4 h-4 shrink-0 ${isActive ? 'text-white' : 'text-zinc-500'}`} />
                  <span className="truncate">{item.label}</span>
                </div>
                {item.badge !== undefined && (
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full font-bold shrink-0 ${
                      isActive
                        ? 'bg-zinc-800 text-zinc-200'
                        : item.badgeColor || 'bg-zinc-200 text-zinc-700'
                    }`}
                  >
                    {item.badge}
                  </span>
                )}
              </button>
            );
          })}
        </nav>

        {/* Quick Household Progress Widget */}
        <div className="mt-auto pt-4 border-t border-zinc-100">
          <div className="p-3 bg-emerald-50/70 border border-emerald-200/60 rounded-xl">
            <div className="flex items-center justify-between text-xs font-semibold text-emerald-950 mb-1.5">
              <span className="flex items-center gap-1">
                <Award className="w-3.5 h-3.5 text-emerald-600" /> Household Pace
              </span>
              <span>{completedChoresCount} Done</span>
            </div>
            <div className="w-full bg-emerald-200/60 h-1.5 rounded-full overflow-hidden">
              <div
                className="bg-emerald-600 h-full transition-all duration-300"
                style={{
                  width: `${
                    choresCount + completedChoresCount > 0
                      ? Math.min(
                          100,
                          Math.round((completedChoresCount / (choresCount + completedChoresCount)) * 100)
                        )
                      : 0
                  }%`,
                }}
              />
            </div>
            <div className="mt-1.5 flex justify-between text-[11px] text-emerald-700">
              <span>{choresCount} active tasks remaining</span>
            </div>
          </div>
        </div>
      </aside>

      {/* Mobile Bottom Navigation Bar */}
      <nav className="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-white/95 backdrop-blur border-t border-zinc-200 px-2 py-1.5 flex items-center justify-around">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = currentTab === item.id;
          return (
            <button
              key={item.id}
              id={`mobile-nav-${item.id}`}
              onClick={() => onSelectTab(item.id)}
              className={`relative flex flex-col items-center gap-0.5 p-1.5 min-w-[50px] rounded-lg transition-colors cursor-pointer ${
                isActive ? 'text-zinc-950 font-semibold' : 'text-zinc-400 hover:text-zinc-700'
              }`}
            >
              <div className="relative">
                <Icon className={`w-5 h-5 ${isActive ? 'text-zinc-900 stroke-[2.2]' : 'text-zinc-400'}`} />
                {item.badge !== undefined && (
                  <span
                    className={`absolute -top-1 -right-2 text-[10px] w-4 h-4 rounded-full flex items-center justify-center font-bold ${
                      item.badgeColor || 'bg-zinc-800 text-white'
                    }`}
                  >
                    {item.badge}
                  </span>
                )}
              </div>
              <span className="text-[10px] tracking-tight">{item.label.split(' ')[0]}</span>
            </button>
          );
        })}
      </nav>
    </>
  );
};
