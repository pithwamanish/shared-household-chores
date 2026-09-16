import React, { useState, useRef, useEffect, useCallback } from 'react';
import { ActivityLog, Member } from '../types';
import {
  History,
  CheckCircle2,
  Repeat,
  BellRing,
  Coins,
  ShieldCheck,
  XCircle,
  Clock,
  PlusCircle,
  Filter,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';

interface ActivityFeedViewProps {
  activities: ActivityLog[];
  members: Member[];
}

export const ActivityFeedView: React.FC<ActivityFeedViewProps> = ({ activities, members }) => {
  const [filter, setFilter] = useState<'all' | 'completions' | 'swaps' | 'nudges'>('all');
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(false);

  const checkScroll = useCallback(() => {
    const el = scrollRef.current;
    if (el) {
      setCanScrollLeft(el.scrollLeft > 2);
      setCanScrollRight(el.scrollLeft < el.scrollWidth - el.clientWidth - 2);
    }
  }, []);

  useEffect(() => {
    const el = scrollRef.current;
    if (el) {
      checkScroll();
      el.addEventListener('scroll', checkScroll);
      window.addEventListener('resize', checkScroll);
      return () => {
        el.removeEventListener('scroll', checkScroll);
        window.removeEventListener('resize', checkScroll);
      };
    }
  }, [checkScroll]);

  const handleScroll = (direction: 'left' | 'right') => {
    if (scrollRef.current) {
      scrollRef.current.scrollBy({
        left: direction === 'left' ? -180 : 180,
        behavior: 'smooth',
      });
    }
  };

  const filtered = activities.filter((act) => {
    if (filter === 'completions') {
      return (
        act.event_type === 'chore_completed' ||
        act.event_type === 'chore_approved' ||
        act.event_type === 'approval_requested'
      );
    }
    if (filter === 'swaps') {
      return act.event_type === 'swap_proposed' || act.event_type === 'swap_accepted';
    }
    if (filter === 'nudges') {
      return act.event_type === 'chore_nudge';
    }
    return true;
  });

  const getEventBadge = (type: ActivityLog['event_type']) => {
    switch (type) {
      case 'chore_completed':
        return { icon: CheckCircle2, bg: 'bg-emerald-50 text-emerald-700 border-emerald-200', label: 'Done' };
      case 'chore_approved':
        return { icon: ShieldCheck, bg: 'bg-emerald-50 text-emerald-700 border-emerald-200', label: 'Approved' };
      case 'approval_requested':
        return { icon: Clock, bg: 'bg-amber-50 text-amber-700 border-amber-200', label: 'Review' };
      case 'chore_rejected':
        return { icon: XCircle, bg: 'bg-rose-50 text-rose-700 border-rose-200', label: 'Rework' };
      case 'swap_proposed':
      case 'swap_accepted':
        return { icon: Repeat, bg: 'bg-indigo-50 text-indigo-700 border-indigo-200', label: 'Swap' };
      case 'chore_nudge':
        return { icon: BellRing, bg: 'bg-purple-50 text-purple-700 border-purple-200', label: 'Nudge' };
      case 'reward_redeemed':
      case 'reward_fulfilled':
        return { icon: Coins, bg: 'bg-amber-50 text-amber-700 border-amber-200', label: 'Reward' };
      default:
        return { icon: PlusCircle, bg: 'bg-zinc-50 text-zinc-700 border-zinc-200', label: 'Chore' };
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <div className="w-8 h-8 rounded-xl bg-zinc-100 text-zinc-800 flex items-center justify-center font-bold">
              <History className="w-4 h-4" />
            </div>
            <h1 className="text-lg font-bold text-zinc-900 tracking-tight">Household Activity & Audit Trail</h1>
          </div>
          <p className="text-xs text-zinc-500">
            Transparent event log tracking task completions, verifications, peer swaps, and gentle reminders.
          </p>
        </div>

        {/* Filter Pills with Left/Right Arrow Navigation */}
        <div className="relative flex items-center gap-1.5">
          <button
            type="button"
            id="activity-scroll-left"
            onClick={() => handleScroll('left')}
            disabled={!canScrollLeft}
            className="p-1.5 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-600 hover:text-zinc-900 disabled:opacity-25 disabled:pointer-events-none transition-all shrink-0 cursor-pointer shadow-2xs"
            aria-label="Previous filters"
            title="Previous"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>

          <div
            ref={scrollRef}
            className="flex-1 flex items-center gap-1.5 overflow-x-auto no-scrollbar scroll-smooth pb-1 sm:pb-0"
          >
            <button
              onClick={() => setFilter('all')}
              className={`px-3 py-1.5 text-xs font-bold rounded-xl transition-colors cursor-pointer shrink-0 ${
                filter === 'all'
                  ? 'bg-zinc-900 text-white'
                  : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
              }`}
            >
              All Events
            </button>
            <button
              onClick={() => setFilter('completions')}
              className={`px-3 py-1.5 text-xs font-bold rounded-xl transition-colors cursor-pointer shrink-0 ${
                filter === 'completions'
                  ? 'bg-zinc-900 text-white'
                  : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
              }`}
            >
              Completions & Approvals
            </button>
            <button
              onClick={() => setFilter('swaps')}
              className={`px-3 py-1.5 text-xs font-bold rounded-xl transition-colors cursor-pointer shrink-0 ${
                filter === 'swaps'
                  ? 'bg-zinc-900 text-white'
                  : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
              }`}
            >
              Swaps & Trades
            </button>
            <button
              onClick={() => setFilter('nudges')}
              className={`px-3 py-1.5 text-xs font-bold rounded-xl transition-colors cursor-pointer shrink-0 ${
                filter === 'nudges'
                  ? 'bg-zinc-900 text-white'
                  : 'bg-zinc-100 hover:bg-zinc-200/70 text-zinc-700'
              }`}
            >
              Nudges
            </button>
          </div>

          <button
            type="button"
            id="activity-scroll-right"
            onClick={() => handleScroll('right')}
            disabled={!canScrollRight}
            className="p-1.5 rounded-xl bg-zinc-100 hover:bg-zinc-200 text-zinc-600 hover:text-zinc-900 disabled:opacity-25 disabled:pointer-events-none transition-all shrink-0 cursor-pointer shadow-2xs"
            aria-label="Next filters"
            title="Next"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Events Timeline */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 shadow-xs">
        {filtered.length === 0 ? (
          <div className="text-center py-12 text-zinc-400 text-xs">
            No events match the selected filter.
          </div>
        ) : (
          <div className="space-y-4">
            {filtered.map((act) => {
              const actor = members.find((m) => m.id === act.actor_id);
              const { icon: Icon, bg, label } = getEventBadge(act.event_type);
              const date = new Date(act.created_at);

              return (
                <div
                  key={act.id}
                  className="flex items-start gap-3.5 pb-4 border-b border-zinc-100 last:border-0 last:pb-0"
                >
                  {/* Actor Avatar or Icon */}
                  <div className="relative shrink-0 mt-0.5">
                    {actor?.avatar_url ? (
                      <img
                        src={actor.avatar_url}
                        alt={actor.name}
                        className="w-8 h-8 rounded-full object-cover border border-zinc-200"
                      />
                    ) : (
                      <div className="w-8 h-8 rounded-full bg-zinc-200 flex items-center justify-center font-bold text-xs">
                        {actor?.name[0] || '?'}
                      </div>
                    )}
                  </div>

                  {/* Message & Context */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 flex-wrap mb-1">
                      <span className={`inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full border ${bg}`}>
                        <Icon className="w-3 h-3" />
                        {label}
                      </span>
                      <span className="text-[11px] text-zinc-400">
                        {date.toLocaleDateString([], { month: 'short', day: 'numeric' })} at{' '}
                        {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                    <div className="text-xs text-zinc-800 font-medium leading-relaxed">
                      {act.message}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};
