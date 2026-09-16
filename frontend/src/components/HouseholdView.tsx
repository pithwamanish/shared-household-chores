import React, { useState } from 'react';
import { Household, Member, HouseholdMode, MemberRole } from '../types';
import {
  Home,
  Users,
  Copy,
  Check,
  Plus,
  RotateCcw,
  Sparkles,
  Shield,
  Smile,
  Coins,
  Flame,
  UserPlus,
  Share2,
  Lock,
  X,
} from 'lucide-react';

interface HouseholdViewProps {
  household: Household | null;
  members: Member[];
  currentMember?: Member | null;
  onUpdateMode: (mode: HouseholdMode) => Promise<void>;
  onUpdateAdminPin?: (pin: string) => Promise<void>;
  onAddMember: (payload: { name: string; role: MemberRole }) => Promise<void>;
  onResetData: () => Promise<void>;
}

export const HouseholdView: React.FC<HouseholdViewProps> = ({
  household,
  members,
  currentMember,
  onUpdateMode,
  onUpdateAdminPin,
  onAddMember,
  onResetData,
}) => {
  const [copied, setCopied] = useState(false);
  const [isAddMemberOpen, setIsAddMemberOpen] = useState(false);
  const [newMemberName, setNewMemberName] = useState('');
  const [newMemberRole, setNewMemberRole] = useState<MemberRole>('member');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Admin PIN modification state
  const [isEditPinOpen, setIsEditPinOpen] = useState(false);
  const [newPin, setNewPin] = useState('');
  const [pinError, setPinError] = useState('');
  const [isSavingPin, setIsSavingPin] = useState(false);

  const handleSavePinSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const clean = newPin.trim();
    if (clean.length !== 4 || !/^\d{4}$/.test(clean)) {
      setPinError('Admin PIN must be exactly 4 digits (0-9)');
      return;
    }
    setIsSavingPin(true);
    setPinError('');
    try {
      if (onUpdateAdminPin) {
        await onUpdateAdminPin(clean);
      }
      setIsEditPinOpen(false);
      setNewPin('');
    } catch (err: any) {
      setPinError(err?.message || 'Failed to update Admin PIN');
    } finally {
      setIsSavingPin(false);
    }
  };

  const currentMode = household?.settings.default_mode || 'flatmate';

  const handleCopyInvite = () => {
    if (household?.invite_code) {
      const inviteUrl = `${window.location.origin}/?invite=${encodeURIComponent(household.invite_code)}`;
      navigator.clipboard.writeText(inviteUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    }
  };

  const handleAddMemberSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMemberName.trim()) return;
    setIsSubmitting(true);
    try {
      await onAddMember({
        name: newMemberName.trim(),
        role: newMemberRole,
      });
      setIsAddMemberOpen(false);
      setNewMemberName('');
    } catch (err) {
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Household Overview & Invite Banner */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-xs">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <div className="w-8 h-8 rounded-xl bg-zinc-900 text-white flex items-center justify-center font-bold">
              <Home className="w-4 h-4" />
            </div>
            <h1 className="text-lg font-bold text-zinc-900 tracking-tight">{household?.name}</h1>
          </div>
          <p className="text-xs text-zinc-500">
            Household workspace settings, member permissions, and operational mode presets.
          </p>
        </div>

        {/* Invite Link & PIN Badges */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Admin PIN Badge */}
          <div className="flex items-center gap-2 bg-amber-50/70 border border-amber-200/80 p-2 rounded-xl">
            <Lock className="w-3.5 h-3.5 text-amber-700 ml-1" />
            <div className="text-left px-1">
              <div className="text-[10px] text-amber-700 font-bold uppercase tracking-wider">Admin PIN</div>
              <div className="font-mono text-xs font-bold text-amber-900">
                {household?.settings.admin_pin || '1234'}
              </div>
            </div>
            {currentMember?.role === 'admin' ? (
              <button
                type="button"
                id="edit-admin-pin-btn"
                onClick={() => {
                  setNewPin(household?.settings.admin_pin || '1234');
                  setPinError('');
                  setIsEditPinOpen(true);
                }}
                className="text-xs font-semibold text-amber-900 hover:text-amber-950 bg-amber-200/70 hover:bg-amber-300/80 px-2 py-1 rounded-lg transition-colors cursor-pointer ml-1 border border-amber-300 shadow-2xs"
                title="Change Admin PIN for kiosk unlock & approvals"
              >
                Change PIN
              </button>
            ) : (
              <span className="text-[10px] text-amber-600/80 font-medium px-1" title="Only admins can modify PIN">
                (Admin only)
              </span>
            )}
          </div>

          {/* Invite Code / Link Badge */}
          <div className="flex items-center gap-2 bg-zinc-50 border border-zinc-200/80 p-2 rounded-xl">
            <div className="text-left px-2">
              <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-wider">Invite Link</div>
              <div className="font-mono text-xs font-bold text-zinc-800">{household?.invite_code}</div>
            </div>
            <button
              onClick={handleCopyInvite}
              title="Copy shareable invite link for roommate/partner"
              className="flex items-center gap-1 bg-white hover:bg-zinc-100 text-zinc-700 text-xs font-semibold py-1.5 px-3 rounded-lg border border-zinc-200 transition-colors cursor-pointer shadow-2xs"
            >
              {copied ? (
                <>
                  <Check className="w-3.5 h-3.5 text-emerald-600" />
                  <span>Link Copied!</span>
                </>
              ) : (
                <>
                  <Share2 className="w-3.5 h-3.5 text-zinc-500" />
                  <span>Copy Link</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Household Dynamic Mode Presets */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 shadow-xs space-y-4">
        <div>
          <h2 className="text-xs font-bold text-zinc-800 uppercase tracking-wider mb-1">
            Household Dynamic & Rules Preset
          </h2>
          <p className="text-xs text-zinc-500">
            Switch how ChoreSync coordinates responsibilities, approval workflows, and task claiming.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {/* Flatmate Preset */}
          <button
            type="button"
            onClick={() => onUpdateMode('flatmate')}
            className={`p-4 rounded-2xl border text-left transition-all cursor-pointer ${
              currentMode === 'flatmate'
                ? 'bg-indigo-50/70 border-indigo-300 ring-2 ring-indigo-500/20'
                : 'bg-zinc-50/60 hover:bg-zinc-100/60 border-zinc-200'
            }`}
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-bold text-sm text-zinc-900">Flatmates / Roommates</span>
              {currentMode === 'flatmate' && (
                <span className="text-[10px] font-bold bg-indigo-600 text-white px-2 py-0.5 rounded-full">
                  Active
                </span>
              )}
            </div>
            <p className="text-xs text-zinc-600 leading-relaxed">
              Automated round-robin rotations, chore trading / swapping marketplace, and transparent contribution histories.
            </p>
          </button>

          {/* Family Preset */}
          <button
            type="button"
            onClick={() => onUpdateMode('family')}
            className={`p-4 rounded-2xl border text-left transition-all cursor-pointer ${
              currentMode === 'family'
                ? 'bg-emerald-50/70 border-emerald-300 ring-2 ring-emerald-500/20'
                : 'bg-zinc-50/60 hover:bg-zinc-100/60 border-zinc-200'
            }`}
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-bold text-sm text-zinc-900">Family with Children</span>
              {currentMode === 'family' && (
                <span className="text-[10px] font-bold bg-emerald-600 text-white px-2 py-0.5 rounded-full">
                  Active
                </span>
              )}
            </div>
            <p className="text-xs text-zinc-600 leading-relaxed">
              Parent approval queue, kid-friendly checklist, photo & notes proof gates, and gamified reward catalog.
            </p>
          </button>

          {/* Couple Preset */}
          <button
            type="button"
            onClick={() => onUpdateMode('casual')}
            className={`p-4 rounded-2xl border text-left transition-all cursor-pointer ${
              currentMode === 'casual'
                ? 'bg-rose-50/70 border-rose-300 ring-2 ring-rose-500/20'
                : 'bg-zinc-50/60 hover:bg-zinc-100/60 border-zinc-200'
            }`}
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-bold text-sm text-zinc-900">Couples / Casual</span>
              {currentMode === 'casual' && (
                <span className="text-[10px] font-bold bg-rose-600 text-white px-2 py-0.5 rounded-full">
                  Active
                </span>
              )}
            </div>
            <p className="text-xs text-zinc-600 leading-relaxed">
              Low-overhead shared board, voluntary task claiming from open pool, 1-tap completions, and gentle nudges.
            </p>
          </button>
        </div>
      </div>

      {/* Member Roster */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 shadow-xs space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-xs font-bold text-zinc-800 uppercase tracking-wider mb-0.5">
              Household Roster ({members.length})
            </h2>
            <p className="text-xs text-zinc-500">Manage member profiles and assigned roles.</p>
          </div>
          <button
            onClick={() => setIsAddMemberOpen(true)}
            className="flex items-center gap-1.5 text-xs font-bold bg-zinc-900 hover:bg-zinc-800 text-white px-3.5 py-2 rounded-xl transition-all shadow-xs cursor-pointer"
          >
            <UserPlus className="w-3.5 h-3.5" />
            Add Member
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          {members.map((member) => (
            <div
              key={member.id}
              className="p-4 rounded-2xl border border-zinc-200 bg-zinc-50/40 flex flex-col justify-between gap-3"
            >
              <div className="flex items-center gap-3">
                <img
                  src={member.avatar_url}
                  alt={member.name}
                  className="w-11 h-11 rounded-full object-cover border border-zinc-200"
                />
                <div className="min-w-0">
                  <div className="font-bold text-sm text-zinc-900 truncate">{member.name}</div>
                  <span
                    className={`inline-block text-[10px] uppercase font-extrabold px-2 py-0.5 rounded-full border mt-0.5 ${
                      member.role === 'admin'
                        ? 'bg-indigo-50 text-indigo-700 border-indigo-200'
                        : member.role === 'child'
                        ? 'bg-amber-50 text-amber-700 border-amber-200'
                        : 'bg-zinc-100 text-zinc-700 border-zinc-200'
                    }`}
                  >
                    {member.role}
                  </span>
                </div>
              </div>

              <div className="pt-2 border-t border-zinc-100 flex items-center justify-between text-xs text-zinc-500">
                <div className="flex items-center gap-1 font-bold text-amber-700">
                  <Coins className="w-3.5 h-3.5 text-amber-500" />
                  <span>{member.points_balance} pts</span>
                </div>
                <div className="flex items-center gap-1 text-orange-600 font-semibold">
                  <Flame className="w-3 h-3" />
                  <span>{member.streak}d streak</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Reset State Control */}
      <div className="bg-zinc-50 rounded-2xl border border-zinc-200 p-5 flex items-center justify-between">
        <div>
          <div className="text-xs font-bold text-zinc-900 mb-0.5">Reset Demo Data</div>
          <div className="text-xs text-zinc-500">
            Restore initial realistic chores, members, and point balances for testing.
          </div>
        </div>
        <button
          onClick={onResetData}
          className="flex items-center gap-1.5 text-xs font-bold text-zinc-700 hover:text-zinc-900 bg-white hover:bg-zinc-100 px-3 py-2 rounded-xl border border-zinc-300 transition-colors cursor-pointer"
        >
          <RotateCcw className="w-3.5 h-3.5" />
          Reset Demo Store
        </button>
      </div>

      {/* Add Member Modal */}
      {isAddMemberOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="bg-white w-full max-w-sm rounded-3xl shadow-xl border border-zinc-200 p-6 space-y-4">
            <h3 className="text-base font-bold text-zinc-900">Add Household Member</h3>
            <form onSubmit={handleAddMemberSubmit} className="space-y-3">
              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Member Display Name
                </label>
                <input
                  type="text"
                  required
                  value={newMemberName}
                  onChange={(e) => setNewMemberName(e.target.value)}
                  placeholder="e.g. Jordan, Oliver..."
                  className="w-full text-xs font-medium p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Role
                </label>
                <select
                  value={newMemberRole}
                  onChange={(e) => setNewMemberRole(e.target.value as MemberRole)}
                  className="w-full text-xs font-semibold bg-zinc-50 p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                >
                  <option value="admin">Admin / Parent (Full management & approvals)</option>
                  <option value="member">Member / Roommate (Standard peer)</option>
                  <option value="child">Child (Simplified checklist, needs verification)</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsAddMemberOpen(false)}
                  className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-900 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || !newMemberName.trim()}
                  className="px-5 py-2 text-xs font-bold text-white bg-zinc-900 hover:bg-zinc-800 rounded-xl shadow-xs disabled:opacity-50 cursor-pointer"
                >
                  {isSubmitting ? 'Adding...' : 'Add Member'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Change Admin PIN Modal */}
      {isEditPinOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-3xl max-w-sm w-full p-6 shadow-2xl border border-zinc-200 text-left animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between pb-3 border-b border-zinc-100">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-full bg-amber-100 flex items-center justify-center text-amber-700">
                  <Lock className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="font-bold text-sm text-zinc-900">Change Admin PIN</h3>
                  <p className="text-[11px] text-zinc-500">Requires 4 numeric digits</p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setIsEditPinOpen(false)}
                className="text-zinc-400 hover:text-zinc-600 p-1 rounded-lg"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleSavePinSubmit} className="pt-4 space-y-4">
              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
                  New 4-Digit Security PIN
                </label>
                <input
                  id="update-admin-pin-input"
                  type="text"
                  inputMode="numeric"
                  maxLength={4}
                  autoFocus
                  required
                  value={newPin}
                  onChange={(e) => {
                    setNewPin(e.target.value.replace(/\D/g, '').slice(0, 4));
                    setPinError('');
                  }}
                  placeholder="e.g. 5678"
                  className="w-full text-center tracking-widest font-mono text-xl font-bold p-3 rounded-xl border border-zinc-300 focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500"
                />
                <p className="text-[11px] text-zinc-500 mt-1">
                  Used for unlocking shared tablet kiosk mode and approving kid chore submissions.
                </p>
                {pinError && (
                  <p id="admin-pin-error" className="text-xs text-rose-600 font-bold mt-1.5">
                    {pinError}
                  </p>
                )}
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsEditPinOpen(false)}
                  className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-900 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  id="save-admin-pin-btn"
                  type="submit"
                  disabled={isSavingPin || newPin.length !== 4}
                  className="px-5 py-2 text-xs font-bold text-white bg-amber-600 hover:bg-amber-700 rounded-xl shadow-xs disabled:opacity-50 cursor-pointer transition-colors"
                >
                  {isSavingPin ? 'Saving...' : 'Update PIN'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
