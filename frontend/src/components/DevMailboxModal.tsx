import React, { useEffect, useState } from 'react';
import { Mail, RefreshCw, Trash2, X, ExternalLink, Clock, User, CheckCircle2, KeyRound } from 'lucide-react';
import { api } from '../services/api';
import { DevEmail } from '../types';

interface DevMailboxModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSelectResetToken?: (token: string) => void;
  onSelectMagicToken?: (token: string) => void;
}

export const DevMailboxModal: React.FC<DevMailboxModalProps> = ({
  isOpen,
  onClose,
  onSelectResetToken,
  onSelectMagicToken,
}) => {
  const [emails, setEmails] = useState<DevEmail[]>([]);
  const [selectedEmail, setSelectedEmail] = useState<DevEmail | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [viewMode, setViewMode] = useState<'html' | 'text'>('html');

  const fetchEmails = async () => {
    setIsLoading(true);
    try {
      const list = await api.getDevEmails();
      setEmails(list);
      if (list.length > 0 && !selectedEmail) {
        setSelectedEmail(list[0]);
      } else if (list.length > 0 && selectedEmail) {
        // Keep selected if still exists, or default to first
        const found = list.find((e) => e.id === selectedEmail.id);
        setSelectedEmail(found || list[0]);
      } else {
        setSelectedEmail(null);
      }
    } catch (err) {
      console.error('Failed to load dev emails:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      fetchEmails();
    }
  }, [isOpen]);

  const handleClear = async () => {
    try {
      await api.clearDevEmails();
      setEmails([]);
      setSelectedEmail(null);
    } catch (err) {
      console.error('Failed to clear dev emails:', err);
    }
  };

  const extractToken = (email: DevEmail): string => {
    if (email.token) return email.token;
    if (email.type === 'magic_link') {
      const match = email.text_body.match(/magic_token=([A-Za-z0-9_-]+)/) || email.text_body.match(/MAGIC-[A-Za-z0-9_-]+/);
      return match ? (match[1] || match[0]) : '';
    }
    if (email.type === 'password_reset') {
      const match = email.text_body.match(/reset_token=([A-Za-z0-9_-]+)/) || email.text_body.match(/RESET-[A-Za-z0-9_-]+/);
      return match ? (match[1] || match[0]) : '';
    }
    return '';
  };

  if (!isOpen) return null;

  return (
    <div id="dev-mailbox-modal" className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-black/75 backdrop-blur-sm animate-fadeIn">
      <div className="bg-zinc-900 border border-zinc-750 rounded-2xl w-full max-w-4xl h-[85vh] max-h-[750px] flex flex-col shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-800 bg-zinc-900/90">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
              <Mail className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-base font-bold text-white">ChoreSync Mailbox</h2>
                <span className="px-2 py-0.5 text-[11px] font-semibold bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 rounded-full">
                  Transactional Outbox
                </span>
              </div>
              <p className="text-xs text-zinc-400">
                Captured transactional emails (Magic Links, Password Resets, Reminders)
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              type="button"
              id="refresh-mailbox-btn"
              onClick={fetchEmails}
              disabled={isLoading}
              className="p-2 text-zinc-400 hover:text-white bg-zinc-800 hover:bg-zinc-700 rounded-lg transition-colors cursor-pointer"
              title="Refresh Emails"
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
            <button
              type="button"
              id="clear-mailbox-btn"
              onClick={handleClear}
              disabled={emails.length === 0}
              className="p-2 text-zinc-400 hover:text-rose-400 bg-zinc-800 hover:bg-zinc-700 rounded-lg transition-colors cursor-pointer disabled:opacity-40"
              title="Clear Mailbox"
            >
              <Trash2 className="w-4 h-4" />
            </button>
            <button
              type="button"
              id="close-dev-mailbox-btn"
              onClick={onClose}
              className="p-2 text-zinc-400 hover:text-white bg-zinc-800 hover:bg-zinc-700 rounded-lg transition-colors cursor-pointer ml-1"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Content Body: Split view */}
        <div className="flex-1 flex overflow-hidden">
          {/* Email List Sidebar */}
          <div className="w-1/3 min-w-[260px] border-r border-zinc-800 flex flex-col bg-zinc-950/50">
            <div className="p-3 border-b border-zinc-800/80 bg-zinc-900/40 flex items-center justify-between text-xs text-zinc-400">
              <span>{emails.length} message{emails.length === 1 ? '' : 's'}</span>
              <span>Latest first</span>
            </div>

            <div className="flex-1 overflow-y-auto divide-y divide-zinc-850">
              {emails.length === 0 ? (
                <div className="p-8 text-center text-zinc-500 text-xs">
                  <Mail className="w-8 h-8 mx-auto mb-2 opacity-30" />
                  No emails sent yet.<br />Trigger a magic link, password reset, or chore nudge!
                </div>
              ) : (
                emails.map((item, idx) => {
                  const isSelected = selectedEmail?.id === item.id;
                  return (
                    <button
                      key={item.id}
                      type="button"
                      id={`dev-email-item-${idx}`}
                      onClick={() => setSelectedEmail(item)}
                      className={`w-full text-left p-3.5 transition-colors cursor-pointer block ${
                        isSelected
                          ? 'bg-indigo-600/15 border-l-4 border-indigo-500'
                          : 'hover:bg-zinc-850/50 border-l-4 border-transparent'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-1 mb-1">
                        <span className="text-xs font-semibold text-zinc-200 truncate">
                          {item.to_name ? `${item.to_name} (${item.to})` : item.to}
                        </span>
                        <span className="text-[10px] text-zinc-500 whitespace-nowrap">
                          {new Date(item.sent_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                      </div>
                      <div className="text-xs text-white font-medium truncate mb-1">
                        {item.subject}
                      </div>
                      <div className="flex items-center gap-1.5">
                        <span className={`text-[10px] px-1.5 py-0.5 rounded font-mono uppercase ${
                          item.type === 'password_reset'
                            ? 'bg-rose-500/15 text-rose-300 border border-rose-500/30'
                            : item.type === 'magic_link'
                            ? 'bg-indigo-500/15 text-indigo-300 border border-indigo-500/30'
                            : 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30'
                        }`}>
                          {item.type.replace('_', ' ')}
                        </span>
                      </div>
                    </button>
                  );
                })
              )}
            </div>
          </div>

          {/* Email Preview Area */}
          <div className="flex-1 flex flex-col bg-zinc-900 overflow-hidden">
            {selectedEmail ? (
              <div className="flex-1 flex flex-col overflow-hidden">
                {/* Email Meta Bar */}
                <div className="p-4 border-b border-zinc-800 bg-zinc-900/60 space-y-2">
                  <div className="flex items-start justify-between gap-4">
                    <h3 id="dev-email-subject" className="text-sm font-bold text-white leading-tight">
                      {selectedEmail.subject}
                    </h3>
                    <div className="flex items-center gap-2 shrink-0">
                      {selectedEmail.type === 'magic_link' && (
                        <button
                          type="button"
                          id="mailbox-signin-magic-btn"
                          onClick={() => {
                            const tok = extractToken(selectedEmail);
                            if (tok) {
                              if (onSelectMagicToken) {
                                onSelectMagicToken(tok);
                              } else {
                                window.location.href = `/?magic_token=${encodeURIComponent(tok)}`;
                              }
                            }
                          }}
                          className="flex items-center gap-1.5 px-3 py-1 bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white text-xs font-bold rounded-lg transition-colors cursor-pointer shadow"
                        >
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          <span>Sign In with Magic Link</span>
                        </button>
                      )}
                      {selectedEmail.type === 'password_reset' && onSelectResetToken && (
                        <button
                          type="button"
                          id="mailbox-use-reset-btn"
                          onClick={() => {
                            const tok = extractToken(selectedEmail);
                            if (tok) {
                              onSelectResetToken(tok);
                            }
                          }}
                          className="flex items-center gap-1.5 px-3 py-1 bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white text-xs font-bold rounded-lg transition-colors cursor-pointer shadow"
                        >
                          <KeyRound className="w-3.5 h-3.5" />
                          <span>Use Reset Token</span>
                        </button>
                      )}
                      <div className="flex items-center gap-1 bg-zinc-800 p-0.5 rounded-lg border border-zinc-700/60">
                        <button
                          type="button"
                          onClick={() => setViewMode('html')}
                          className={`px-2 py-1 text-[11px] font-semibold rounded-md transition-colors cursor-pointer ${
                            viewMode === 'html' ? 'bg-zinc-700 text-white' : 'text-zinc-400 hover:text-zinc-200'
                          }`}
                        >
                          HTML
                        </button>
                        <button
                          type="button"
                          onClick={() => setViewMode('text')}
                          className={`px-2 py-1 text-[11px] font-semibold rounded-md transition-colors cursor-pointer ${
                            viewMode === 'text' ? 'bg-zinc-700 text-white' : 'text-zinc-400 hover:text-zinc-200'
                          }`}
                        >
                          Text
                        </button>
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-2 text-xs text-zinc-400 pt-1">
                    <div>
                      <span className="text-zinc-500 font-medium">To: </span>
                      <strong className="text-zinc-300">{selectedEmail.to_name}</strong> &lt;{selectedEmail.to}&gt;
                    </div>
                    <div>
                      <span className="text-zinc-500 font-medium">From: </span>
                      <span className="text-zinc-300">{selectedEmail.from}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Clock className="w-3.5 h-3.5 text-zinc-500" />
                      <span>{new Date(selectedEmail.sent_at).toLocaleString()}</span>
                    </div>
                    <div>
                      <span className="text-zinc-500 font-medium">Provider: </span>
                      <span className="capitalize text-zinc-300">{selectedEmail.provider}</span>
                    </div>
                  </div>
                </div>

                {/* Email Body Viewer */}
                <div className="flex-1 overflow-auto p-4 bg-zinc-950">
                  {viewMode === 'html' ? (
                    <div className="bg-white rounded-xl shadow-md overflow-hidden text-zinc-900">
                      <iframe
                        title="email-preview"
                        srcDoc={selectedEmail.html_body}
                        className="w-full h-[460px] border-0"
                        sandbox="allow-same-origin allow-popups allow-scripts"
                      />
                    </div>
                  ) : (
                    <pre className="p-4 bg-zinc-900 border border-zinc-800 rounded-xl text-xs text-zinc-300 font-mono whitespace-pre-wrap leading-relaxed">
                      {selectedEmail.text_body}
                    </pre>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex-1 flex flex-col items-center justify-center text-zinc-500 text-xs p-8">
                <Mail className="w-10 h-10 mb-2 opacity-20" />
                Select an email from the left sidebar to preview contents.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
