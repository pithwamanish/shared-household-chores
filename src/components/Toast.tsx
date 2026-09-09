import React, { createContext, useContext, useState, useCallback } from 'react';
import { CheckCircle2, AlertCircle, Info, X, BellRing } from 'lucide-react';

export type ToastType = 'success' | 'info' | 'warning' | 'nudge';

export interface ToastMessage {
  id: string;
  title?: string;
  message: string;
  type?: ToastType;
}

interface ToastContextType {
  showToast: (message: string, type?: ToastType, title?: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const showToast = useCallback((message: string, type: ToastType = 'info', title?: string) => {
    const id = `${Date.now()}-${Math.random().toString(36).substring(2, 6)}`;
    setToasts((prev) => [...prev, { id, message, type, title }]);

    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4500);
  }, []);

  const dismissToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      {/* Toast container */}
      <div className="fixed bottom-20 md:bottom-6 right-4 md:right-6 z-50 flex flex-col gap-2.5 max-w-sm w-full pointer-events-none">
        {toasts.map((toast) => {
          let bgClass = 'bg-white border-zinc-200 text-zinc-800 shadow-lg';
          let icon = <Info className="w-5 h-5 text-blue-500 shrink-0" />;

          if (toast.type === 'success') {
            bgClass = 'bg-emerald-50 border-emerald-200 text-emerald-950 shadow-emerald-100/50';
            icon = <CheckCircle2 className="w-5 h-5 text-emerald-600 shrink-0" />;
          } else if (toast.type === 'warning') {
            bgClass = 'bg-amber-50 border-amber-200 text-amber-950 shadow-amber-100/50';
            icon = <AlertCircle className="w-5 h-5 text-amber-600 shrink-0" />;
          } else if (toast.type === 'nudge') {
            bgClass = 'bg-indigo-50 border-indigo-200 text-indigo-950 shadow-indigo-100/50';
            icon = <BellRing className="w-5 h-5 text-indigo-600 shrink-0" />;
          }

          return (
            <div
              key={toast.id}
              className={`pointer-events-auto flex items-start gap-3 p-3.5 rounded-xl border text-sm transition-all animate-in fade-in slide-in-from-bottom-3 duration-200 ${bgClass}`}
              role="alert"
            >
              {icon}
              <div className="flex-1 min-w-0 pt-0.5">
                {toast.title && <div className="font-semibold text-xs mb-0.5">{toast.title}</div>}
                <div className="leading-snug">{toast.message}</div>
              </div>
              <button
                onClick={() => dismissToast(toast.id)}
                className="text-zinc-400 hover:text-zinc-700 p-1 rounded-md transition-colors"
                aria-label="Dismiss notification"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};
