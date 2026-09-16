import React, { useState } from 'react';
import { Chore, Member, ChoreCategory } from '../types';
import {
  Sparkles,
  Utensils,
  Trees,
  Dog,
  Wrench,
  Check,
  CheckCircle2,
  Clock,
  RotateCw,
  MoreVertical,
  Repeat,
  BellRing,
  MessageSquare,
  Trash2,
  Edit2,
  AlertTriangle,
  UserPlus,
  ShieldAlert,
  Camera,
} from 'lucide-react';
import confetti from 'canvas-confetti';

interface ChoreCardProps {
  chore: Chore;
  assignee?: Member | null;
  currentMember: Member | null;
  onComplete: (chore: Chore) => void;
  onClaim: (choreId: string) => void;
  onRequestSwap: (chore: Chore) => void;
  onSendNudge: (chore: Chore) => void;
  onOpenDetails: (chore: Chore) => void;
  onEdit: (chore: Chore) => void;
  onDelete: (choreId: string) => void;
}

export const ChoreCard: React.FC<ChoreCardProps> = ({
  chore,
  assignee,
  currentMember,
  onComplete,
  onClaim,
  onRequestSwap,
  onSendNudge,
  onOpenDetails,
  onEdit,
  onDelete,
}) => {
  const [menuOpen, setMenuOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const getCategoryMeta = (cat: ChoreCategory) => {
    switch (cat) {
      case 'cleaning':
        return { icon: Sparkles, color: 'text-sky-600 bg-sky-50 border-sky-200' };
      case 'kitchen':
        return { icon: Utensils, color: 'text-amber-600 bg-amber-50 border-amber-200' };
      case 'yard':
        return { icon: Trees, color: 'text-emerald-600 bg-emerald-50 border-emerald-200' };
      case 'pets':
        return { icon: Dog, color: 'text-orange-600 bg-orange-50 border-orange-200' };
      case 'maintenance':
        return { icon: Wrench, color: 'text-indigo-600 bg-indigo-50 border-indigo-200' };
      default:
        return { icon: Sparkles, color: 'text-zinc-600 bg-zinc-50 border-zinc-200' };
    }
  };

  const { icon: CategoryIcon, color: catColor } = getCategoryMeta(chore.category);

  const isCompleted = chore.status === 'completed';
  const isPendingApproval = chore.status === 'pending_approval';
  const isOverdue = chore.status === 'overdue';
  const isOpenPool = chore.assignment_type === 'open_pool' && !chore.current_assignee_id;
  const isAssignedToMe = currentMember && chore.current_assignee_id === currentMember.id;

  // Format due date relative label
  const formatDueDate = (isoString?: string | null) => {
    if (!isoString) return null;
    const due = new Date(isoString);
    const now = new Date();
    const diffMs = due.getTime() - now.getTime();
    const diffHours = Math.round(diffMs / 3600000);
    const diffDays = Math.round(diffMs / 86400000);

    if (diffMs < 0) {
      const pastDays = Math.abs(diffDays);
      return pastDays <= 0 ? 'Overdue today' : `Overdue by ${pastDays}d`;
    }
    if (diffHours < 24) {
      return diffHours <= 1 ? 'Due within 1h' : `Due in ${diffHours}h`;
    }
    if (diffDays === 1) return 'Due tomorrow';
    return `Due in ${diffDays}d`;
  };

  const dueLabel = formatDueDate(chore.due_date);

  const handleCompleteClick = () => {
    if (isCompleted || isPendingApproval) return;

    if (!chore.requires_approval && !chore.requires_proof) {
      // Instant confetti celebration
      try {
        confetti({
          particleCount: 50,
          spread: 60,
          origin: { y: 0.7 },
        });
      } catch (e) {
        // ignore in test or headless
      }
    }
    onComplete(chore);
  };

  return (
    <div
      id={`chore-card-${chore.id}`}
      className={`relative bg-white rounded-2xl border transition-all duration-200 p-4 sm:p-5 flex flex-col justify-between gap-4 ${
        isCompleted
          ? 'border-emerald-200/80 bg-emerald-50/20 opacity-85'
          : isPendingApproval
          ? 'border-amber-200 bg-amber-50/20'
          : isOverdue
          ? 'border-rose-200 bg-rose-50/15'
          : 'border-zinc-200/90 hover:border-zinc-300 hover:shadow-sm'
      }`}
    >
      {/* Top row: Category, Recurrence, Points, Options Menu */}
      <div>
        <div className="flex items-center justify-between gap-2 mb-2.5">
          <div className="flex items-center gap-1.5 flex-wrap">
            {/* Category badge */}
            <span
              className={`inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-lg border ${catColor}`}
            >
              <CategoryIcon className="w-3.5 h-3.5" />
              <span className="capitalize">{chore.category}</span>
            </span>

            {/* Recurrence badge */}
            {chore.recurrence_type !== 'none' && (
              <span className="inline-flex items-center gap-1 text-[11px] font-medium text-zinc-600 bg-zinc-100 px-2 py-0.5 rounded-md border border-zinc-200">
                <RotateCw className="w-3 h-3 text-zinc-500" />
                <span className="capitalize">{chore.recurrence_type}</span>
              </span>
            )}

            {/* Round robin badge */}
            {chore.assignment_type === 'round_robin' && (
              <span className="inline-flex items-center gap-1 text-[11px] font-medium text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded-md border border-indigo-200">
                <Repeat className="w-3 h-3" />
                <span>Rotation</span>
              </span>
            )}

            {/* Verification required flag */}
            {chore.requires_approval && (
              <span
                className="inline-flex items-center gap-0.5 text-[11px] font-medium text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded border border-amber-200"
                title="Requires Parent/Admin Verification"
              >
                <ShieldAlert className="w-3 h-3" />
                <span>Approval</span>
              </span>
            )}

            {/* Photo proof required flag */}
            {chore.requires_proof && (
              <span
                className="inline-flex items-center gap-0.5 text-[11px] font-medium text-blue-700 bg-blue-50 px-1.5 py-0.5 rounded border border-blue-200"
                title="Photo or note proof required"
              >
                <Camera className="w-3 h-3" />
                <span>Proof</span>
              </span>
            )}
          </div>

          <div className="flex items-center gap-1.5">
            {/* Points pill */}
            <span className="inline-flex items-center text-xs font-bold px-2 py-0.5 rounded-full bg-amber-100/80 text-amber-900 border border-amber-200">
              +{chore.effort_points} pts
            </span>

            {/* Actions dropdown */}
            <div className="relative">
              <button
                id={`chore-menu-btn-${chore.id}`}
                onClick={() => setMenuOpen(!menuOpen)}
                className="p-1 text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 rounded-lg transition-colors cursor-pointer"
                aria-label="Chore actions"
              >
                <MoreVertical className="w-4 h-4" />
              </button>

              {menuOpen && (
                <>
                  <div
                    className="fixed inset-0 z-20"
                    onClick={() => setMenuOpen(false)}
                  />
                  <div className="absolute right-0 top-7 z-30 w-48 bg-white rounded-xl shadow-lg border border-zinc-200 py-1 text-xs font-medium text-zinc-700 animate-in fade-in zoom-in-95 duration-150">
                    <button
                      onClick={() => {
                        setMenuOpen(false);
                        onOpenDetails(chore);
                      }}
                      className="w-full text-left px-3 py-2 hover:bg-zinc-50 flex items-center gap-2 cursor-pointer"
                    >
                      <MessageSquare className="w-3.5 h-3.5 text-zinc-400" />
                      View Comments & Info
                    </button>

                    {!isCompleted && !isOpenPool && (
                      <button
                        onClick={() => {
                          setMenuOpen(false);
                          onRequestSwap(chore);
                        }}
                        className="w-full text-left px-3 py-2 hover:bg-zinc-50 flex items-center gap-2 cursor-pointer"
                      >
                        <Repeat className="w-3.5 h-3.5 text-indigo-500" />
                        Request Swap / Trade
                      </button>
                    )}

                    {!isCompleted && chore.current_assignee_id && (
                      <button
                        id={`nudge-chore-btn-${chore.id}`}
                        onClick={() => {
                          setMenuOpen(false);
                          onSendNudge(chore);
                        }}
                        className="w-full text-left px-3 py-2 hover:bg-zinc-50 flex items-center gap-2 cursor-pointer"
                      >
                        <BellRing className="w-3.5 h-3.5 text-amber-500" />
                        Send Friendly Nudge
                      </button>
                    )}

                    <button
                      onClick={() => {
                        setMenuOpen(false);
                        onEdit(chore);
                      }}
                      className="w-full text-left px-3 py-2 hover:bg-zinc-50 flex items-center gap-2 cursor-pointer"
                    >
                      <Edit2 className="w-3.5 h-3.5 text-zinc-400" />
                      Edit Chore
                    </button>

                    <div className="border-t border-zinc-100 my-1" />

                    <button
                      onClick={() => {
                        setMenuOpen(false);
                        onDelete(chore.id);
                      }}
                      className="w-full text-left px-3 py-2 hover:bg-rose-50 text-rose-600 flex items-center gap-2 cursor-pointer"
                    >
                      <Trash2 className="w-3.5 h-3.5 text-rose-500" />
                      Delete Chore
                    </button>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Chore Title */}
        <h3
          className={`font-bold text-base leading-snug tracking-tight cursor-pointer hover:text-indigo-600 transition-colors ${
            isCompleted ? 'line-through text-zinc-400' : 'text-zinc-900'
          }`}
          onClick={() => onOpenDetails(chore)}
        >
          {chore.title}
        </h3>

        {/* Description snippet */}
        {chore.description && (
          <p className="mt-1 text-xs text-zinc-500 leading-relaxed line-clamp-2">
            {chore.description}
          </p>
        )}
      </div>

      {/* Due date & Status pill */}
      <div className="flex items-center justify-between text-xs text-zinc-500 pt-1 border-t border-zinc-100">
        <div className="flex items-center gap-1.5 font-medium">
          {isOverdue ? (
            <span className="inline-flex items-center gap-1 text-rose-700 bg-rose-50 px-2 py-0.5 rounded-md font-semibold border border-rose-200">
              <AlertTriangle className="w-3 h-3 text-rose-600" />
              {dueLabel || 'Overdue'}
            </span>
          ) : isPendingApproval ? (
            <span className="inline-flex items-center gap-1 text-amber-800 bg-amber-50 px-2 py-0.5 rounded-md font-semibold border border-amber-200">
              <Clock className="w-3 h-3 text-amber-600" />
              Pending Parent Review
            </span>
          ) : isCompleted ? (
            <span className="inline-flex items-center gap-1 text-emerald-800 bg-emerald-50 px-2 py-0.5 rounded-md font-semibold border border-emerald-200">
              <CheckCircle2 className="w-3 h-3 text-emerald-600" />
              Completed
            </span>
          ) : (
            <span className="inline-flex items-center gap-1 text-zinc-600">
              <Clock className="w-3 h-3 text-zinc-400" />
              {dueLabel || 'Flexible date'}
            </span>
          )}
        </div>

        {/* Assignee display */}
        <div>
          {isOpenPool ? (
            <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-blue-700 bg-blue-50 px-2 py-0.5 rounded-full border border-blue-200">
              Open Pool
            </span>
          ) : assignee ? (
            <div
              className="flex items-center gap-1.5"
              title={`Assigned to ${assignee.name} (${assignee.role})`}
            >
              <img
                src={assignee.avatar_url}
                alt={assignee.name}
                className="w-5 h-5 rounded-full object-cover border border-zinc-200"
              />
              <span className="text-xs font-semibold text-zinc-700 truncate max-w-[100px]">
                {assignee.name.split(' ')[0]}
              </span>
            </div>
          ) : (
            <span className="text-xs text-zinc-400 italic">Unassigned</span>
          )}
        </div>
      </div>

      {/* Action Footer: 1-Tap Complete OR Claim Button */}
      <div className="pt-2">
        {isOpenPool ? (
          <button
            id={`claim-btn-${chore.id}`}
            onClick={() => onClaim(chore.id)}
            className="w-full flex items-center justify-center gap-1.5 bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white text-xs font-semibold py-2 px-3 rounded-xl transition-colors cursor-pointer shadow-xs"
          >
            <UserPlus className="w-3.5 h-3.5" />
            Claim Task
          </button>
        ) : (
          <button
            id={`complete-btn-${chore.id}`}
            onClick={handleCompleteClick}
            disabled={isCompleted || isPendingApproval}
            className={`w-full flex items-center justify-center gap-2 py-2 px-3 rounded-xl text-xs font-semibold transition-all cursor-pointer ${
              isCompleted
                ? 'bg-emerald-100/70 text-emerald-800 cursor-default border border-emerald-200'
                : isPendingApproval
                ? 'bg-amber-100/80 text-amber-900 cursor-default border border-amber-200'
                : 'bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 text-white shadow-xs'
            }`}
          >
            {isCompleted ? (
              <>
                <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                Done & Verified
              </>
            ) : isPendingApproval ? (
              <>
                <Clock className="w-4 h-4 text-amber-600 animate-spin" />
                Awaiting Verification
              </>
            ) : (
              <>
                <Check className="w-4 h-4 stroke-[2.5]" />
                {chore.requires_proof || chore.requires_approval
                  ? 'Submit for Approval'
                  : 'Mark Completed'}
              </>
            )}
          </button>
        )}
      </div>
    </div>
  );
};
