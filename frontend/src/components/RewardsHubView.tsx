import React, { useState } from 'react';
import { RewardItem, RewardRedemption, Member } from '../types';
import {
  Gift,
  Coins,
  Flame,
  Award,
  Trophy,
  Coffee,
  Sparkles,
  Tv,
  Moon,
  IceCream,
  Gamepad2,
  UtensilsCrossed,
  HeartHandshake,
  CheckCircle2,
  Plus,
  X,
  Clock,
  ShieldCheck,
} from 'lucide-react';

interface RewardsHubViewProps {
  rewards: RewardItem[];
  redemptions: Array<{ redemption: RewardRedemption; reward: RewardItem; member: Member }>;
  currentMember: Member | null;
  householdMembers: Member[];
  onRedeem: (rewardId: string) => Promise<void>;
  onFulfill: (redemptionId: string) => Promise<void>;
  onCreateReward: (payload: { title: string; description: string; points_cost: number; icon?: string }) => Promise<void>;
}

export const RewardsHubView: React.FC<RewardsHubViewProps> = ({
  rewards,
  redemptions,
  currentMember,
  householdMembers,
  onRedeem,
  onFulfill,
  onCreateReward,
}) => {
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newDesc, setNewDesc] = useState('');
  const [newCost, setNewCost] = useState(50);
  const [newIcon, setNewIcon] = useState('Gift');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [redeemingId, setRedeemingId] = useState<string | null>(null);

  const isAdmin = currentMember?.role === 'admin';
  const pointsBalance = currentMember?.points_balance || 0;

  // Icon mapping
  const renderRewardIcon = (iconName?: string) => {
    switch (iconName) {
      case 'Coffee':
        return <Coffee className="w-5 h-5 text-amber-600" />;
      case 'Sparkles':
        return <Sparkles className="w-5 h-5 text-indigo-600" />;
      case 'Tv':
        return <Tv className="w-5 h-5 text-sky-600" />;
      case 'Moon':
        return <Moon className="w-5 h-5 text-purple-600" />;
      case 'IceCream':
        return <IceCream className="w-5 h-5 text-pink-600" />;
      case 'Gamepad2':
        return <Gamepad2 className="w-5 h-5 text-emerald-600" />;
      case 'UtensilsCrossed':
        return <UtensilsCrossed className="w-5 h-5 text-orange-600" />;
      case 'HeartHandshake':
        return <HeartHandshake className="w-5 h-5 text-rose-600" />;
      default:
        return <Gift className="w-5 h-5 text-amber-500" />;
    }
  };

  // Sort members for leaderboard by total lifetime points earned
  const leaderboard = [...householdMembers].sort(
    (a, b) => b.total_points_earned - a.total_points_earned
  );

  const pendingRedemptions = (redemptions || []).filter((r) => r.redemption.status === 'requested');
  const fulfilledRedemptions = (redemptions || []).filter((r) => r.redemption.status === 'fulfilled');

  const handleRedeemClick = async (rewardId: string) => {
    setRedeemingId(rewardId);
    try {
      await onRedeem(rewardId);
    } finally {
      setRedeemingId(null);
    }
  };

  const handleAddRewardSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim()) return;
    setIsSubmitting(true);
    try {
      await onCreateReward({
        title: newTitle.trim(),
        description: newDesc.trim(),
        points_cost: Number(newCost),
        icon: newIcon,
      });
      setIsAddModalOpen(false);
      setNewTitle('');
      setNewDesc('');
      setNewCost(50);
    } catch (err) {
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Top Banner: Spendable balance & Leaderboard overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Active Member Balance Card */}
        <div className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col justify-between md:col-span-1 shadow-xs">
          <div>
            <div className="flex items-center justify-between text-xs text-zinc-500 font-semibold mb-1">
              <span>Your Spendable Points</span>
              <span className="text-[11px] font-normal text-zinc-400">Persona: {currentMember?.name}</span>
            </div>
            <div className="flex items-baseline gap-2 mt-1">
              <span className="text-3xl font-black text-zinc-900 tracking-tight">{pointsBalance}</span>
              <span className="text-sm font-bold text-amber-600 flex items-center gap-1">
                <Coins className="w-4 h-4" /> pts
              </span>
            </div>
          </div>

          <div className="mt-4 pt-3 border-t border-zinc-100 flex items-center justify-between text-xs">
            <div className="flex items-center gap-1 text-orange-600 font-semibold">
              <Flame className="w-4 h-4" />
              <span>{currentMember?.streak || 0} Day Streak</span>
            </div>
            <div className="text-zinc-500 font-medium">
              Lifetime: <span className="font-bold text-zinc-700">{currentMember?.total_points_earned || 0} pts</span>
            </div>
          </div>
        </div>

        {/* Household Leaderboard Card */}
        <div className="bg-white rounded-2xl border border-zinc-200 p-5 md:col-span-2 shadow-xs">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Trophy className="w-4 h-4 text-amber-500" />
              <h2 className="text-xs font-bold text-zinc-800 uppercase tracking-wider">
                Household Contributors Leaderboard
              </h2>
            </div>
            <span className="text-[11px] text-zinc-400">Based on lifetime task points</span>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
            {leaderboard.map((m, idx) => {
              const isFirst = idx === 0;
              return (
                <div
                  key={m.id}
                  className={`p-3 rounded-xl border text-center transition-all ${
                    isFirst
                      ? 'bg-amber-50/50 border-amber-200/80 shadow-xs'
                      : 'bg-zinc-50/60 border-zinc-200'
                  }`}
                >
                  <div className="relative inline-block mb-1.5">
                    <img
                      src={m.avatar_url}
                      alt={m.name}
                      className="w-10 h-10 rounded-full mx-auto object-cover border border-white shadow-xs"
                    />
                    <span
                      className={`absolute -bottom-1 -right-1 text-[10px] w-4 h-4 rounded-full flex items-center justify-center font-black ${
                        isFirst ? 'bg-amber-500 text-white' : 'bg-zinc-700 text-white'
                      }`}
                    >
                      {idx + 1}
                    </span>
                  </div>
                  <div className="font-bold text-xs text-zinc-900 truncate">{m.name.split(' ')[0]}</div>
                  <div className="text-[11px] font-extrabold text-amber-700 mt-0.5">{m.total_points_earned} pts</div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Admin Pending Redemptions Queue */}
      {isAdmin && pendingRedemptions.length > 0 && (
        <div className="bg-amber-50/70 border border-amber-200 rounded-2xl p-5 shadow-xs">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <ShieldCheck className="w-5 h-5 text-amber-600" />
              <h3 className="text-sm font-bold text-amber-950">
                Pending Perk Redemptions Awaiting Fulfillment ({pendingRedemptions.length})
              </h3>
            </div>
            <span className="text-xs text-amber-700">Admin Action Required</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {pendingRedemptions.map(({ redemption, reward, member }) => (
              <div
                key={redemption.id}
                className="bg-white rounded-xl border border-amber-200/80 p-3.5 flex items-center justify-between gap-3 shadow-xs"
              >
                <div className="flex items-center gap-2.5">
                  <img
                    src={member.avatar_url}
                    alt={member.name}
                    className="w-8 h-8 rounded-full object-cover border border-zinc-200"
                  />
                  <div>
                    <div className="text-xs font-bold text-zinc-900">
                      {member.name} redeemed <span className="text-indigo-600">"{reward.title}"</span>
                    </div>
                    <div className="text-[11px] text-zinc-400">
                      Spent {redemption.points_spent} pts • {new Date(redemption.requested_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </div>
                  </div>
                </div>

                <button
                  id={`fulfill-btn-${redemption.id}`}
                  onClick={() => onFulfill(redemption.id)}
                  className="bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white text-xs font-bold py-1.5 px-3 rounded-lg shadow-xs transition-colors cursor-pointer shrink-0"
                >
                  Mark Fulfilled
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Rewards Catalog */}
      <div>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-xs font-bold text-zinc-700 uppercase tracking-wider">
            Household Reward Catalog ({rewards.length})
          </h2>

          {isAdmin && (
            <button
              onClick={() => setIsAddModalOpen(true)}
              className="flex items-center gap-1 text-xs font-bold text-zinc-900 hover:text-zinc-700 bg-zinc-100 hover:bg-zinc-200/70 px-3 py-1.5 rounded-xl transition-colors cursor-pointer"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Reward Item
            </button>
          )}
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {rewards.map((reward) => {
            const canAfford = pointsBalance >= reward.points_cost;
            const isRedeeming = redeemingId === reward.id;

            return (
              <div
                key={reward.id}
                id={`reward-card-${reward.id}`}
                className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col justify-between gap-4 shadow-xs hover:border-zinc-300 transition-all"
              >
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <div className="w-10 h-10 rounded-xl bg-zinc-100 flex items-center justify-center border border-zinc-200">
                      {renderRewardIcon(reward.icon)}
                    </div>
                    <span className="inline-flex items-center gap-1 text-xs font-extrabold px-2.5 py-1 rounded-full bg-amber-50 text-amber-900 border border-amber-200">
                      <Coins className="w-3.5 h-3.5 text-amber-600" />
                      {reward.points_cost} pts
                    </span>
                  </div>

                  <h3 className="font-bold text-sm text-zinc-900 leading-snug">{reward.title}</h3>
                  <p className="text-xs text-zinc-500 mt-1 leading-relaxed">{reward.description}</p>
                </div>

                <div className="pt-2 border-t border-zinc-100">
                  <button
                    id={`redeem-btn-${reward.id}`}
                    onClick={() => handleRedeemClick(reward.id)}
                    disabled={!canAfford || isRedeeming}
                    className={`w-full py-2 px-3 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                      canAfford
                        ? 'bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 text-white shadow-xs'
                        : 'bg-zinc-100 text-zinc-400 cursor-not-allowed'
                    }`}
                  >
                    {isRedeeming ? (
                      'Redeeming...'
                    ) : canAfford ? (
                      'Redeem Reward'
                    ) : (
                      `Need ${reward.points_cost - pointsBalance} more pts`
                    )}
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Add Custom Reward Modal */}
      {isAddModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
          <div className="bg-white w-full max-w-md rounded-3xl shadow-xl border border-zinc-200 p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
                  <Gift className="w-4 h-4" />
                </div>
                <h3 className="text-base font-bold text-zinc-900">Add Household Reward</h3>
              </div>
              <button onClick={() => setIsAddModalOpen(false)} className="p-1.5 text-zinc-400 hover:text-zinc-700">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleAddRewardSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Reward Title <span className="text-rose-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                  placeholder="e.g. Skip Kitchen Duty, Pick Weekend Dinner..."
                  className="w-full text-xs font-semibold p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Description
                </label>
                <textarea
                  rows={2}
                  value={newDesc}
                  onChange={(e) => setNewDesc(e.target.value)}
                  placeholder="What perk or treat does this reward grant?"
                  className="w-full text-xs p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Points Cost (Effort Required): {newCost} pts
                </label>
                <input
                  type="range"
                  min="20"
                  max="150"
                  step="5"
                  value={newCost}
                  onChange={(e) => setNewCost(Number(e.target.value))}
                  className="w-full accent-zinc-900"
                />
              </div>

              <div>
                <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1">
                  Icon Style
                </label>
                <select
                  value={newIcon}
                  onChange={(e) => setNewIcon(e.target.value)}
                  className="w-full text-xs font-semibold bg-zinc-50 p-2.5 rounded-xl border border-zinc-300 focus:outline-none"
                >
                  <option value="Gift">Gift Box</option>
                  <option value="Coffee">Coffee / Cafe</option>
                  <option value="Sparkles">Sparkles / Clean Pass</option>
                  <option value="Tv">TV / Movie Night</option>
                  <option value="Moon">Late Bedtime</option>
                  <option value="IceCream">Ice Cream / Dessert</option>
                  <option value="Gamepad2">Video / Board Game</option>
                  <option value="UtensilsCrossed">Home Cooked Meal</option>
                  <option value="HeartHandshake">Massage / Favor</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsAddModalOpen(false)}
                  className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-900 rounded-xl"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || !newTitle.trim()}
                  className="px-5 py-2 text-xs font-bold text-white bg-zinc-900 hover:bg-zinc-800 rounded-xl shadow-xs disabled:opacity-50 cursor-pointer"
                >
                  {isSubmitting ? 'Saving...' : 'Add to Catalog'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
