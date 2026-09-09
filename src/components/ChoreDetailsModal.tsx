import React, { useState, useEffect } from 'react';
import { Chore, Member, ChoreComment } from '../types';
import { api } from '../services/api';
import {
  X,
  MessageSquare,
  Send,
  Coins,
  Calendar,
  Repeat,
  RotateCw,
  BellRing,
  ShieldCheck,
  Camera,
  CheckCircle2,
  Clock,
  User,
} from 'lucide-react';

interface ChoreDetailsModalProps {
  isOpen: boolean;
  onClose: () => void;
  chore: Chore | null;
  assignee?: Member | null;
  currentMember: Member | null;
  onSendNudge: (chore: Chore) => void;
  onRequestSwap: (chore: Chore) => void;
}

export const ChoreDetailsModal: React.FC<ChoreDetailsModalProps> = ({
  isOpen,
  onClose,
  chore,
  assignee,
  currentMember,
  onSendNudge,
  onRequestSwap,
}) => {
  const [comments, setComments] = useState<Array<{ comment: ChoreComment; member?: Member }>>([]);
  const [newComment, setNewComment] = useState('');
  const [isLoadingComments, setIsLoadingComments] = useState(false);

  useEffect(() => {
    if (chore && isOpen) {
      loadComments();
    }
  }, [chore, isOpen]);

  const loadComments = async () => {
    if (!chore) return;
    setIsLoadingComments(true);
    try {
      const data = await api.getComments(chore.id);
      setComments(data);
    } catch (e) {
      console.error(e);
    } finally {
      setIsLoadingComments(false);
    }
  };

  const handlePostComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim() || !chore || !currentMember) return;
    try {
      await api.addComment(chore.id, currentMember.id, newComment.trim());
      setNewComment('');
      await loadComments();
    } catch (e) {
      console.error(e);
    }
  };

  if (!isOpen || !chore) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="bg-white w-full max-w-xl rounded-3xl shadow-xl border border-zinc-200 overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-150"
        role="dialog"
        aria-modal="true"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-100">
          <div>
            <span className="text-[11px] font-bold text-zinc-400 uppercase tracking-wider capitalize">
              {chore.category} Chore
            </span>
            <h2 className="text-base font-bold text-zinc-900 leading-snug">{chore.title}</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 text-zinc-400 hover:text-zinc-700 rounded-xl"
            aria-label="Close dialog"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-5">
          {/* Description */}
          {chore.description && (
            <div className="p-3.5 bg-zinc-50 rounded-2xl border border-zinc-200 text-xs text-zinc-700 leading-relaxed">
              <span className="font-bold text-zinc-900 block mb-0.5">Task Instructions:</span>
              {chore.description}
            </div>
          )}

          {/* Details Grid */}
          <div className="grid grid-cols-2 gap-3 text-xs">
            <div className="p-3 bg-zinc-50 rounded-xl border border-zinc-200">
              <span className="text-[11px] text-zinc-400 font-semibold block mb-1">Effort Reward</span>
              <div className="flex items-center gap-1 font-extrabold text-amber-900 text-sm">
                <Coins className="w-4 h-4 text-amber-600" />
                <span>+{chore.effort_points} points</span>
              </div>
            </div>

            <div className="p-3 bg-zinc-50 rounded-xl border border-zinc-200">
              <span className="text-[11px] text-zinc-400 font-semibold block mb-1">Assignee</span>
              <div className="flex items-center gap-1.5 font-bold text-zinc-900">
                {assignee ? (
                  <>
                    <img src={assignee.avatar_url} alt="" className="w-4 h-4 rounded-full object-cover" />
                    <span>{assignee.name}</span>
                  </>
                ) : (
                  <span className="text-blue-600 font-medium">Open Household Pool</span>
                )}
              </div>
            </div>

            <div className="p-3 bg-zinc-50 rounded-xl border border-zinc-200">
              <span className="text-[11px] text-zinc-400 font-semibold block mb-1">Cadence</span>
              <div className="font-bold text-zinc-900 flex items-center gap-1 capitalize">
                <RotateCw className="w-3.5 h-3.5 text-zinc-400" />
                {chore.recurrence_type}
              </div>
            </div>

            <div className="p-3 bg-zinc-50 rounded-xl border border-zinc-200">
              <span className="text-[11px] text-zinc-400 font-semibold block mb-1">Verification</span>
              <div className="font-bold text-zinc-900 flex items-center gap-1">
                {chore.requires_approval ? (
                  <span className="text-amber-700 flex items-center gap-1">
                    <ShieldCheck className="w-3.5 h-3.5" /> Requires Approval
                  </span>
                ) : (
                  <span className="text-emerald-700 flex items-center gap-1">
                    <CheckCircle2 className="w-3.5 h-3.5" /> Instant Check-off
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Quick Actions Row */}
          <div className="flex items-center gap-2 pt-2 border-t border-zinc-100">
            {chore.current_assignee_id && (
              <button
                type="button"
                onClick={() => {
                  onSendNudge(chore);
                  onClose();
                }}
                className="flex-1 flex items-center justify-center gap-1.5 py-2 px-3 bg-amber-50 hover:bg-amber-100 text-amber-900 border border-amber-200 text-xs font-bold rounded-xl transition-colors cursor-pointer"
              >
                <BellRing className="w-3.5 h-3.5 text-amber-600" />
                Send Friendly Nudge
              </button>
            )}

            {chore.current_assignee_id && (
              <button
                type="button"
                onClick={() => {
                  onRequestSwap(chore);
                  onClose();
                }}
                className="flex-1 flex items-center justify-center gap-1.5 py-2 px-3 bg-indigo-50 hover:bg-indigo-100 text-indigo-900 border border-indigo-200 text-xs font-bold rounded-xl transition-colors cursor-pointer"
              >
                <Repeat className="w-3.5 h-3.5 text-indigo-600" />
                Request Swap
              </button>
            )}
          </div>

          {/* Comments Discussion Thread */}
          <div className="pt-3 border-t border-zinc-100">
            <div className="flex items-center gap-1.5 mb-3">
              <MessageSquare className="w-4 h-4 text-zinc-400" />
              <h3 className="text-xs font-bold text-zinc-800 uppercase tracking-wider">
                Household Discussion ({comments.length})
              </h3>
            </div>

            {/* Comments List */}
            <div className="space-y-2.5 max-h-48 overflow-y-auto mb-3">
              {comments.length === 0 ? (
                <div className="text-center py-4 text-zinc-400 text-xs italic">
                  No comments yet. Leave a note or question for your housemates below!
                </div>
              ) : (
                comments.map(({ comment, member }) => (
                  <div key={comment.id} className="p-3 bg-zinc-50 rounded-xl border border-zinc-200 text-xs">
                    <div className="flex items-center justify-between mb-1">
                      <div className="flex items-center gap-1.5 font-bold text-zinc-800">
                        {member?.avatar_url && (
                          <img src={member.avatar_url} alt="" className="w-4 h-4 rounded-full object-cover" />
                        )}
                        <span>{member?.name || 'Housemate'}</span>
                      </div>
                      <span className="text-[10px] text-zinc-400">
                        {new Date(comment.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                    <p className="text-zinc-600 leading-relaxed">{comment.message}</p>
                  </div>
                ))
              )}
            </div>

            {/* Add Comment Input */}
            <form onSubmit={handlePostComment} className="flex items-center gap-2">
              <input
                type="text"
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder={`Comment as ${currentMember?.name.split(' ')[0]}...`}
                className="flex-1 text-xs px-3 py-2 rounded-xl border border-zinc-300 focus:outline-none focus:ring-2 focus:ring-zinc-900/20"
              />
              <button
                type="submit"
                disabled={!newComment.trim()}
                className="p-2 bg-zinc-900 hover:bg-zinc-800 text-white rounded-xl disabled:opacity-40 cursor-pointer"
                aria-label="Send comment"
              >
                <Send className="w-3.5 h-3.5" />
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};
