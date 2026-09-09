import React, { useState } from 'react';
import { ChoreCompletion, Chore, Member } from '../types';
import {
  ClipboardCheck,
  CheckCircle2,
  XCircle,
  Clock,
  Coins,
  ShieldCheck,
  MessageSquare,
  Camera,
  AlertCircle,
  ExternalLink,
} from 'lucide-react';

interface ApprovalQueueViewProps {
  approvals: Array<{ completion: ChoreCompletion; chore: Chore; member: Member }>;
  currentMember: Member | null;
  onApprove: (completionId: string) => Promise<void>;
  onReject: (completionId: string, reason: string) => Promise<void>;
  onSwitchToAdminPersona: () => void;
}

export const ApprovalQueueView: React.FC<ApprovalQueueViewProps> = ({
  approvals,
  currentMember,
  onApprove,
  onReject,
  onSwitchToAdminPersona,
}) => {
  const [rejectingId, setRejectingId] = useState<string | null>(null);
  const [rejectionReason, setRejectionReason] = useState('');
  const [isProcessing, setIsProcessing] = useState<string | null>(null);

  const isAdmin = currentMember?.role === 'admin';

  const handleApprove = async (completionId: string) => {
    setIsProcessing(completionId);
    try {
      await onApprove(completionId);
    } finally {
      setIsProcessing(null);
    }
  };

  const handleConfirmReject = async (completionId: string) => {
    if (!rejectionReason.trim()) return;
    setIsProcessing(completionId);
    try {
      await onReject(completionId, rejectionReason.trim());
      setRejectingId(null);
      setRejectionReason('');
    } finally {
      setIsProcessing(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <div className="w-8 h-8 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
              <ClipboardCheck className="w-4 h-4" />
            </div>
            <h1 className="text-lg font-bold text-zinc-900 tracking-tight">Parent & Admin Verification Queue</h1>
          </div>
          <p className="text-xs text-zinc-500">
            Inspect submitted chores, review photo evidence & notes, and award gamification points.
          </p>
        </div>

        {/* Permission status */}
        {!isAdmin && (
          <div className="flex items-center gap-2 p-2.5 rounded-xl bg-amber-50 border border-amber-200 text-xs text-amber-900">
            <AlertCircle className="w-4 h-4 text-amber-600 shrink-0" />
            <span>Currently in {currentMember?.role.toUpperCase()} persona.</span>
            <button
              onClick={onSwitchToAdminPersona}
              className="font-bold underline text-amber-950 hover:text-amber-800 cursor-pointer ml-1"
            >
              Switch to Parent/Admin
            </button>
          </div>
        )}
      </div>

      {/* Queue List */}
      {approvals.length === 0 ? (
        <div className="bg-white rounded-2xl border border-zinc-200 p-12 text-center">
          <div className="w-12 h-12 rounded-full bg-emerald-50 text-emerald-600 mx-auto flex items-center justify-center mb-3">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <h3 className="font-bold text-zinc-900 text-sm">All caught up!</h3>
          <p className="text-xs text-zinc-500 max-w-sm mx-auto mt-1">
            There are no chore submissions waiting for review. When a member completes a task requiring approval, it will appear here.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {approvals.map(({ completion, chore, member }) => {
            const isItemProcessing = isProcessing === completion.id;
            const isRejectingThis = rejectingId === completion.id;

            return (
              <div
                key={completion.id}
                id={`approval-card-${completion.id}`}
                className="bg-white rounded-2xl border border-zinc-200 p-5 flex flex-col justify-between gap-4 shadow-xs"
              >
                <div>
                  {/* Submitter row */}
                  <div className="flex items-center justify-between gap-2 mb-3">
                    <div className="flex items-center gap-2.5">
                      <img
                        src={member.avatar_url}
                        alt={member.name}
                        className="w-9 h-9 rounded-full object-cover border border-zinc-200"
                      />
                      <div>
                        <div className="font-bold text-sm text-zinc-900 flex items-center gap-1.5">
                          <span>{member.name}</span>
                          <span className="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded-full bg-zinc-100 text-zinc-600 border border-zinc-200">
                            {member.role}
                          </span>
                        </div>
                        <div className="text-[11px] text-zinc-400">
                          Submitted {new Date(completion.completed_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-1 text-xs font-extrabold px-2.5 py-1 rounded-full bg-amber-50 text-amber-900 border border-amber-200">
                      <Coins className="w-3.5 h-3.5 text-amber-600" />
                      <span>+{chore.effort_points} pts</span>
                    </div>
                  </div>

                  {/* Chore details */}
                  <div className="p-3 bg-zinc-50 rounded-xl border border-zinc-200 mb-3">
                    <div className="font-bold text-sm text-zinc-900">{chore.title}</div>
                    {chore.description && (
                      <p className="text-xs text-zinc-500 mt-0.5 leading-relaxed">{chore.description}</p>
                    )}
                  </div>

                  {/* Member proof notes */}
                  {completion.proof_notes && (
                    <div className="mb-3 text-xs">
                      <span className="font-bold text-zinc-700 block mb-1">Notes from {member.name.split(' ')[0]}:</span>
                      <div className="bg-amber-50/50 border border-amber-200/60 p-2.5 rounded-xl text-zinc-800 leading-relaxed italic">
                        "{completion.proof_notes}"
                      </div>
                    </div>
                  )}

                  {/* Member photo proof */}
                  {completion.proof_photo_url && (
                    <div className="mb-3">
                      <span className="text-xs font-bold text-zinc-700 block mb-1 flex items-center gap-1">
                        <Camera className="w-3.5 h-3.5 text-zinc-500" />
                        Submitted Photo Evidence:
                      </span>
                      <div className="relative rounded-xl overflow-hidden border border-zinc-200 max-h-48">
                        <img
                          src={completion.proof_photo_url}
                          alt="Chore proof"
                          className="w-full h-44 object-cover hover:scale-105 transition-transform"
                        />
                      </div>
                    </div>
                  )}

                  {/* Rejection input if expanding */}
                  {isRejectingThis && (
                    <div className="p-3 bg-rose-50 rounded-xl border border-rose-200 text-xs space-y-2 mb-2 animate-in fade-in">
                      <label className="font-bold text-rose-950 block">Reason for Re-work / Feedback:</label>
                      <textarea
                        rows={2}
                        value={rejectionReason}
                        onChange={(e) => setRejectionReason(e.target.value)}
                        placeholder="e.g. Please dust the baseboards as well, or put the vacuum back in the closet..."
                        className="w-full p-2 bg-white rounded-lg border border-rose-300 text-xs text-zinc-900 focus:outline-none"
                      />
                      <div className="flex justify-end gap-2">
                        <button
                          type="button"
                          onClick={() => setRejectingId(null)}
                          className="px-2.5 py-1 text-zinc-600 hover:text-zinc-900 font-semibold"
                        >
                          Cancel
                        </button>
                        <button
                          type="button"
                          onClick={() => handleConfirmReject(completion.id)}
                          disabled={!rejectionReason.trim()}
                          className="px-3 py-1 bg-rose-600 hover:bg-rose-700 text-white font-bold rounded-lg disabled:opacity-50 cursor-pointer"
                        >
                          Send Feedback
                        </button>
                      </div>
                    </div>
                  )}
                </div>

                {/* Actions */}
                <div className="pt-2 border-t border-zinc-100 flex items-center gap-2">
                  <button
                    id={`approve-btn-${completion.id}`}
                    onClick={() => handleApprove(completion.id)}
                    disabled={isItemProcessing}
                    className="flex-1 flex items-center justify-center gap-1.5 bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white text-xs font-bold py-2.5 px-3 rounded-xl transition-all shadow-xs cursor-pointer disabled:opacity-50"
                  >
                    <CheckCircle2 className="w-4 h-4" />
                    Approve (+{chore.effort_points} pts)
                  </button>

                  {!isRejectingThis && (
                    <button
                      id={`reject-btn-${completion.id}`}
                      onClick={() => {
                        setRejectingId(completion.id);
                        setRejectionReason('');
                      }}
                      disabled={isItemProcessing}
                      className="flex items-center justify-center gap-1.5 bg-zinc-100 hover:bg-rose-50 hover:text-rose-700 text-zinc-700 text-xs font-semibold py-2.5 px-3 rounded-xl transition-colors cursor-pointer"
                    >
                      <XCircle className="w-4 h-4" />
                      Needs Re-work
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
