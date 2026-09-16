import React, { useState } from 'react';
import { Lock, Check, AlertCircle, X } from 'lucide-react';
import { api } from '../services/api';

interface PinVerificationModalProps {
  householdId: string;
  isOpen: boolean;
  onSuccess: () => void;
  onClose: () => void;
  title?: string;
  description?: string;
}

export const PinVerificationModal: React.FC<PinVerificationModalProps> = ({
  householdId,
  isOpen,
  onSuccess,
  onClose,
  title = 'Parent & Admin PIN Required',
  description = 'Enter your 4-digit household admin PIN to approve this chore on this device.',
}) => {
  const [pin, setPin] = useState('');
  const [error, setError] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);

  if (!isOpen) return null;

  const handleDigit = (digit: string) => {
    if (pin.length < 4) {
      const nextPin = pin + digit;
      setPin(nextPin);
      setError('');
      if (nextPin.length === 4) {
        verify(nextPin);
      }
    }
  };

  const handleBackspace = () => {
    setPin((prev) => prev.slice(0, -1));
    setError('');
  };

  const verify = async (codeToVerify: string) => {
    setIsVerifying(true);
    try {
      const res = await api.verifyPIN(householdId, codeToVerify);
      if (res.valid) {
        onSuccess();
        onClose();
        setPin('');
      } else {
        setError('Incorrect PIN. Please try again.');
        setPin('');
      }
    } catch {
      setError('Incorrect PIN. (Default PIN: 1234)');
      setPin('');
    } finally {
      setIsVerifying(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-3xl max-w-sm w-full p-6 shadow-2xl border border-zinc-200 text-center animate-in fade-in zoom-in-95 duration-150">
        <div className="flex justify-end">
          <button
            onClick={onClose}
            className="text-zinc-400 hover:text-zinc-600 p-1 rounded-lg hover:bg-zinc-100 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="w-12 h-12 rounded-2xl bg-amber-50 border border-amber-100 text-amber-600 flex items-center justify-center mx-auto mb-3 shadow-xs">
          <Lock className="w-6 h-6" />
        </div>

        <h3 className="text-base font-bold text-zinc-900 tracking-tight">{title}</h3>
        <p className="text-xs text-zinc-500 mt-1 mb-4">{description}</p>

        {/* PIN Indicators */}
        <div className="flex justify-center gap-3 mb-5">
          {[0, 1, 2, 3].map((i) => (
            <div
              key={i}
              className={`w-3.5 h-3.5 rounded-full transition-all ${
                i < pin.length ? 'bg-zinc-900 scale-110' : 'bg-zinc-200'
              }`}
            />
          ))}
        </div>

        {error && (
          <div id="kiosk-pin-error" className="flex items-center justify-center gap-1.5 text-xs text-rose-600 font-medium mb-4">
            <AlertCircle className="w-3.5 h-3.5" />
            <span>{error}</span>
          </div>
        )}

        {/* Keypad */}
        <div className="grid grid-cols-3 gap-2 max-w-[240px] mx-auto mb-2">
          {['1', '2', '3', '4', '5', '6', '7', '8', '9'].map((digit) => (
            <button
              key={digit}
              type="button"
              disabled={isVerifying}
              onClick={() => handleDigit(digit)}
              className="w-16 h-12 rounded-xl bg-zinc-50 hover:bg-zinc-100 active:bg-zinc-200 border border-zinc-200 font-bold text-lg text-zinc-800 transition-colors cursor-pointer"
            >
              {digit}
            </button>
          ))}
          <button
            type="button"
            onClick={() => setPin('')}
            className="w-16 h-12 rounded-xl text-xs font-semibold text-zinc-500 hover:bg-zinc-100 transition-colors cursor-pointer"
          >
            Clear
          </button>
          <button
            type="button"
            disabled={isVerifying}
            onClick={() => handleDigit('0')}
            className="w-16 h-12 rounded-xl bg-zinc-50 hover:bg-zinc-100 active:bg-zinc-200 border border-zinc-200 font-bold text-lg text-zinc-800 transition-colors cursor-pointer"
          >
            0
          </button>
          <button
            type="button"
            onClick={handleBackspace}
            className="w-16 h-12 rounded-xl text-xs font-semibold text-zinc-500 hover:bg-zinc-100 transition-colors cursor-pointer"
          >
            Del
          </button>
        </div>

        <div className="text-[11px] text-zinc-400 mt-3">
          Tip: Demo default admin PIN is <span className="font-mono font-bold text-zinc-600">1234</span>
        </div>
      </div>
    </div>
  );
};
