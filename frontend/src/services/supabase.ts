import { createClient, SupabaseClient } from '@supabase/supabase-js';

const supabaseUrl = (import.meta.env.VITE_SUPABASE_URL || '').trim();
const supabaseAnonKey = (import.meta.env.VITE_SUPABASE_ANON_KEY || '').trim();

export const isSupabaseConfigured = Boolean(
  supabaseUrl &&
  supabaseAnonKey &&
  !supabaseUrl.includes('your_supabase_project_url')
);

export const supabase: SupabaseClient | null = isSupabaseConfigured
  ? createClient(supabaseUrl, supabaseAnonKey, {
      auth: {
        autoRefreshToken: true,
        persistSession: true,
        detectSessionInUrl: true,
      },
    })
  : null;

/**
 * Requests a Magic Login link via Supabase Auth.
 * Supabase sends the email automatically via its shared mailer (noreply@mail.app.supabase.io)
 * with zero custom domain or DNS requirements.
 */
export async function sendSupabaseMagicLink(email: string): Promise<{ error: Error | null }> {
  if (!supabase) {
    return { error: new Error('Supabase is not configured') };
  }

  const redirectTo = typeof window !== 'undefined' ? window.location.origin : undefined;

  const { error } = await supabase.auth.signInWithOtp({
    email,
    options: {
      emailRedirectTo: redirectTo,
    },
  });

  return { error: error ? new Error(error.message) : null };
}

/**
 * Requests a password recovery email via Supabase Auth.
 */
export async function sendSupabasePasswordReset(email: string): Promise<{ error: Error | null }> {
  if (!supabase) {
    return { error: new Error('Supabase is not configured') };
  }

  const redirectTo = typeof window !== 'undefined' ? `${window.location.origin}/?reset_password=true` : undefined;

  const { error } = await supabase.auth.resetPasswordForEmail(email, {
    redirectTo,
  });

  return { error: error ? new Error(error.message) : null };
}

/**
 * Updates user password in Supabase Auth (used when resetting password via recovery session).
 */
export async function updateSupabasePassword(password: string): Promise<{ error: Error | null }> {
  if (!supabase) {
    return { error: new Error('Supabase is not configured') };
  }

  const { error } = await supabase.auth.updateUser({ password });
  return { error: error ? new Error(error.message) : null };
}

/**
 * Retrieves the current Supabase session, if any.
 */
export async function getSupabaseSession() {
  if (!supabase) return null;
  const { data } = await supabase.auth.getSession();
  return data.session;
}

/**
 * Subscribes to Supabase Auth state changes (e.g. SIGNED_IN, PASSWORD_RECOVERY).
 */
export function onSupabaseAuthStateChange(callback: (event: string, session: any) => void) {
  if (!supabase) return () => {};
  const {
    data: { subscription },
  } = supabase.auth.onAuthStateChange((event, session) => {
    callback(event, session);
  });
  return () => subscription.unsubscribe();
}

