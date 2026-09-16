import React, { useState } from 'react';
import { Home, X, Users, Copy, Check, LogOut, ShieldCheck, Sparkles } from 'lucide-react';
import { Household, Member } from '../types';

interface HouseholdInfoModalProps {
  isOpen: boolean;
  onClose: () => void;
  household: Household | null;
  members: Member[];
  currentMember: Member | null;
  onSignOut: () => void;
}

export const HouseholdInfoModal: React.FC<HouseholdInfoModalProps> = ({
  isOpen,
  onClose,
  household,
  members,
  currentMember,
  onSignOut,
}) => {
  const [copied, setCopied] = useState(false);

  if (!isOpen || !household) return null;

  const inviteCode = household.invite_code || 'CHORE-SYNC';
  const inviteLink = `${window.location.origin}?invite=${encodeURIComponent(inviteCode)}`;

  const handleCopyLink = () => {
    navigator.clipboard.writeText(inviteLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2500);
  };

  const modeBadge = {
    flatmate: {
      label: 'Flatmate Mode',
      desc: 'Peer-to-peer accountability, round-robin auto-rotation, and chore swaps marketplace.',
      color: 'bg-indigo-50 text-indigo-700 border-indigo-200',
    },
    family: {
      label: 'Family Mode',
      desc: 'Parent approval gates, chore assignments, photo proof verification, and points redemption.',
      color: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    },
    casual: {
      label: 'Couple / Casual Mode',
      desc: 'Shared task backlog, voluntary claiming, one-tap completions, and gentle nudges.',
      color: 'bg-rose-50 text-rose-700 border-rose-200',
    },
  }[household.settings.default_mode] || {
    label: household.settings.default_mode,
    desc: 'Household chores coordination workspace',
    color: 'bg-zinc-50 text-zinc-700 border-zinc-200',
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div
        id="household-info-modal"
        className="bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl border border-zinc-200 text-left animate-in fade-in zoom-in-95 duration-150"
      >
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-zinc-100">
          <div className="flex items-center gap-3">
            <div className="w-11 h-11 rounded-2xl bg-indigo-50 border border-indigo-100 text-indigo-600 flex items-center justify-center font-bold shadow-xs">
              <Home className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-base font-black text-zinc-900 tracking-tight">{household.name}</h3>
                <span className={`text-[10px] font-bold uppercase px-2 py-0.5 rounded-full border ${modeBadge.color}`}>
                  {modeBadge.label}
                </span>
              </div>
              <p className="text-xs text-zinc-500 mt-0.5">{modeBadge.desc}</p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-zinc-400 hover:text-zinc-600 p-1.5 rounded-xl hover:bg-zinc-100 transition-colors cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Invite Code Box */}
        <div className="my-5 p-4 rounded-2xl bg-indigo-50/60 border border-indigo-100">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-bold uppercase tracking-wider text-indigo-900 flex items-center gap-1.5">
              <Sparkles className="w-3.5 h-3.5 text-indigo-600" />
              <span>Invite Household Members</span>
            </span>
            <span className="text-[11px] font-medium text-indigo-600">Active Invite Code</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex-1 bg-white border border-indigo-200 rounded-xl px-3.5 py-2 font-mono font-bold text-sm text-indigo-950 select-all shadow-2xs tracking-wider">
              {inviteCode}
            </div>
            <button
              type="button"
              id="copy-invite-link-btn"
              onClick={handleCopyLink}
              className="flex items-center gap-1.5 bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-xs font-bold px-3.5 py-2.5 rounded-xl transition-all cursor-pointer shadow-xs shrink-0"
            >
              {copied ? (
                <>
                  <Check className="w-4 h-4 text-emerald-300" />
                  <span>Copied!</span>
                </>
              ) : (
                <>
                  <Copy className="w-4 h-4" />
                  <span>Copy Link</span>
                </>
              )}
            </button>
          </div>
          <p className="text-[11px] text-indigo-700/80 mt-2 leading-relaxed">
            Share this code or link with your flatmates or family members. When they register, they'll instantly join <strong className="font-semibold text-indigo-950">{household.name}</strong>.
          </p>
        </div>

        {/* Members Roster */}
        <div className="mb-5">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-bold uppercase tracking-wider text-zinc-500 flex items-center gap-1.5">
              <Users className="w-3.5 h-3.5 text-zinc-400" />
              <span>Household Roster ({members.length})</span>
            </span>
          </div>
          <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
            {members.map((m) => {
              const isYou = m.id === currentMember?.id;
              return (
                <div
                  key={m.id}
                  className={`flex items-center justify-between p-2.5 rounded-xl border ${
                    isYou
                      ? 'bg-zinc-50 border-zinc-300 shadow-2xs'
                      : 'bg-white border-zinc-200/80'
                  }`}
                >
                  <div className="flex items-center gap-2.5 min-w-0">
                    {m.avatar_url ? (
                      <img
                        src={m.avatar_url}
                        alt={m.name}
                        className="w-7 h-7 rounded-full object-cover shrink-0 border border-zinc-200"
                      />
                    ) : (
                      <div className="w-7 h-7 rounded-full bg-zinc-200 flex items-center justify-center text-xs font-bold shrink-0 text-zinc-700">
                        {m.name[0]}
                      </div>
                    )}
                    <div className="min-w-0">
                      <div className="text-xs font-bold text-zinc-900 flex items-center gap-1.5">
                        <span className="truncate">{m.name}</span>
                        {isYou && (
                          <span className="text-[9px] bg-indigo-100 text-indigo-800 font-bold px-1.5 py-0.2 rounded">
                            You
                          </span>
                        )}
                      </div>
                      <div className="text-[11px] text-zinc-500 capitalize">{m.role}</div>
                    </div>
                  </div>
                  <div className="text-xs font-semibold text-amber-700 shrink-0">
                    {m.points_balance} pts
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Footer & Multi-tenant note */}
        <div className="pt-4 border-t border-zinc-100 flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="text-[11px] text-zinc-500 text-center sm:text-left">
            <span>To access a different household, sign out first.</span>
          </div>
          <div className="flex items-center gap-2 w-full sm:w-auto">
            <button
              type="button"
              id="household-sign-out-btn"
              onClick={() => {
                onClose();
                onSignOut();
              }}
              className="flex-1 sm:flex-initial flex items-center justify-center gap-1.5 text-xs font-bold px-3 py-2 rounded-xl text-rose-600 hover:bg-rose-50 border border-rose-200 transition-colors cursor-pointer"
            >
              <LogOut className="w-3.5 h-3.5" />
              <span>Sign Out</span>
            </button>
            <button
              type="button"
              onClick={onClose}
              className="flex-1 sm:flex-initial px-4 py-2 rounded-xl bg-zinc-900 hover:bg-zinc-800 text-white text-xs font-bold transition-colors cursor-pointer"
            >
              Done
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
