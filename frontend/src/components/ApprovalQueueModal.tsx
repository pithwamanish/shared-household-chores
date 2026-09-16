import React from 'react';
import { ChoreCompletion, Chore, Member } from '../types';
import { X, ClipboardCheck } from 'lucide-react';
import { ApprovalQueueView } from './ApprovalQueueView';

export interface ApprovalQueueModalProps {
  isOpen: boolean;
  onClose: () => void;
  approvals: Array<{ completion: ChoreCompletion; chore: Chore; member: Member }>;
  currentMember: Member | null;
  onApprove: (completionId: string) => Promise<void>;
  onReject: (completionId: string, reason: string) => Promise<void>;
  onSwitchToAdminPersona: () => void;
}

export const ApprovalQueueModal: React.FC<ApprovalQueueModalProps> = ({
  isOpen,
  onClose,
  approvals,
  currentMember,
  onApprove,
  onReject,
  onSwitchToAdminPersona,
}) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="bg-white w-full max-w-4xl max-h-[90vh] rounded-3xl shadow-xl border border-zinc-200 overflow-hidden flex flex-col animate-in zoom-in-95 duration-150"
        role="dialog"
        aria-modal="true"
      >
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-100 shrink-0">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
              <ClipboardCheck className="w-4 h-4" />
            </div>
            <div>
              <h2 className="text-base font-bold text-zinc-900 tracking-tight">Parent & Admin Verification</h2>
              <p className="text-xs text-zinc-500">Review proof submissions and award points.</p>
            </div>
          </div>
          <button
            onClick={onClose}
            aria-label="Close modal"
            className="w-8 h-8 rounded-full flex items-center justify-center text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 transition-colors cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="p-6 overflow-y-auto">
          <ApprovalQueueView
            approvals={approvals}
            currentMember={currentMember}
            onApprove={onApprove}
            onReject={onReject}
            onSwitchToAdminPersona={onSwitchToAdminPersona}
          />
        </div>
      </div>
    </div>
  );
};

export { ApprovalQueueView };
export default ApprovalQueueModal;
