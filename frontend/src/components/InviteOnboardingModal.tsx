import React, { useState } from 'react';
import { Household, MemberRole } from '../types';
import { Sparkles, Home, UserPlus, Check, ShieldCheck } from 'lucide-react';

interface InviteOnboardingModalProps {
  household: Household;
  isOpen: boolean;
  onJoin: (name: string, role: MemberRole) => Promise<void>;
  onClose: () => void;
}

export const InviteOnboardingModal: React.FC<InviteOnboardingModalProps> = ({
  household,
  isOpen,
  onJoin,
  onClose,
}) => {
  const [name, setName] = useState('');
  const [role, setRole] = useState<MemberRole>(
    household.settings.default_mode === 'family' ? 'child' : 'member'
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsSubmitting(true);
    try {
      await onJoin(name.trim(), role);
      onClose();
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-zinc-200 animate-in fade-in zoom-in-95 duration-150">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="w-14 h-14 rounded-2xl bg-indigo-50 border border-indigo-100 text-indigo-600 flex items-center justify-center mx-auto mb-3 shadow-xs">
            <Home className="w-7 h-7" />
          </div>
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 mb-2">
            <Sparkles className="w-3.5 h-3.5" />
            Invite Link Accepted
          </span>
          <h2 className="text-xl font-bold text-zinc-900 tracking-tight">
            Join {household.name}
          </h2>
          <p className="text-xs text-zinc-500 mt-1 max-w-xs mx-auto">
            You've been invited to coordinate chores, track points, and share responsibilities.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
              Your Name
            </label>
            <input
              type="text"
              required
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Jordan"
              className="w-full px-3.5 py-2.5 bg-zinc-50 border border-zinc-200 rounded-xl text-sm font-medium text-zinc-800 placeholder:text-zinc-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all"
            />
          </div>

          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
              Role in Household
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setRole('member')}
                className={`flex items-center justify-center gap-2 py-2 px-3 rounded-xl border text-xs font-semibold transition-all cursor-pointer ${
                  role === 'member'
                    ? 'bg-zinc-900 text-white border-zinc-900 shadow-xs'
                    : 'bg-zinc-50 text-zinc-600 border-zinc-200 hover:bg-zinc-100'
                }`}
              >
                <UserPlus className="w-3.5 h-3.5" />
                <span>Roommate / Partner</span>
              </button>
              <button
                type="button"
                onClick={() => setRole('child')}
                className={`flex items-center justify-center gap-2 py-2 px-3 rounded-xl border text-xs font-semibold transition-all cursor-pointer ${
                  role === 'child'
                    ? 'bg-zinc-900 text-white border-zinc-900 shadow-xs'
                    : 'bg-zinc-50 text-zinc-600 border-zinc-200 hover:bg-zinc-100'
                }`}
              >
                <Sparkles className="w-3.5 h-3.5" />
                <span>Child / Kid</span>
              </button>
            </div>
          </div>

          <div className="pt-2 flex items-center justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-800 rounded-xl hover:bg-zinc-100 transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || !name.trim()}
              className="flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-bold rounded-xl shadow-xs transition-all cursor-pointer"
            >
              {isSubmitting ? (
                <span>Joining...</span>
              ) : (
                <>
                  <Check className="w-4 h-4" />
                  <span>Join & Start Syncing</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
