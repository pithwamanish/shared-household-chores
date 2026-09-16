import React, { useState, useRef, useEffect } from 'react';
import { Household, Member, HouseholdMode, DeviceMode } from '../types';
import {
  Sparkles,
  Users,
  Flame,
  Plus,
  Coins,
  Bell,
  ChevronDown,
  Home,
  Tablet,
  Smartphone,
  LogOut,
  Info,
  KeyRound,
} from 'lucide-react';

interface NavbarProps {
  activeHousehold: Household | null;
  onOpenHouseholdInfo: () => void;
  members: Member[];
  activeMember: Member | null;
  onSignOut: () => void;
  onOpenCreateChore: () => void;
  onOpenActivities: () => void;
  unreadActivitiesCount: number;
  pendingApprovalsCount: number;
  onModeChange?: (mode: HouseholdMode) => void;
  deviceMode: DeviceMode;
  onToggleDeviceMode: (mode: DeviceMode) => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeHousehold,
  onOpenHouseholdInfo,
  members,
  activeMember,
  onSignOut,
  onOpenCreateChore,
  onOpenActivities,
  unreadActivitiesCount,
  pendingApprovalsCount,
  onModeChange,
  deviceMode,
  onToggleDeviceMode,
}) => {
  const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);
  const profileMenuRef = useRef<HTMLDivElement>(null);

  // Close dropdown on click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
        setIsProfileMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const mode = activeHousehold?.settings.default_mode || 'flatmate';

  const modeBadgeText = {
    flatmate: 'Flatmate Mode',
    family: 'Family Mode',
    casual: 'Couple Mode',
  }[mode] || mode;

  const modeBadgeColor = {
    flatmate: 'bg-indigo-50 text-indigo-700 border-indigo-200',
    family: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    casual: 'bg-rose-50 text-rose-700 border-rose-200',
  }[mode] || 'bg-zinc-50 text-zinc-700 border-zinc-200';

  return (
    <header className="sticky top-0 z-30 bg-white/95 backdrop-blur border-b border-zinc-200 px-4 lg:px-6 py-3">
      <div className="max-w-7xl mx-auto flex items-center justify-between gap-3">
        {/* Brand & Household Badge */}
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

          {/* Household Badge (Click to open household info & invite code) */}
          <div className="relative flex items-center">
            <button
              type="button"
              id="household-badge-btn"
              onClick={onOpenHouseholdInfo}
              className="relative inline-flex items-center gap-2 bg-zinc-100 hover:bg-zinc-200/80 transition-colors rounded-xl px-3 py-1.5 border border-zinc-200/80 cursor-pointer text-left"
              title="View Household Details & Invite Code"
            >
              <Home className="w-4 h-4 text-zinc-600 shrink-0" />
              <span className="text-sm font-bold text-zinc-900 leading-tight">
                {activeHousehold?.name || 'Household'}
              </span>
              <Info className="w-3.5 h-3.5 text-zinc-400" />
            </button>

            {/* Household mode tag */}
            <span
              className={`hidden md:inline-flex ml-2 items-center px-2 py-0.5 rounded-full text-xs font-medium border ${modeBadgeColor}`}
            >
              {modeBadgeText}
            </span>
          </div>
        </div>

        {/* Action Bar & User Profile */}
        <div className="flex items-center gap-2.5 shrink-0">
          {/* Segmented Device Mode Toggle */}
          <div
            id="device-mode-toggle"
            className="hidden md:inline-flex items-center bg-zinc-100 p-0.5 rounded-xl border border-zinc-200/80 shadow-2xs"
            role="group"
            aria-label="Device Mode Toggle"
          >
            <button
              type="button"
              id="toggle-mode-tablet"
              onClick={() => onToggleDeviceMode('tablet_kiosk')}
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold transition-all cursor-pointer ${
                deviceMode === 'tablet_kiosk'
                  ? 'bg-white text-zinc-900 shadow-xs'
                  : 'text-zinc-500 hover:text-zinc-800'
              }`}
              title="Shared Kitchen Tablet: All-household board for shared counter display"
            >
              <Tablet className="w-3.5 h-3.5 text-indigo-600" />
              <span>Kitchen Tablet</span>
            </button>
            <button
              type="button"
              id="toggle-mode-personal"
              onClick={() => onToggleDeviceMode('personal')}
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-bold transition-all cursor-pointer ${
                deviceMode === 'personal'
                  ? 'bg-white text-zinc-900 shadow-xs'
                  : 'text-zinc-500 hover:text-zinc-800'
              }`}
              title="Personal Device: Defaults to My Tasks and tracks your personal responsibilities"
            >
              <Smartphone className="w-3.5 h-3.5 text-emerald-600" />
              <span>Personal Device</span>
            </button>
          </div>

          {/* Compact Toggle for Mobile (< md) */}
          <button
            type="button"
            onClick={() => onToggleDeviceMode(deviceMode === 'tablet_kiosk' ? 'personal' : 'tablet_kiosk')}
            className="inline-flex md:hidden items-center justify-center p-1.5 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-700 border border-zinc-200 text-xs cursor-pointer"
            title={deviceMode === 'tablet_kiosk' ? 'Switch to Personal Device' : 'Switch to Kitchen Tablet'}
            aria-label="Toggle device mode"
          >
            {deviceMode === 'tablet_kiosk' ? (
              <Tablet className="w-4 h-4 text-indigo-600" />
            ) : (
              <Smartphone className="w-4 h-4 text-emerald-600" />
            )}
          </button>

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

          {/* Authenticated User Profile Menu */}
          <div className="relative" ref={profileMenuRef}>
            <button
              type="button"
              id="user-profile-button"
              onClick={() => setIsProfileMenuOpen((prev) => !prev)}
              className="flex items-center gap-2 bg-zinc-100 hover:bg-zinc-200/80 transition-colors rounded-xl px-2.5 py-1.5 border border-zinc-200 cursor-pointer"
              aria-label="User account menu"
            >
              <div className="relative shrink-0">
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
                    title="Admin Role"
                  >
                    ★
                  </span>
                )}
              </div>

              <div className="flex flex-col text-left hidden sm:block">
                <span className="text-xs font-bold text-zinc-900 leading-tight">
                  {activeMember?.name}
                </span>
                <span className="text-[10px] text-zinc-500 font-medium leading-none capitalize">
                  {activeMember?.role}
                </span>
              </div>

              <ChevronDown className="w-3.5 h-3.5 text-zinc-400 ml-0.5" />
            </button>

            {/* Profile Dropdown Menu */}
            {isProfileMenuOpen && (
              <div
                id="user-profile-dropdown"
                className="absolute right-0 mt-2 w-64 bg-white rounded-2xl shadow-xl border border-zinc-200 py-2 z-50 animate-in fade-in zoom-in-95 duration-100 text-left"
              >
                {/* User Identity Details */}
                <div className="px-3.5 py-2.5 border-b border-zinc-100">
                  <div className="font-bold text-zinc-900 text-sm">{activeMember?.name}</div>
                  <div className="text-xs text-zinc-500 truncate mt-0.5">
                    {activeMember?.email || `${activeMember?.name.toLowerCase().replace(/\s+/g, '.')}@choresync.home`}
                  </div>
                  <div className="flex items-center gap-2 mt-2">
                    <span className="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-zinc-100 text-zinc-700 border border-zinc-200">
                      {activeMember?.role}
                    </span>
                    <span className="text-xs text-amber-700 font-bold">
                      {activeMember?.points_balance} spendable pts
                    </span>
                  </div>
                </div>

                {/* Household Info */}
                <div className="p-1.5 space-y-1">
                  <button
                    type="button"
                    id="profile-household-info-btn"
                    onClick={() => {
                      setIsProfileMenuOpen(false);
                      onOpenHouseholdInfo();
                    }}
                    className="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs text-zinc-700 hover:bg-zinc-100 transition-colors text-left cursor-pointer font-medium"
                  >
                    <div className="flex items-center gap-2">
                      <Home className="w-4 h-4 text-indigo-600" />
                      <span>{activeHousehold?.name || 'Household'} Details</span>
                    </div>
                    <KeyRound className="w-3.5 h-3.5 text-zinc-400" />
                  </button>

                  <button
                    type="button"
                    id="profile-sign-out-btn"
                    onClick={() => {
                      setIsProfileMenuOpen(false);
                      onSignOut();
                    }}
                    className="w-full flex items-center gap-2 px-3 py-2 rounded-xl text-xs text-rose-600 hover:bg-rose-50 transition-colors text-left cursor-pointer font-bold"
                  >
                    <LogOut className="w-4 h-4 text-rose-500" />
                    <span>Sign Out</span>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};
