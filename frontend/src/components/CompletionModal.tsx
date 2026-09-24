import React, { useState } from 'react';
import { Chore, Member } from '../types';
import {
  X,
  Camera,
  CheckCircle2,
  Clock,
  ShieldCheck,
  FileText,
  UploadCloud,
  Coins,
  Sparkles,
} from 'lucide-react';
import { api } from '../services/api';

interface CompletionModalProps {
  isOpen: boolean;
  onClose: () => void;
  chore: Chore | null;
  currentMember: Member | null;
  onSubmit: (proof: { proof_notes?: string; proof_photo_url?: string }) => Promise<void>;
}

export const CompletionModal: React.FC<CompletionModalProps> = ({
  isOpen,
  onClose,
  chore,
  currentMember,
  onSubmit,
}) => {
  const [proofNotes, setProofNotes] = useState('');
  const [photoUrl, setPhotoUrl] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  if (!isOpen || !chore) return null;

  // Sample quick photos for testing / rapid submission
  const samplePhotos = [
    {
      title: 'Tidy Room',
      url: 'https://images.unsplash.com/photo-1513694203232-719a280e022f?w=600&auto=format&fit=crop&q=80',
    },
    {
      title: 'Clean Kitchen',
      url: 'https://images.unsplash.com/photo-1556911220-e15b29be8c8f?w=600&auto=format&fit=crop&q=80',
    },
    {
      title: 'Sparkling Dishes',
      url: 'https://images.unsplash.com/photo-1584622650111-993a426fbf0a?w=600&auto=format&fit=crop&q=80',
    },
    {
      title: 'Happy Dog',
      url: 'https://images.unsplash.com/photo-1543466835-00a7907e9de1?w=600&auto=format&fit=crop&q=80',
    },
  ];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      await onSubmit({
        proof_notes: proofNotes.trim() || undefined,
        proof_photo_url: photoUrl.trim() || undefined,
      });
      onClose();
    } catch (err) {
      console.error('Failed to submit completion', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setIsUploading(true);
      try {
        const uploadRes = await api.uploadPhoto(file);
        if (uploadRes?.url) {
          setPhotoUrl(uploadRes.url);
        }
      } catch (err) {
        console.warn('Cloud upload fallback to local preview', err);
        const reader = new FileReader();
        reader.onloadend = () => {
          setPhotoUrl(reader.result as string);
        };
        reader.readAsDataURL(file);
      } finally {
        setIsUploading(false);
      }
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-zinc-950/40 backdrop-blur-xs animate-in fade-in duration-150">
      <div
        className="bg-white w-full max-w-lg rounded-3xl shadow-xl border border-zinc-200 overflow-hidden flex flex-col animate-in zoom-in-95 duration-150"
        role="dialog"
        aria-modal="true"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-100">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-base font-bold text-zinc-900 tracking-tight">Complete Chore</h2>
              <p className="text-xs text-zinc-500">Record proof and submit for verification.</p>
            </div>
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
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {/* Chore Summary Card */}
          <div className="p-3.5 bg-zinc-50 rounded-2xl border border-zinc-200/80 flex items-start justify-between gap-3">
            <div>
              <div className="font-bold text-sm text-zinc-900">{chore.title}</div>
              <div className="text-xs text-zinc-500 mt-0.5 capitalize">
                Category: {chore.category}
              </div>
            </div>
            <div className="flex items-center gap-1 text-xs font-extrabold px-2.5 py-1 rounded-full bg-amber-100 text-amber-900 shrink-0">
              <Coins className="w-3.5 h-3.5 text-amber-600" />
              <span>+{chore.effort_points} pts</span>
            </div>
          </div>

          {/* Workflow Alert */}
          {chore.requires_approval ? (
            <div className="p-3 rounded-xl bg-amber-50 border border-amber-200 text-amber-900 text-xs flex items-start gap-2.5">
              <ShieldCheck className="w-4 h-4 text-amber-600 shrink-0 mt-0.5" />
              <div>
                <span className="font-bold">Approval Gate Active:</span> This chore requires inspection from a
                Parent or Admin. Once submitted, it will be placed in the Approval Queue and points will be credited upon review.
              </div>
            </div>
          ) : (
            <div className="p-3 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-900 text-xs flex items-start gap-2.5">
              <Sparkles className="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
              <div>
                <span className="font-bold">Instant Completion:</span> Points (+{chore.effort_points} pts) will be immediately credited to your balance upon submission!
              </div>
            </div>
          )}

          {/* Completion Notes */}
          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5 flex items-center gap-1">
              <FileText className="w-3.5 h-3.5 text-zinc-400" />
              Completion Notes
            </label>
            <textarea
              rows={2}
              value={proofNotes}
              onChange={(e) => setProofNotes(e.target.value)}
              placeholder="e.g. Cleaned under the couch, emptied the trash bag, folded laundry..."
              className="w-full text-xs font-medium px-3 py-2 rounded-xl border border-zinc-300 focus:outline-none focus:ring-2 focus:ring-zinc-900/20"
            />
          </div>

          {/* Photo Proof */}
          <div>
            <label className="block text-xs font-bold text-zinc-700 uppercase tracking-wider mb-1.5 flex items-center gap-1">
              <Camera className="w-3.5 h-3.5 text-zinc-400" />
              Photo Proof {chore.requires_proof && <span className="text-rose-500">*</span>}
            </label>

            {photoUrl ? (
              <div className="relative rounded-2xl overflow-hidden border border-zinc-200 mb-2">
                <img src={photoUrl} alt="Proof preview" className="w-full h-40 object-cover" />
                <button
                  type="button"
                  onClick={() => setPhotoUrl('')}
                  className="absolute top-2 right-2 bg-black/70 hover:bg-black text-white p-1 rounded-full text-xs cursor-pointer"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ) : (
              <div className="space-y-2">
                {/* Upload or Drop */}
                <label className="flex flex-col items-center justify-center p-4 border-2 border-dashed border-zinc-300 hover:border-zinc-400 rounded-2xl cursor-pointer bg-zinc-50 hover:bg-zinc-100/70 transition-colors">
                  <UploadCloud className="w-6 h-6 text-zinc-400 mb-1" />
                  <span className="text-xs font-semibold text-zinc-700">Click to upload photo or drag & drop</span>
                  <span className="text-[10px] text-zinc-400 mt-0.5">PNG, JPG, or WEBP</span>
                  <input type="file" accept="image/*" onChange={handleFileUpload} className="hidden" />
                </label>

                {/* Quick Sample Photos */}
                <div className="text-[11px] text-zinc-500 font-medium">Or choose quick sample proof photo:</div>
                <div className="grid grid-cols-4 gap-2">
                  {samplePhotos.map((s, idx) => (
                    <button
                      key={idx}
                      type="button"
                      onClick={() => setPhotoUrl(s.url)}
                      className="group relative rounded-xl overflow-hidden border border-zinc-200 hover:border-zinc-900 transition-all cursor-pointer h-14"
                    >
                      <img src={s.url} alt={s.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform" />
                      <div className="absolute inset-0 bg-black/40 flex items-end p-1 text-[9px] text-white font-semibold">
                        {s.title}
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Buttons */}
          <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-zinc-100">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-semibold text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 rounded-xl transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || (chore.requires_proof && !photoUrl && !proofNotes)}
              className="px-5 py-2 text-xs font-bold text-white bg-zinc-900 hover:bg-zinc-800 active:bg-zinc-950 rounded-xl transition-all shadow-xs disabled:opacity-50 cursor-pointer flex items-center gap-1.5"
            >
              {isSubmitting ? (
                'Submitting...'
              ) : chore.requires_approval ? (
                <>
                  <Clock className="w-3.5 h-3.5" />
                  Submit for Parent Approval
                </>
              ) : (
                <>
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  Confirm Completion
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
