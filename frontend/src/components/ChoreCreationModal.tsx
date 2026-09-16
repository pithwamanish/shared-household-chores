import React, { useState, useEffect } from 'react';
import {
  Chore,
  Member,
  ChoreCategory,
  ChoreAssignmentType,
  ChoreRecurrenceType,
} from '../types';
import {
  X,
  Sparkles,
  Utensils,
  Trees,
  Dog,
  Wrench,
  Users,
  Repeat,
  Layers,
  Calendar,
  ShieldCheck,
  Camera,
  Coins,
  ArrowUpDown,
} from 'lucide-react';

interface ChoreCreationModalProps {
  isOpen: boolean;
  onClose: () => void;
  householdId: string;
  members: Member[];
  currentMember: Member | null;
  editingChore?: Chore | null;
  onSave: (payload: any) => Promise<void>;
}

export const ChoreCreationModal: React.FC<ChoreCreationModalProps> = ({
  isOpen,
  onClose,
  householdId,
  members,
  currentMember,
  editingChore,
  onSave,
}) => {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState<ChoreCategory>('cleaning');
  const [effortPoints, setEffortPoints] = useState(15);
  const [assignmentType, setAssignmentType] = useState<ChoreAssignmentType>('direct');
  const [directAssigneeId, setDirectAssigneeId] = useState<string>('');
  const [rotationMemberIds, setRotationMemberIds] = useState<string[]>([]);
  const [recurrenceType, setRecurrenceType] = useState<ChoreRecurrenceType>('weekly');
  const [recurrenceDays, setRecurrenceDays] = useState<string[]>(['sat']);
  const [dueDate, setDueDate] = useState('');
  const [requiresApproval, setRequiresApproval] = useState(false);
  const [requiresProof, setRequiresProof] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (editingChore) {
      setTitle(editingChore.title);
      setDescription(editingChore.description || '');
      setCategory(editingChore.category);
      setEffortPoints(editingChore.effort_points);
      setAssignmentType(editingChore.assignment_type);
      setDirectAssigneeId(editingChore.current_assignee_id || members[0]?.id || '');
      setRotationMemberIds(editingChore.rotation_member_ids || members.map((m) => m.id));
      setRecurrenceType(editingChore.recurrence_type);
      setRecurrenceDays(editingChore.recurrence_rule?.days || ['sat']);
      setDueDate(editingChore.due_date ? editingChore.due_date.substring(0, 16) : '');
      setRequiresApproval(editingChore.requires_approval);
      setRequiresProof(editingChore.requires_proof);
    } else {
      setTitle('');
      setDescription('');
      setCategory('cleaning');
      setEffortPoints(15);
      setAssignmentType('direct');
      setDirectAssigneeId(currentMember?.id || members[0]?.id || '');
      setRotationMemberIds(members.map((m) => m.id));
      setRecurrenceType('weekly');
      setRecurrenceDays(['sat']);

      // Default due date: tomorrow at 6 PM
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      tomorrow.setHours(18, 0, 0, 0);
      setDueDate(tomorrow.toISOString().substring(0, 16));

      // Default approval toggle: true if family mode and member is child, false otherwise
      setRequiresApproval(false);
      setRequiresProof(false);
    }
  }, [editingChore, members, currentMember, isOpen]);

  if (!isOpen) return null;

  const categories: Array<{ id: ChoreCategory; label: string; icon: any }> = [
    { id: 'cleaning', label: 'Cleaning', icon: Sparkles },
    { id: 'kitchen', label: 'Kitchen', icon: Utensils },
    { id: 'yard', label: 'Yard / Plants', icon: Trees },
    { id: 'pets', label: 'Pet Care', icon: Dog },
    { id: 'maintenance', label: 'Maintenance', icon: Wrench },
    { id: 'other', label: 'General', icon: Layers },
  ];

  const weekDays = [
    { id: 'mon', label: 'M' },
    { id: 'tue', label: 'T' },
    { id: 'wed', label: 'W' },
    { id: 'thu', label: 'Th' },
    { id: 'fri', label: 'F' },
    { id: 'sat', label: 'Sa' },
    { id: 'sun', label: 'Su' },
  ];

  const toggleDay = (day: string) => {
    if (recurrenceDays.includes(day)) {
      if (recurrenceDays.length > 1) {
        setRecurrenceDays(recurrenceDays.filter((d) => d !== day));
      }
    } else {
      setRecurrenceDays([...recurrenceDays, day]);
    }
  };

  const toggleRotationMember = (memberId: string) => {
    if (rotationMemberIds.includes(memberId)) {
      if (rotationMemberIds.length > 1) {
        setRotationMemberIds(rotationMemberIds.filter((id) => id !== memberId));
      }
    } else {
      setRotationMemberIds([...rotationMemberIds, memberId]);
    }
  };

  const moveMemberInRotation = (index: number, direction: 'up' | 'down') => {
    const newIdx = direction === 'up' ? index - 1 : index + 1;
    if (newIdx < 0 || newIdx >= rotationMemberIds.length) return;
    const copy = [...rotationMemberIds];
    const temp = copy[index];
    copy[index] = copy[newIdx];
    copy[newIdx] = temp;
    setRotationMemberIds(copy);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    setIsSubmitting(true);
    try {
      const payload = {
        household_id: householdId,
        title: title.trim(),
        description: description.trim() || undefined,
        category,
        effort_points: Number(effortPoints),
        assignment_type: assignmentType,
        current_assignee_id: assignmentType === 'direct' ? directAssigneeId : null,
        rotation_member_ids: assignmentType === 'round_robin' ? rotationMemberIds : undefined,
        current_rotation_index: 0,
        recurrence_type: recurrenceType,
        recurrence_rule: recurrenceType === 'custom_days' ? { days: recurrenceDays } : undefined,
        due_date: dueDate ? new Date(dueDate).toISOString() : null,
        requires_approval: requiresApproval,
        requires_proof: requiresProof,
        created_by: currentMember?.id || members[0]?.id || 'unknown',
      };

      await onSave(payload);
      onClose();
    } catch (err) {
      console.error('Failed to save chore', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="bg-white w-full max-w-xl rounded-3xl shadow-xl border border-zinc-200 overflow-hidden max-h-[92vh] flex flex-col animate-in zoom-in-95 duration-150"
        role="dialog"
        aria-modal="true"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-100">
          <div>
            <h2 className="text-lg font-bold text-zinc-900 tracking-tight">
              {editingChore ? 'Edit Chore Definition' : 'Create New Household Chore'}
            </h2>
            <p className="text-xs text-zinc-500">
              Configure assignment rules, recurrence schedule, and verification settings.
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 rounded-xl transition-colors cursor-pointer"
            aria-label="Close dialog"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-5">
          {/* Title */}
          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
              Chore Title <span className="text-rose-500">*</span>
            </label>
            <input
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Mop kitchen floor, Take out recycling..."
              className="w-full text-sm font-medium px-3.5 py-2.5 rounded-xl border border-zinc-300 focus:outline-none focus:ring-2 focus:ring-zinc-900/20 focus:border-zinc-900"
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
              Instructions & Notes (Optional)
            </label>
            <textarea
              rows={2}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Specific checklist notes, locations of cleaning supplies, or reminders..."
              className="w-full text-sm px-3.5 py-2 rounded-xl border border-zinc-300 focus:outline-none focus:ring-2 focus:ring-zinc-900/20 focus:border-zinc-900"
            />
          </div>

          {/* Category Chips */}
          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
              Category
            </label>
            <div className="grid grid-cols-3 sm:grid-cols-6 gap-2">
              {categories.map((cat) => {
                const Icon = cat.icon;
                const isSelected = category === cat.id;
                return (
                  <button
                    key={cat.id}
                    type="button"
                    onClick={() => setCategory(cat.id)}
                    className={`flex flex-col items-center justify-center p-2.5 rounded-xl border text-xs font-semibold transition-all cursor-pointer ${
                      isSelected
                        ? 'bg-zinc-900 text-white border-zinc-900 shadow-xs'
                        : 'bg-zinc-50 hover:bg-zinc-100 text-zinc-700 border-zinc-200'
                    }`}
                  >
                    <Icon className={`w-4 h-4 mb-1 ${isSelected ? 'text-amber-400' : 'text-zinc-500'}`} />
                    <span className="truncate text-[11px]">{cat.label}</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Effort Points Stepper */}
          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-xs font-bold text-zinc-700 uppercase tracking-wider flex items-center gap-1.5">
                <Coins className="w-3.5 h-3.5 text-amber-500" />
                Effort Points Reward: <span className="text-zinc-900 font-extrabold">{effortPoints} pts</span>
              </label>
            </div>
            <input
              type="range"
              min="5"
              max="50"
              step="5"
              value={effortPoints}
              onChange={(e) => setEffortPoints(Number(e.target.value))}
              className="w-full accent-zinc-900 cursor-pointer"
            />
            <div className="flex justify-between text-[10px] text-zinc-400 font-medium px-1 mt-1">
              <span>Quick 5m task (5 pts)</span>
              <span>Standard (15-20 pts)</span>
              <span>Major Deep Clean (50 pts)</span>
            </div>
          </div>

          {/* Assignment Strategy */}
          <div className="pt-2 border-t border-zinc-100">
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-2">
              Assignment Strategy
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2.5 mb-3">
              {/* Direct */}
              <button
                type="button"
                onClick={() => setAssignmentType('direct')}
                className={`p-3 rounded-2xl border text-left cursor-pointer transition-all ${
                  assignmentType === 'direct'
                    ? 'border-zinc-900 bg-zinc-900 text-white shadow-xs'
                    : 'border-zinc-200 bg-zinc-50 hover:bg-zinc-100/70 text-zinc-800'
                }`}
              >
                <div className="flex items-center gap-2 mb-1">
                  <Users className="w-4 h-4 text-indigo-400" />
                  <span className="font-bold text-xs">Direct Assign</span>
                </div>
                <p className={`text-[11px] leading-tight ${assignmentType === 'direct' ? 'text-zinc-300' : 'text-zinc-500'}`}>
                  Designate a specific household member.
                </p>
              </button>

              {/* Round Robin */}
              <button
                type="button"
                onClick={() => setAssignmentType('round_robin')}
                className={`p-3 rounded-2xl border text-left cursor-pointer transition-all ${
                  assignmentType === 'round_robin'
                    ? 'border-zinc-900 bg-zinc-900 text-white shadow-xs'
                    : 'border-zinc-200 bg-zinc-50 hover:bg-zinc-100/70 text-zinc-800'
                }`}
              >
                <div className="flex items-center gap-2 mb-1">
                  <Repeat className="w-4 h-4 text-emerald-400" />
                  <span className="font-bold text-xs">Round-Robin</span>
                </div>
                <p className={`text-[11px] leading-tight ${assignmentType === 'round_robin' ? 'text-zinc-300' : 'text-zinc-500'}`}>
                  Auto-rotates through sequence upon completion.
                </p>
              </button>

              {/* Open Pool */}
              <button
                type="button"
                onClick={() => setAssignmentType('open_pool')}
                className={`p-3 rounded-2xl border text-left cursor-pointer transition-all ${
                  assignmentType === 'open_pool'
                    ? 'border-zinc-900 bg-zinc-900 text-white shadow-xs'
                    : 'border-zinc-200 bg-zinc-50 hover:bg-zinc-100/70 text-zinc-800'
                }`}
              >
                <div className="flex items-center gap-2 mb-1">
                  <Layers className="w-4 h-4 text-sky-400" />
                  <span className="font-bold text-xs">Open Pool</span>
                </div>
                <p className={`text-[11px] leading-tight ${assignmentType === 'open_pool' ? 'text-zinc-300' : 'text-zinc-500'}`}>
                  Unassigned task claimable by anyone.
                </p>
              </button>
            </div>

            {/* Sub-inputs depending on assignment strategy */}
            {assignmentType === 'direct' && (
              <div className="bg-zinc-50 p-3 rounded-xl border border-zinc-200">
                <label className="block text-xs font-semibold text-zinc-700 mb-1">Assign To Member:</label>
                <select
                  value={directAssigneeId}
                  onChange={(e) => setDirectAssigneeId(e.target.value)}
                  className="w-full text-xs font-semibold bg-white px-3 py-2 rounded-lg border border-zinc-300 focus:outline-none"
                >
                  {members.map((m) => (
                    <option key={m.id} value={m.id}>
                      {m.name} ({m.role})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {assignmentType === 'round_robin' && (
              <div className="bg-zinc-50 p-3.5 rounded-xl border border-zinc-200 space-y-2">
                <div className="flex items-center justify-between text-xs font-semibold text-zinc-700">
                  <span>Rotation Sequence (In Order):</span>
                  <span className="text-[11px] text-zinc-500 font-normal">Check members to include</span>
                </div>
                <div className="space-y-1.5">
                  {members.map((m, idx) => {
                    const isChecked = rotationMemberIds.includes(m.id);
                    const positionInRotation = rotationMemberIds.indexOf(m.id);
                    return (
                      <div
                        key={m.id}
                        className="flex items-center justify-between bg-white px-3 py-1.5 rounded-lg border border-zinc-200 text-xs"
                      >
                        <label className="flex items-center gap-2 cursor-pointer select-none">
                          <input
                            type="checkbox"
                            checked={isChecked}
                            onChange={() => toggleRotationMember(m.id)}
                            className="rounded accent-zinc-900"
                          />
                          <span className="font-medium text-zinc-800">{m.name}</span>
                        </label>
                        {isChecked && (
                          <div className="flex items-center gap-1">
                            <span className="text-[10px] font-bold bg-zinc-100 text-zinc-600 px-1.5 py-0.5 rounded">
                              Turn #{positionInRotation + 1}
                            </span>
                            <button
                              type="button"
                              onClick={() => moveMemberInRotation(positionInRotation, 'up')}
                              disabled={positionInRotation === 0}
                              className="p-1 hover:bg-zinc-100 rounded disabled:opacity-30 cursor-pointer"
                              title="Move earlier in rotation"
                            >
                              ▲
                            </button>
                            <button
                              type="button"
                              onClick={() => moveMemberInRotation(positionInRotation, 'down')}
                              disabled={positionInRotation === rotationMemberIds.length - 1}
                              className="p-1 hover:bg-zinc-100 rounded disabled:opacity-30 cursor-pointer"
                              title="Move later in rotation"
                            >
                              ▼
                            </button>
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Recurrence & Due Date */}
          <div className="pt-2 border-t border-zinc-100 grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5">
                Recurrence Cadence
              </label>
              <select
                value={recurrenceType}
                onChange={(e) => setRecurrenceType(e.target.value as ChoreRecurrenceType)}
                className="w-full text-xs font-semibold bg-white px-3 py-2 rounded-xl border border-zinc-300 focus:outline-none"
              >
                <option value="none">One-Off (No Recurrence)</option>
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
                <option value="custom_days">Weekly (Selected Days)</option>
                <option value="monthly">Monthly</option>
              </select>

              {recurrenceType === 'custom_days' && (
                <div className="flex gap-1 mt-2">
                  {weekDays.map((d) => {
                    const isSelected = recurrenceDays.includes(d.id);
                    return (
                      <button
                        key={d.id}
                        type="button"
                        onClick={() => toggleDay(d.id)}
                        className={`w-7 h-7 rounded-lg text-xs font-bold transition-colors cursor-pointer ${
                          isSelected
                            ? 'bg-zinc-900 text-white'
                            : 'bg-zinc-100 text-zinc-600 hover:bg-zinc-200'
                        }`}
                      >
                        {d.label}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>

            <div>
              <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5 flex items-center gap-1">
                <Calendar className="w-3.5 h-3.5 text-zinc-400" />
                Target Due Date & Time
              </label>
              <input
                type="datetime-local"
                value={dueDate}
                onChange={(e) => setDueDate(e.target.value)}
                className="w-full text-xs font-semibold bg-white px-3 py-2 rounded-xl border border-zinc-300 focus:outline-none"
              />
            </div>
          </div>

          {/* Verification & Proof Toggles */}
          <div className="pt-2 border-t border-zinc-100 space-y-3">
            <div className="flex items-center justify-between p-3 rounded-xl bg-zinc-50 border border-zinc-200">
              <div className="flex items-center gap-2.5">
                <ShieldCheck className="w-5 h-5 text-amber-600" />
                <div>
                  <div className="text-xs font-bold text-zinc-900">Requires Parent / Admin Approval</div>
                  <div className="text-[11px] text-zinc-500">
                    Chore enters verification queue; points are credited only after approval.
                  </div>
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={requiresApproval}
                  onChange={(e) => setRequiresApproval(e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-9 h-5 bg-zinc-300 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-zinc-900"></div>
              </label>
            </div>

            <div className="flex items-center justify-between p-3 rounded-xl bg-zinc-50 border border-zinc-200">
              <div className="flex items-center gap-2.5">
                <Camera className="w-5 h-5 text-blue-600" />
                <div>
                  <div className="text-xs font-bold text-zinc-900">Requires Completion Proof</div>
                  <div className="text-[11px] text-zinc-500">
                    Prompts member to submit a completion note or photo before completing.
                  </div>
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={requiresProof}
                  onChange={(e) => setRequiresProof(e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-9 h-5 bg-zinc-300 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-zinc-900"></div>
              </label>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-2.5 pt-4 border-t border-zinc-100">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 rounded-xl transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || !title.trim()}
              className="px-5 py-2 text-xs font-bold text-white bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 rounded-xl transition-all shadow-xs disabled:opacity-50 cursor-pointer"
            >
              {isSubmitting ? 'Saving...' : editingChore ? 'Update Chore' : 'Create Chore'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
