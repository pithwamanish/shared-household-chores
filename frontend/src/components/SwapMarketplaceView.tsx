import React, { useState } from 'react';
import { ChoreSwapRequest, Chore, Member } from '../types';
import {
  Repeat,
  ArrowRightLeft,
  Check,
  X,
  Plus,
  Coins,
  Clock,
  User,
  Users,
  AlertCircle,
  Sparkles,
} from 'lucide-react';

interface SwapMarketplaceViewProps {
  swaps: Array<{ swap: ChoreSwapRequest; chore: Chore; requester: Member; targetMember?: Member | null }>;
  currentMember: Member | null;
  myAssignedChores: Chore[];
  householdMembers: Member[];
  onAcceptSwap: (swapId: string) => Promise<void>;
  onRejectSwap: (swapId: string) => Promise<void>;
  onCancelSwap: (swapId: string) => Promise<void>;
  onRequestSwap: (payload: { chore_id: string; requester_id: string; target_member_id?: string | null; reason?: string }) => Promise<void>;
}

export const SwapMarketplaceView: React.FC<SwapMarketplaceViewProps> = ({
  swaps,
  currentMember,
  myAssignedChores,
  householdMembers,
  onAcceptSwap,
  onRejectSwap,
  onCancelSwap,
  onRequestSwap,
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedChoreId, setSelectedChoreId] = useState('');
  const [targetMemberId, setTargetMemberId] = useState<string>('open_board');
  const [reason, setReason] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const pendingSwaps = swaps.filter((s) => s.swap.status === 'pending');
  const resolvedSwaps = swaps.filter((s) => s.swap.status !== 'pending');

  const eligiblePeers = householdMembers.filter((m) => m.id !== currentMember?.id);

  const handleOpenModal = (defaultChoreId?: string) => {
    setSelectedChoreId(defaultChoreId || myAssignedChores[0]?.id || '');
    setTargetMemberId('open_board');
    setReason('');
    setIsModalOpen(true);
  };

  const handleCreateSwapSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedChoreId || !currentMember) return;
    setIsSubmitting(true);
    try {
      await onRequestSwap({
        chore_id: selectedChoreId,
        requester_id: currentMember.id,
        target_member_id: targetMemberId === 'open_board' ? null : targetMemberId,
        reason: reason.trim() || undefined,
      });
      setIsModalOpen(false);
    } catch (err) {
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <div className="w-8 h-8 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold">
              <Repeat className="w-4 h-4" />
            </div>
            <h1 className="text-lg font-bold text-zinc-900 tracking-tight">Chore Trading & Swapping Board</h1>
          </div>
          <p className="text-xs text-zinc-500">
            Busy with exams or working late? Propose a trade with a housemate or post to the shared swap pool.
          </p>
        </div>

        <button
          id="propose-swap-btn"
          onClick={() => handleOpenModal()}
          disabled={myAssignedChores.length === 0}
          className="flex items-center justify-center gap-1.5 bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 text-white text-xs font-bold py-2.5 px-4 rounded-xl transition-all shadow-xs disabled:opacity-50 cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          Propose Chore Swap
        </button>
      </div>

      {/* Active Swap Proposals */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xs font-bold text-zinc-700 uppercase tracking-wider">
            Open & Pending Swap Offers ({pendingSwaps.length})
          </h2>
        </div>

        {pendingSwaps.length === 0 ? (
          <div className="bg-white rounded-2xl border border-zinc-200 p-10 text-center">
            <ArrowRightLeft className="w-8 h-8 text-zinc-300 mx-auto mb-2" />
            <div className="text-sm font-bold text-zinc-900">No active swaps right now</div>
            <p className="text-xs text-zinc-500 max-w-sm mx-auto mt-1">
              Have a task you need help covering this week? Propose a swap and a housemate can take over your turn!
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {pendingSwaps.map(({ swap, chore, requester, targetMember }) => {
              const isMine = currentMember && requester.id === currentMember.id;
              const isTargetedToMe = currentMember && targetMember && targetMember.id === currentMember.id;
              const isOpenToAll = !targetMember;

              return (
                <div
                  key={swap.id}
                  id={`swap-card-${swap.id}`}
                  className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col justify-between gap-4 shadow-xs"
                >
                  <div>
                    {/* Top Requester / Audience */}
                    <div className="flex items-center justify-between gap-2 mb-3">
                      <div className="flex items-center gap-2">
                        <img
                          src={requester.avatar_url}
                          alt={requester.name}
                          className="w-8 h-8 rounded-full object-cover border border-zinc-200"
                        />
                        <div>
                          <div className="text-xs font-bold text-zinc-900">
                            {requester.name} {isMine && '(You)'}
                          </div>
                          <div className="text-[11px] text-zinc-400">
                            {isOpenToAll ? (
                              <span className="text-indigo-600 font-medium">Offered to Entire Household</span>
                            ) : (
                              <span>Offered to {targetMember?.name}</span>
                            )}
                          </div>
                        </div>
                      </div>

                      <span className="inline-flex items-center gap-1 text-xs font-bold px-2 py-0.5 rounded-full bg-amber-50 text-amber-900 border border-amber-200">
                        <Coins className="w-3 h-3 text-amber-600" />
                        +{chore.effort_points} pts
                      </span>
                    </div>

                    {/* Chore box */}
                    <div className="p-3 bg-zinc-50 rounded-xl border border-zinc-200 mb-3">
                      <div className="font-bold text-sm text-zinc-900">{chore.title}</div>
                      <div className="text-[11px] text-zinc-500 mt-0.5 capitalize">
                        Category: {chore.category}
                      </div>
                    </div>

                    {/* Swap Reason */}
                    {swap.reason && (
                      <div className="p-2.5 bg-indigo-50/40 rounded-xl border border-indigo-100 text-xs text-zinc-800 leading-relaxed italic mb-2">
                        "{swap.reason}"
                      </div>
                    )}
                  </div>

                  {/* Actions */}
                  <div className="pt-2 border-t border-zinc-100 flex items-center gap-2">
                    {isMine ? (
                      <button
                        onClick={() => onCancelSwap(swap.id)}
                        className="w-full py-2 px-3 text-xs font-semibold text-rose-600 hover:bg-rose-50 rounded-xl transition-colors cursor-pointer border border-rose-200"
                      >
                        Cancel Proposal
                      </button>
                    ) : (
                      <>
                        <button
                          id={`accept-swap-${swap.id}`}
                          onClick={() => onAcceptSwap(swap.id)}
                          className="flex-1 flex items-center justify-center gap-1.5 bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 text-white text-xs font-bold py-2.5 px-3 rounded-xl transition-all shadow-xs cursor-pointer"
                        >
                          <Check className="w-4 h-4" />
                          Accept Trade & Take Chore
                        </button>
                        {isTargetedToMe && (
                          <button
                            onClick={() => onRejectSwap(swap.id)}
                            className="p-2.5 text-zinc-500 hover:text-zinc-900 hover:bg-zinc-100 rounded-xl transition-colors cursor-pointer"
                            title="Decline trade"
                          >
                            <X className="w-4 h-4" />
                          </button>
                        )}
                      </>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Resolved Swaps History */}
      {resolvedSwaps.length > 0 && (
        <div className="pt-4 border-t border-zinc-200">
          <h2 className="text-xs font-bold text-zinc-700 uppercase tracking-wider mb-3">
            Recent Resolved Trades
          </h2>
          <div className="space-y-2">
            {resolvedSwaps.slice(0, 4).map(({ swap, chore, requester }) => (
              <div
                key={swap.id}
                className="bg-white rounded-xl border border-zinc-200 px-4 py-2.5 flex items-center justify-between text-xs"
              >
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-zinc-900">{chore.title}</span>
                  <span className="text-zinc-400">•</span>
                  <span className="text-zinc-500">Requested by {requester.name}</span>
                </div>
                <span
                  className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase ${
                    swap.status === 'accepted'
                      ? 'bg-emerald-100 text-emerald-800'
                      : 'bg-zinc-100 text-zinc-600'
                  }`}
                >
                  {swap.status}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Propose Swap Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="bg-white w-full max-w-md rounded-3xl shadow-xl border border-zinc-200 p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
                  <Repeat className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-zinc-900">Propose Chore Swap</h3>
                  <p className="text-xs text-zinc-500">Offer an assigned task to a peer.</p>
                </div>
              </div>
              <button
                onClick={() => setIsModalOpen(false)}
                className="p-1.5 text-zinc-400 hover:text-zinc-700 rounded-xl"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateSwapSubmit} className="space-y-4">
              {/* Chore selector */}
              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Select Chore to Swap:
                </label>
                <select
                  value={selectedChoreId}
                  onChange={(e) => setSelectedChoreId(e.target.value)}
                  className="w-full text-xs font-semibold bg-zinc-50 p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                >
                  {myAssignedChores.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.title} (+{c.effort_points} pts)
                    </option>
                  ))}
                </select>
              </div>

              {/* Target recipient */}
              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Offer To:
                </label>
                <select
                  value={targetMemberId}
                  onChange={(e) => setTargetMemberId(e.target.value)}
                  className="w-full text-xs font-semibold bg-zinc-50 p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                >
                  <option value="open_board">Entire Household (Open Swap Board)</option>
                  {eligiblePeers.map((m) => (
                    <option key={m.id} value={m.id}>
                      Specific Peer: {m.name} ({m.role})
                    </option>
                  ))}
                </select>
              </div>

              {/* Reason */}
              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Reason or Message (Optional):
                </label>
                <input
                  type="text"
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="e.g. Working double shift, studying for finals..."
                  className="w-full text-xs p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-900 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || !selectedChoreId}
                  className="px-5 py-2 text-xs font-bold text-white bg-zinc-900 hover:bg-zinc-800 rounded-xl shadow-xs disabled:opacity-50 cursor-pointer"
                >
                  {isSubmitting ? 'Posting...' : 'Post Swap Offer'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
