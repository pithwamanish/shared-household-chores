import React, { useState, useEffect } from 'react';
import { Sparkles, Mail, Lock, UserPlus, KeyRound, ArrowRight, CheckCircle2, Home, Users, Tablet, Inbox, ShieldCheck, Video } from 'lucide-react';
import { api } from '../services/api';
import { Household, AuthTokenResponse, HouseholdMode } from '../types';
import { DevMailboxModal } from './DevMailboxModal';

interface LoginViewProps {
  onLoginSuccess: (authData: AuthTokenResponse) => void;
  onEnterKioskMode?: () => void;
  availableHouseholds?: Household[];
}

export const LoginView: React.FC<LoginViewProps> = ({
  onLoginSuccess,
  onEnterKioskMode,
}) => {
  const [activeTab, setActiveTab] = useState<'login' | 'register' | 'magic' | 'forgot' | 'reset'>('login');

  // Password Login state
  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [isLoggingIn, setIsLoggingIn] = useState(false);
  const [loginError, setLoginError] = useState('');

  // Registration state
  const [regName, setRegName] = useState('');
  const [regEmail, setRegEmail] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [regMode, setRegMode] = useState<'create' | 'join'>('create');
  const [regHouseholdName, setRegHouseholdName] = useState('');
  const [regHouseholdMode, setRegHouseholdMode] = useState<HouseholdMode>('flatmate');
  const [regInviteCode, setRegInviteCode] = useState('');
  const [regRole, setRegRole] = useState<'admin' | 'member' | 'child'>('admin');
  const [isRegistering, setIsRegistering] = useState(false);
  const [registerError, setRegisterError] = useState('');

  // Magic link state
  const [magicEmail, setMagicEmail] = useState('');
  const [magicCode, setMagicCode] = useState('');
  const [sentCode, setSentCode] = useState('');
  const [isRequestingMagic, setIsRequestingMagic] = useState(false);
  const [isVerifyingMagic, setIsVerifyingMagic] = useState(false);
  const [magicMessage, setMagicMessage] = useState('');
  const [magicError, setMagicError] = useState('');

  // Forgot password state
  const [forgotEmail, setForgotEmail] = useState('');
  const [isRequestingForgot, setIsRequestingForgot] = useState(false);
  const [forgotMessage, setForgotMessage] = useState('');
  const [forgotError, setForgotError] = useState('');
  const [forgotToken, setForgotToken] = useState('');

  // Reset password state
  const [resetToken, setResetToken] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isResetting, setIsResetting] = useState(false);
  const [resetError, setResetError] = useState('');
  const [resetSuccess, setResetSuccess] = useState(false);

  // Dev mailbox modal
  const [isMailboxOpen, setIsMailboxOpen] = useState(false);

  // Check URL parameters for reset or magic tokens
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const resetTok = params.get('reset_token');
    if (resetTok) {
      setResetToken(resetTok);
      setActiveTab('reset');
    }
    const magicTok = params.get('magic_token');
    if (magicTok) {
      setMagicCode(magicTok);
      setActiveTab('magic');
      setIsVerifyingMagic(true);
      setMagicError('');
      api.verifyMagicLink(magicTok.trim())
        .then((res) => {
          onLoginSuccess(res);
        })
        .catch((err: any) => {
          setMagicError(err?.message || 'Invalid or expired magic link. Magic links are single-use and expire after 15 minutes.');
        })
        .finally(() => {
          setIsVerifyingMagic(false);
        });
    }
  }, [onLoginSuccess]);

  // Demo accounts for quick credentials population
  const demoAccounts = [
    {
      id: 'm-sarah',
      name: 'Sarah Chen',
      email: 'sarah@example.com',
      password: 'password123',
      role: 'Admin',
      household: 'Apartment 4B',
      avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=120&auto=format&fit=crop&q=80',
    },
    {
      id: 'm-liam',
      name: 'Liam Vance',
      email: 'liam@example.com',
      password: 'password123',
      role: 'Member',
      household: 'Apartment 4B',
      avatar: 'https://images.unsplash.com/photo-1539571696357-5a69c17a67c6?w=120&auto=format&fit=crop&q=80',
    },
    {
      id: 'm-david-fam',
      name: 'David (Dad)',
      email: 'david@miller.family',
      password: 'password123',
      role: 'Admin',
      household: 'The Miller Family',
      avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=120&auto=format&fit=crop&q=80',
    },
    {
      id: 'm-alex',
      name: 'Alex',
      email: 'alex@example.com',
      password: 'password123',
      role: 'Admin',
      household: 'Alex & Sam',
      avatar: 'https://images.unsplash.com/photo-1522075469751-3a6694fb2f61?w=120&auto=format&fit=crop&q=80',
    },
  ];

  const fillDemoCredentials = (email: string, pass: string) => {
    setLoginEmail(email);
    setLoginPassword(pass);
    setLoginError('');
  };

  // Submit Password Login
  const handlePasswordLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!loginEmail.trim() || !loginPassword.trim()) {
      setLoginError('Please enter both email and password.');
      return;
    }
    setIsLoggingIn(true);
    setLoginError('');
    try {
      const res = await api.login(loginEmail.trim(), loginPassword.trim());
      onLoginSuccess(res);
    } catch (err: any) {
      setLoginError(err?.message || 'Invalid email or password. Please try again.');
    } finally {
      setIsLoggingIn(false);
    }
  };

  // Submit Registration
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!regName.trim() || !regEmail.trim() || !regPassword.trim()) {
      setRegisterError('Please fill in all required account fields.');
      return;
    }
    if (regMode === 'create' && !regHouseholdName.trim()) {
      setRegisterError('Please enter a household name.');
      return;
    }
    if (regMode === 'join' && !regInviteCode.trim()) {
      setRegisterError('Please enter an invite code.');
      return;
    }

    setIsRegistering(true);
    setRegisterError('');
    try {
      const payload: any = {
        name: regName.trim(),
        email: regEmail.trim(),
        password: regPassword.trim(),
        role: regRole,
      };
      if (regMode === 'create') {
        payload.household_name = regHouseholdName.trim();
      } else {
        payload.invite_code = regInviteCode.trim();
      }

      const res = await api.register(payload);
      onLoginSuccess(res);
    } catch (err: any) {
      setRegisterError(err?.message || 'Registration failed. Email might already be registered.');
    } finally {
      setIsRegistering(false);
    }
  };

  // Magic Link Request
  const handleRequestMagicLink = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!magicEmail.trim()) return;
    setIsRequestingMagic(true);
    setMagicError('');
    setMagicMessage('');
    try {
      const res = await api.requestMagicLink(magicEmail.trim());
      setMagicMessage(res.message || 'Login code sent to email!');
      if (res.token) {
        setSentCode(res.token);
        setMagicCode(res.token);
      }
    } catch (err: any) {
      setMagicError(err?.message || 'Could not send login link.');
    } finally {
      setIsRequestingMagic(false);
    }
  };

  // Magic Link Verification
  const handleVerifyMagicCode = async (e: React.FormEvent) => {
    e.preventDefault();
    const token = magicCode.trim() || sentCode.trim();
    if (!token) return;
    setIsVerifyingMagic(true);
    setMagicError('');
    try {
      const res = await api.verifyMagicLink(token);
      onLoginSuccess(res);
    } catch (err: any) {
      setMagicError(err?.message || 'Invalid or expired code. Magic links are single-use and expire after 15 minutes.');
    } finally {
      setIsVerifyingMagic(false);
    }
  };

  // Forgot Password submission
  const handleForgotPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!forgotEmail.trim()) {
      setForgotError('Please enter your email address.');
      return;
    }
    setIsRequestingForgot(true);
    setForgotError('');
    setForgotMessage('');
    try {
      const res = await api.requestPasswordReset(forgotEmail.trim());
      setForgotMessage(res.message || 'Password reset link sent! Check your inbox.');
      if (res.token) {
        setResetToken(res.token);
      }
    } catch (err: any) {
      setForgotError(err?.message || 'Failed to request password reset.');
    } finally {
      setIsRequestingForgot(false);
    }
  };

  // Reset Password submission
  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!resetToken.trim()) {
      setResetError('Password reset token is required.');
      return;
    }
    if (!newPassword.trim() || newPassword.length < 6) {
      setResetError('Password must be at least 6 characters.');
      return;
    }
    if (newPassword !== confirmPassword) {
      setResetError('Passwords do not match.');
      return;
    }
    setIsResetting(true);
    setResetError('');
    try {
      const authData = await api.resetPassword(resetToken.trim(), newPassword.trim());
      setResetSuccess(true);
      setTimeout(() => {
        onLoginSuccess(authData);
      }, 800);
    } catch (err: any) {
      setResetError(err?.message || 'Invalid or expired password reset token.');
    } finally {
      setIsResetting(false);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-900 text-zinc-100 flex flex-col justify-center items-center p-4 sm:p-6 font-sans">
      <div className="w-full max-w-xl">
        {/* Top Action Bar (Dev Mailbox & Architecture Video) */}
        <div className="flex justify-end items-center gap-2 mb-3">
          <a
            href="/demo.html"
            target="_blank"
            rel="noreferrer"
            id="open-architecture-video-btn"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-sky-950/60 hover:bg-sky-900/60 border border-sky-600/40 text-xs font-semibold text-sky-300 hover:text-white transition-colors cursor-pointer shadow-sm"
          >
            <Video className="w-3.5 h-3.5 text-sky-400" />
            <span>🎬 Architecture Tour Video</span>
          </a>
          <button
            type="button"
            id="open-dev-mailbox-btn"
            onClick={() => setIsMailboxOpen(true)}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-zinc-800/80 hover:bg-zinc-700/80 border border-zinc-700 text-xs font-semibold text-zinc-300 hover:text-white transition-colors cursor-pointer shadow-sm"
          >
            <Inbox className="w-3.5 h-3.5 text-indigo-400" />
            <span>📬 Dev Mailbox</span>
          </button>
        </div>

        {/* Brand Header */}
        <div className="text-center mb-6">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-indigo-600 shadow-xl shadow-indigo-600/30 mb-3">
            <Sparkles className="w-8 h-8 text-amber-300" />
          </div>
          <h1 className="text-2xl sm:text-3xl font-black tracking-tight text-white">ChoreSync</h1>
          <p className="text-sm text-zinc-400 mt-1">
            Equitable, flexible household chore coordination & shared home harmony
          </p>
        </div>

        {/* Auth Card */}
        <div className="bg-zinc-800/90 backdrop-blur border border-zinc-700/80 rounded-3xl p-6 sm:p-8 shadow-2xl">
          {/* Top Tabs */}
          <div className="grid grid-cols-3 gap-1 bg-zinc-900/80 p-1 rounded-2xl mb-6 border border-zinc-700/50">
            <button
              type="button"
              id="tab-login"
              onClick={() => setActiveTab('login')}
              className={`flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                activeTab === 'login'
                  ? 'bg-zinc-700 text-white shadow-xs'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              <Lock className="w-3.5 h-3.5" />
              <span>Sign In</span>
            </button>
            <button
              type="button"
              id="tab-register"
              onClick={() => setActiveTab('register')}
              className={`flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                activeTab === 'register'
                  ? 'bg-zinc-700 text-white shadow-xs'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              <UserPlus className="w-3.5 h-3.5" />
              <span>Create Account</span>
            </button>
            <button
              type="button"
              id="tab-magic"
              onClick={() => setActiveTab('magic')}
              className={`flex items-center justify-center gap-1.5 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer ${
                activeTab === 'magic'
                  ? 'bg-zinc-700 text-white shadow-xs'
                  : 'text-zinc-400 hover:text-zinc-200'
              }`}
            >
              <Mail className="w-3.5 h-3.5" />
              <span>Magic Link</span>
            </button>
          </div>

          {/* TAB 1: PASSWORD LOGIN */}
          {activeTab === 'login' && (
            <div className="space-y-5">
              <form onSubmit={handlePasswordLogin} className="space-y-4">
                <div>
                  <label
                    htmlFor="login-email-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1.5"
                  >
                    Email Address
                  </label>
                  <div className="relative">
                    <Mail className="w-4 h-4 text-zinc-500 absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" />
                    <input
                      id="login-email-input"
                      type="email"
                      required
                      value={loginEmail}
                      onChange={(e) => setLoginEmail(e.target.value)}
                      placeholder="e.g. sarah@example.com"
                      className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2.5 pl-10 pr-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                </div>

                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <label
                      htmlFor="login-password-input"
                      className="block text-xs font-bold uppercase tracking-wider text-zinc-400"
                    >
                      Password
                    </label>
                    <div className="flex items-center gap-2">
                      <span className="text-[11px] text-zinc-500">Default: password123</span>
                      <span className="text-zinc-600">•</span>
                      <button
                        type="button"
                        id="forgot-password-link"
                        onClick={() => {
                          setForgotEmail(loginEmail);
                          setActiveTab('forgot');
                        }}
                        className="text-[11px] text-indigo-400 hover:text-indigo-300 underline font-medium cursor-pointer"
                      >
                        Forgot password?
                      </button>
                    </div>
                  </div>
                  <div className="relative">
                    <Lock className="w-4 h-4 text-zinc-500 absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" />
                    <input
                      id="login-password-input"
                      type="password"
                      required
                      value={loginPassword}
                      onChange={(e) => setLoginPassword(e.target.value)}
                      placeholder="••••••••"
                      className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2.5 pl-10 pr-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                </div>

                {loginError && (
                  <div
                    id="login-error-message"
                    className="p-3 rounded-xl bg-rose-950/70 border border-rose-800 text-rose-300 text-xs font-medium"
                  >
                    {loginError}
                  </div>
                )}

                <button
                  type="submit"
                  id="login-submit-btn"
                  disabled={isLoggingIn}
                  className="w-full bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-bold py-2.5 rounded-xl text-sm transition-all cursor-pointer flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50"
                >
                  {isLoggingIn ? 'Signing In...' : 'Sign In'}
                  <ArrowRight className="w-4 h-4" />
                </button>
              </form>

              {/* Quick Fill Demo Credentials Bar */}
              <div className="pt-4 border-t border-zinc-700/60">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-[11px] font-bold uppercase tracking-wider text-zinc-400">
                    Demo Credentials (Click to prefill)
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {demoAccounts.map((acc) => (
                    <button
                      key={acc.id}
                      type="button"
                      id={`quick-fill-${acc.id}`}
                      onClick={() => fillDemoCredentials(acc.email, acc.password)}
                      className="flex items-center gap-2 p-2 rounded-xl bg-zinc-900/60 hover:bg-zinc-700/60 border border-zinc-700/60 text-left transition-colors cursor-pointer group"
                    >
                      <img
                        src={acc.avatar}
                        alt={acc.name}
                        className="w-8 h-8 rounded-full object-cover shrink-0 border border-zinc-600 group-hover:border-indigo-400"
                      />
                      <div className="min-w-0">
                        <div className="text-xs font-bold text-white group-hover:text-indigo-300 truncate">
                          {acc.name}
                        </div>
                        <div className="text-[10px] text-zinc-400 truncate">
                          {acc.role} • {acc.household}
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* TAB 2: REGISTRATION / SIGN UP */}
          {activeTab === 'register' && (
            <form onSubmit={handleRegister} className="space-y-3.5">
              <div>
                <label
                  htmlFor="register-name-input"
                  className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                >
                  Full Name
                </label>
                <input
                  id="register-name-input"
                  type="text"
                  required
                  value={regName}
                  onChange={(e) => setRegName(e.target.value)}
                  placeholder="e.g. Maya Lin"
                  className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label
                    htmlFor="register-email-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                  >
                    Email Address
                  </label>
                  <input
                    id="register-email-input"
                    type="email"
                    required
                    value={regEmail}
                    onChange={(e) => setRegEmail(e.target.value)}
                    placeholder="maya@example.com"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label
                    htmlFor="register-password-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                  >
                    Password
                  </label>
                  <input
                    id="register-password-input"
                    type="password"
                    required
                    value={regPassword}
                    onChange={(e) => setRegPassword(e.target.value)}
                    placeholder="At least 6 chars"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              {/* Household Setup Selector */}
              <div className="pt-2 border-t border-zinc-700/60">
                <div className="flex items-center gap-2 mb-2.5">
                  <button
                    type="button"
                    onClick={() => {
                      setRegMode('create');
                      setRegRole('admin');
                    }}
                    className={`flex-1 py-1.5 rounded-lg text-xs font-bold border transition-colors cursor-pointer ${
                      regMode === 'create'
                        ? 'bg-indigo-600 border-indigo-500 text-white'
                        : 'bg-zinc-900 border-zinc-700 text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    Create New Household
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setRegMode('join');
                      setRegRole('member');
                    }}
                    className={`flex-1 py-1.5 rounded-lg text-xs font-bold border transition-colors cursor-pointer ${
                      regMode === 'join'
                        ? 'bg-indigo-600 border-indigo-500 text-white'
                        : 'bg-zinc-900 border-zinc-700 text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    Join with Invite Code
                  </button>
                </div>

                {regMode === 'create' ? (
                  <div className="space-y-3">
                    <div>
                      <label
                        htmlFor="register-household-name-input"
                        className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                      >
                        Household / Flat Name
                      </label>
                      <input
                        id="register-household-name-input"
                        type="text"
                        required
                        value={regHouseholdName}
                        onChange={(e) => setRegHouseholdName(e.target.value)}
                        placeholder="e.g. 742 Evergreen Terrace"
                        className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                      />
                    </div>
                  </div>
                ) : (
                  <div>
                    <label
                      htmlFor="register-invite-code-input"
                      className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                    >
                      Household Invite Code
                    </label>
                    <input
                      id="register-invite-code-input"
                      type="text"
                      required
                      value={regInviteCode}
                      onChange={(e) => setRegInviteCode(e.target.value)}
                      placeholder="e.g. APT4B-SHARE or MILLER-HOME"
                      className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500 uppercase"
                    />
                  </div>
                )}
              </div>

              {registerError && (
                <div
                  id="register-error-message"
                  className="p-3 rounded-xl bg-rose-950/70 border border-rose-800 text-rose-300 text-xs font-medium"
                >
                  {registerError}
                </div>
              )}

              <button
                type="submit"
                id="register-submit-btn"
                disabled={isRegistering}
                className="w-full bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-bold py-2.5 rounded-xl text-sm transition-all cursor-pointer flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50 mt-3"
              >
                {isRegistering ? 'Creating Account...' : 'Create Account & Sign In'}
              </button>
            </form>
          )}

          {/* TAB 3: MAGIC LINK */}
          {activeTab === 'magic' && (
            <div className="space-y-4">
              {isVerifyingMagic && (
                <div className="p-3.5 rounded-xl bg-indigo-950/70 border border-indigo-800 text-indigo-200 text-xs flex items-center gap-2.5 animate-pulse">
                  <Sparkles className="w-4 h-4 text-indigo-400 animate-spin shrink-0" />
                  <span>Verifying magic link token... Signing you into ChoreSync...</span>
                </div>
              )}

              <div className="flex items-center gap-2 p-2.5 rounded-xl bg-zinc-900/80 border border-zinc-800 text-zinc-400 text-xs">
                <Sparkles className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                <span>Magic links are single-use and expire after 15 minutes.</span>
              </div>

              <form onSubmit={handleRequestMagicLink} className="space-y-3">
                <div>
                  <label
                    htmlFor="magic-email-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1.5"
                  >
                    Your Registered Email
                  </label>
                  <div className="relative">
                    <Mail className="w-4 h-4 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                    <input
                      id="magic-email-input"
                      type="email"
                      required
                      value={magicEmail}
                      onChange={(e) => setMagicEmail(e.target.value)}
                      placeholder="sarah@example.com"
                      className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2.5 pl-9 pr-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                </div>

                <button
                  type="submit"
                  id="send-magic-link-btn"
                  disabled={isRequestingMagic}
                  className="w-full bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-bold py-2.5 rounded-xl text-sm transition-all cursor-pointer flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50"
                >
                  {isRequestingMagic ? 'Sending Code...' : 'Send Magic Login Code'}
                </button>
              </form>

              {magicError && (
                <div id="magic-error-message" className="p-3 rounded-xl bg-rose-950/70 border border-rose-800 text-rose-300 text-xs">
                  {magicError}
                </div>
              )}

              {magicMessage && (
                <div className="p-3 rounded-xl bg-emerald-950/70 border border-emerald-800 text-emerald-300 text-xs flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                  <span>{magicMessage}</span>
                </div>
              )}

              {/* Always visible verification input */}
              <div className="pt-3 border-t border-zinc-800/80 space-y-2">
                <div className="flex items-center justify-between">
                  <label
                    htmlFor="magic-link-token-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400"
                  >
                    Enter Magic Code or Token
                  </label>
                  <span className="text-[11px] text-zinc-500">From email or mailbox</span>
                </div>
                <form onSubmit={handleVerifyMagicCode} className="flex gap-2">
                  <input
                    id="magic-link-token-input"
                    type="text"
                    value={magicCode}
                    onChange={(e) => setMagicCode(e.target.value)}
                    placeholder="Paste MAGIC-token or code"
                    className="flex-1 bg-zinc-900 border border-zinc-700 rounded-xl px-3 py-2 text-xs text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                  <button
                    type="submit"
                    id="verify-magic-link-btn"
                    disabled={isVerifyingMagic || !magicCode.trim()}
                    className="bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-bold px-4 py-2 rounded-xl text-xs cursor-pointer disabled:opacity-50 transition-colors flex items-center gap-1.5 shrink-0"
                  >
                    {isVerifyingMagic ? 'Verifying...' : 'Verify & Sign In'}
                  </button>
                </form>
              </div>
            </div>
          )}

          {/* TAB 4: FORGOT PASSWORD */}
          {activeTab === 'forgot' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between pb-2 border-b border-zinc-700/60">
                <div className="flex items-center gap-2">
                  <KeyRound className="w-4 h-4 text-amber-400" />
                  <span className="text-sm font-bold text-white">Reset Your Password</span>
                </div>
                <button
                  type="button"
                  id="back-to-login-btn-from-forgot"
                  onClick={() => setActiveTab('login')}
                  className="text-xs text-indigo-400 hover:text-indigo-300 font-medium cursor-pointer"
                >
                  ← Back to Sign In
                </button>
              </div>

              <p className="text-xs text-zinc-400 leading-relaxed">
                Enter your account email address. We will dispatch a secure single-use password reset token to your inbox.
              </p>

              <form onSubmit={handleForgotPassword} className="space-y-3">
                <div>
                  <label
                    htmlFor="forgot-email-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1.5"
                  >
                    Account Email Address
                  </label>
                  <div className="relative">
                    <Mail className="w-4 h-4 text-zinc-500 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
                    <input
                      id="forgot-email-input"
                      type="email"
                      required
                      value={forgotEmail}
                      onChange={(e) => setForgotEmail(e.target.value)}
                      placeholder="e.g. sarah@example.com"
                      className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2.5 pl-9 pr-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                </div>

                {forgotError && (
                  <div
                    id="forgot-error-message"
                    className="p-3 rounded-xl bg-rose-950/70 border border-rose-800 text-rose-300 text-xs font-medium"
                  >
                    {forgotError}
                  </div>
                )}

                {forgotMessage && (
                  <div
                    id="forgot-success-message"
                    className="p-3.5 rounded-xl bg-emerald-950/70 border border-emerald-800 text-emerald-300 text-xs space-y-2.5"
                  >
                    <div className="flex items-center gap-1.5 font-bold">
                      <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                      <span>{forgotMessage}</span>
                    </div>
                    <p className="text-[11px] text-emerald-400/80">
                      Check your transactional mailbox outbox above to inspect the reset email.
                    </p>
                    <div className="pt-1 flex items-center gap-2">
                      <button
                        type="button"
                        id="proceed-to-reset-btn"
                        onClick={() => setActiveTab('reset')}
                        className="bg-emerald-600 hover:bg-emerald-500 text-white font-bold px-3 py-1.5 rounded-lg text-xs cursor-pointer inline-flex items-center gap-1.5 shadow-xs"
                      >
                        <span>Enter Reset Token</span>
                        <ArrowRight className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                )}

                <button
                  type="submit"
                  id="forgot-password-submit-btn"
                  disabled={isRequestingForgot}
                  className="w-full bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-bold py-2.5 rounded-xl text-sm transition-all cursor-pointer flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50"
                >
                  {isRequestingForgot ? 'Sending Reset Link...' : 'Send Password Reset Email'}
                  <ArrowRight className="w-4 h-4" />
                </button>
              </form>
            </div>
          )}

          {/* TAB 5: RESET PASSWORD WITH TOKEN */}
          {activeTab === 'reset' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between pb-2 border-b border-zinc-700/60">
                <div className="flex items-center gap-2">
                  <ShieldCheck className="w-4 h-4 text-emerald-400" />
                  <span className="text-sm font-bold text-white">Choose New Password</span>
                </div>
                <button
                  type="button"
                  id="back-to-login-btn-from-reset"
                  onClick={() => setActiveTab('login')}
                  className="text-xs text-indigo-400 hover:text-indigo-300 font-medium cursor-pointer"
                >
                  ← Back to Sign In
                </button>
              </div>

              <p className="text-xs text-zinc-400 leading-relaxed">
                Provide your password reset token and choose a new secure password (minimum 6 characters).
              </p>

              <form onSubmit={handleResetPassword} className="space-y-3.5">
                <div>
                  <label
                    htmlFor="reset-token-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                  >
                    Reset Token
                  </label>
                  <input
                    id="reset-token-input"
                    type="text"
                    required
                    value={resetToken}
                    onChange={(e) => setResetToken(e.target.value)}
                    placeholder="Paste reset token from email"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white font-mono placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label
                    htmlFor="reset-new-password-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                  >
                    New Password
                  </label>
                  <input
                    id="reset-new-password-input"
                    type="password"
                    required
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    placeholder="At least 6 characters"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label
                    htmlFor="reset-confirm-password-input"
                    className="block text-xs font-bold uppercase tracking-wider text-zinc-400 mb-1"
                  >
                    Confirm New Password
                  </label>
                  <input
                    id="reset-confirm-password-input"
                    type="password"
                    required
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="Re-type new password"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded-xl py-2 px-3 text-sm text-white placeholder-zinc-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>

                {resetError && (
                  <div
                    id="reset-error-message"
                    className="p-3 rounded-xl bg-rose-950/70 border border-rose-800 text-rose-300 text-xs font-medium"
                  >
                    {resetError}
                  </div>
                )}

                {resetSuccess && (
                  <div
                    id="reset-success-message"
                    className="p-3 rounded-xl bg-emerald-950/70 border border-emerald-800 text-emerald-300 text-xs font-medium flex items-center gap-2"
                  >
                    <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                    <span>Password updated successfully! Signing you in...</span>
                  </div>
                )}

                <button
                  type="submit"
                  id="reset-password-submit-btn"
                  disabled={isResetting || resetSuccess}
                  className="w-full bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-bold py-2.5 rounded-xl text-sm transition-all cursor-pointer flex items-center justify-center gap-2 shadow-md shadow-indigo-600/30 disabled:opacity-50 mt-2"
                >
                  {isResetting ? 'Updating Password...' : 'Update Password & Sign In'}
                  <ArrowRight className="w-4 h-4" />
                </button>
              </form>
            </div>
          )}

          {/* Shared Kitchen Kiosk Quick Shortcut */}
          {onEnterKioskMode && (
            <div className="mt-6 pt-5 border-t border-zinc-700/60 flex items-center justify-between">
              <div className="text-xs text-zinc-400">
                <span className="font-semibold text-zinc-300">Kitchen Fridge Tablet?</span>
                <p className="text-[11px] text-zinc-500">Shared touch display for counter or fridge</p>
              </div>
              <button
                type="button"
                id="launch-kiosk-btn"
                onClick={onEnterKioskMode}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-zinc-700 hover:bg-zinc-600 text-xs font-bold text-white transition-colors cursor-pointer"
              >
                <Tablet className="w-3.5 h-3.5 text-indigo-400" />
                <span>Launch Kiosk Display</span>
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Dev Mailbox Modal */}
      <DevMailboxModal
        isOpen={isMailboxOpen}
        onClose={() => setIsMailboxOpen(false)}
        onSelectResetToken={(token) => {
          setResetToken(token);
          setActiveTab('reset');
          setIsMailboxOpen(false);
        }}
        onSelectMagicToken={(token) => {
          setMagicCode(token);
          setActiveTab('magic');
          setIsMailboxOpen(false);
          setIsVerifyingMagic(true);
          setMagicError('');
          api.verifyMagicLink(token.trim())
            .then((res) => {
              onLoginSuccess(res);
            })
            .catch((err: any) => {
              setMagicError(err?.message || 'Invalid or expired magic link.');
            })
            .finally(() => {
              setIsVerifyingMagic(false);
            });
        }}
      />
    </div>
  );
};
