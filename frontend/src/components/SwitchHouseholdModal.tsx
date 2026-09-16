import React from 'react';
import { Home, X, Check, Users, Sparkles } from 'lucide-react';
import { Household } from '../types';

interface SwitchHouseholdModalProps {
  isOpen: boolean;
  onClose: () => void;
  households: Household[];
  activeHouseholdId: string;
  onSelectHousehold: (householdId: string) => void;
}

export const SwitchHouseholdModal: React.FC<SwitchHouseholdModalProps> = ({
  isOpen,
  onClose,
  households,
  activeHouseholdId,
  onSelectHousehold,
}) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-zinc-200 text-left animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2.5">
            <div className="w-10 h-10 rounded-2xl bg-indigo-50 border border-indigo-100 text-indigo-600 flex items-center justify-center font-bold shadow-xs">
              <Home className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-black text-zinc-900 tracking-tight">Switch Household</h3>
              <p className="text-xs text-zinc-500">Select a household workspace</p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-zinc-400 hover:text-zinc-600 p-1.5 rounded-xl hover:bg-zinc-100 transition-colors cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-2.5 my-4">
          {households.map((h) => {
            const isSelected = h.id === activeHouseholdId;
            const modeBadgeText = {
              flatmate: 'Flatmate Mode',
              family: 'Family Mode',
              casual: 'Couple Mode',
            }[h.settings.default_mode] || h.settings.default_mode;

            const modeBadgeColor = {
              flatmate: 'bg-indigo-50 text-indigo-700 border-indigo-200',
              family: 'bg-emerald-50 text-emerald-700 border-emerald-200',
              casual: 'bg-rose-50 text-rose-700 border-rose-200',
            }[h.settings.default_mode] || 'bg-zinc-50 text-zinc-700 border-zinc-200';

            return (
              <button
                key={h.id}
                type="button"
                id={`switch-house-${h.id}`}
                onClick={() => {
                  onSelectHousehold(h.id);
                  onClose();
                }}
                className={`w-full flex items-center justify-between p-3.5 rounded-2xl border text-left transition-all cursor-pointer ${
                  isSelected
                    ? 'bg-zinc-900 text-white border-zinc-900 shadow-md scale-[1.01]'
                    : 'bg-zinc-50 hover:bg-zinc-100/80 text-zinc-900 border-zinc-200 hover:border-zinc-300'
                }`}
              >
                <div className="flex items-center gap-3">
                  <div
                    className={`w-8 h-8 rounded-xl flex items-center justify-center font-bold text-xs ${
                      isSelected ? 'bg-white/20 text-white' : 'bg-white border border-zinc-200 text-zinc-700'
                    }`}
                  >
                    🏠
                  </div>
                  <div>
                    <div className="font-bold text-sm leading-tight flex items-center gap-2">
                      <span>{h.name}</span>
                    </div>
                    <div
                      className={`text-xs mt-0.5 inline-flex items-center px-1.5 py-0.5 rounded-md font-semibold ${
                        isSelected ? 'bg-white/20 text-zinc-100' : modeBadgeColor
                      }`}
                    >
                      {modeBadgeText}
                    </div>
                  </div>
                </div>

                {isSelected && (
                  <div className="w-6 h-6 rounded-full bg-emerald-500 text-white flex items-center justify-center text-xs font-bold shadow-xs">
                    <Check className="w-3.5 h-3.5" />
                  </div>
                )}
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};
