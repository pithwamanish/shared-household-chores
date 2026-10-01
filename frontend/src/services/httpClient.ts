/// <reference types="vite/client" />

/**
 * ChoreSync Centralized HTTP Client
 * Conforms strictly to contracts/openapi.yaml.
 * Supports configurable base URL, standardized JSON envelopes, and network error classification.
 */

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
}

export interface ApiErrorResponse {
  error?: string;
  message?: string;
  code?: number;
  details?: Record<string, any>;
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly details?: Record<string, any>;

  constructor(status: number, data?: any) {
    let msg = `HTTP ${status}`;
    let code = 'HTTP_ERROR';
    let details: Record<string, any> | undefined;

    if (typeof data === 'string' && data.length > 0) {
      msg = data;
    } else if (data && typeof data === 'object') {
      if (data.message) msg = String(data.message);
      if (data.error) code = String(data.error);
      if (data.details && typeof data.details === 'object') details = data.details;
    }

    super(msg);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

/**
 * Checks whether an error is caused by a network failure, server offline, connectivity timeout,
 * rate limit exhaustion (429 Too Many Requests), or proxy routing loops (508 Loop Detected).
 */
export function isNetworkError(error: unknown): boolean {
  if (!error) return false;
  if (error instanceof ApiError) {
    // 429 Too Many Requests, 502 Bad Gateway, 503 Service Unavailable, 504 Gateway Timeout, 508 Loop Detected
    return (
      error.status === 429 ||
      error.status === 502 ||
      error.status === 503 ||
      error.status === 504 ||
      error.status === 508
    );
  }
  if (error instanceof Error) {
    const msg = (error.message || '').toLowerCase();
    const name = (error.name || '').toLowerCase();
    const code = ((error as any).code || (error as any).cause?.code || '').toString().toUpperCase();
    return (
      name === 'typeerror' ||
      name === 'aborterror' ||
      msg.includes('fetch') ||
      msg.includes('failed to fetch') ||
      msg.includes('network') ||
      msg.includes('offline') ||
      msg.includes('econnrefused') ||
      code === 'ECONNREFUSED' ||
      msg.includes('connection refused') ||
      msg.includes('too many requests') ||
      msg.includes('rate limit') ||
      msg.includes('rate_limit') ||
      msg.includes('security purposes')
    );
  }
  return false;
}

export interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: any;
  params?: Record<string, string | number | boolean | undefined | null>;
  timeoutMs?: number;
}

export interface HttpClientConfig {
  baseURL?: string;
  timeoutMs?: number;
  offlineCooldownMs?: number;
}

export class HttpClient {
  private baseURL: string;
  private timeoutMs: number;
  private offlineCooldownMs: number;
  private offlineUntil = 0;
  private resolvedPrefix: string | null = null;

  constructor(config?: HttpClientConfig) {
    const envUrl =
      typeof import.meta !== 'undefined' && import.meta.env
        ? (import.meta.env.VITE_API_BASE_URL || import.meta.env.VITE_API_URL)
        : undefined;

    const isProd = typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.PROD;

    let base =
      config?.baseURL ||
      envUrl ||
      (isProd ? '/api/v1' : 'http://localhost:8000/api/v1');

    const trimmed = base.replace(/\/+$/, '');
    if (!trimmed.endsWith('/api/v1') && !trimmed.endsWith('/api')) {
      base = `${trimmed}/api/v1`;
    }

    this.baseURL = base;
    this.timeoutMs = config?.timeoutMs ?? 20000;
    this.offlineCooldownMs = config?.offlineCooldownMs ?? 2000;
  }

  getBaseURL(): string {
    return this.resolvedPrefix || this.baseURL;
  }

  setBaseURL(url: string): void {
    this.baseURL = url;
    this.resolvedPrefix = null;
    this.resetOfflineState();
  }

  resetOfflineState(): void {
    this.offlineUntil = 0;
  }

  getAuthToken(): string | null {
    try {
      return localStorage.getItem('choresync_auth_token');
    } catch {
      return null;
    }
  }

  setAuthToken(token: string | null): void {
    try {
      if (token) {
        localStorage.setItem('choresync_auth_token', token);
      } else {
        localStorage.removeItem('choresync_auth_token');
      }
    } catch {
      // ignore
    }
  }

  isOffline(): boolean {
    return Date.now() < this.offlineUntil;
  }

  private markOffline(): void {
    this.offlineUntil = Date.now() + this.offlineCooldownMs;
  }

  private resolveUrl(endpoint: string, base: string): string {
    const root = base.replace(/\/+$/, '');
    let clean = endpoint.trim();
    if (!clean.startsWith('/')) {
      clean = '/' + clean;
    }

    // Deduplicate prefixes if endpoint already contains /api or /api/v1
    if (root.endsWith('/api/v1')) {
      if (clean.startsWith('/api/v1/')) {
        clean = clean.substring('/api/v1'.length);
      } else if (clean.startsWith('/api/')) {
        clean = clean.substring('/api'.length);
      }
    } else if (root.endsWith('/api')) {
      if (clean.startsWith('/api/')) {
        clean = clean.substring('/api'.length);
      }
    }

    return `${root}${clean}`;
  }

  private appendQueryParams(
    url: string,
    params?: Record<string, string | number | boolean | undefined | null>
  ): string {
    if (!params) return url;
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== null && value !== '') {
        query.append(key, String(value));
      }
    }
    const qs = query.toString();
    if (!qs) return url;
    return url.includes('?') ? `${url}&${qs}` : `${url}?${qs}`;
  }

  private async doFetch(url: string, options: RequestOptions): Promise<Response> {
    const { params, timeoutMs, headers, body, ...customInit } = options;
    const finalUrl = this.appendQueryParams(url, params);

    const controller = new AbortController();
    const timeout = timeoutMs || this.timeoutMs;
    const timeoutId = setTimeout(() => {
      try {
        controller.abort(new Error(`Request timed out after ${timeout}ms`));
      } catch {
        controller.abort();
      }
    }, timeout);

    const reqHeaders: Record<string, string> = {
      Accept: 'application/json',
      ...((headers as Record<string, string>) || {}),
    };

    const token = this.getAuthToken();
    if (token && !reqHeaders['Authorization']) {
      reqHeaders['Authorization'] = `Bearer ${token}`;
    }

    let serializedBody: BodyInit | undefined = undefined;
    if (body !== undefined && body !== null) {
      if (typeof body === 'string' || body instanceof FormData || body instanceof URLSearchParams) {
        serializedBody = body;
      } else {
        reqHeaders['Content-Type'] = 'application/json';
        serializedBody = JSON.stringify(body);
      }
    }

    try {
      return await fetch(finalUrl, {
        ...customInit,
        headers: reqHeaders,
        body: serializedBody,
        signal: controller.signal,
      });
    } catch (err: any) {
      if (err?.name === 'AbortError' || controller.signal.aborted) {
        throw new Error('Request timed out. Please check your internet connection and try again.');
      }
      throw err;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    if (this.isOffline()) {
      const offlineErr = new Error('Backend is currently offline (cooldown active)');
      offlineErr.name = 'TypeError';
      throw offlineErr;
    }

    const currentBase = this.resolvedPrefix || this.baseURL;
    const targetUrl = this.resolveUrl(endpoint, currentBase);

    try {
      let res = await this.doFetch(targetUrl, options);

      // If we received 404 and baseURL ends with /api/v1, adaptively attempt /api fallback
      if (res.status === 404 && targetUrl.includes('/api/v1/')) {
        const altUrl = targetUrl.replace('/api/v1/', '/api/');
        try {
          const altRes = await this.doFetch(altUrl, options);
          if (altRes.ok || altRes.status !== 404) {
            this.resolvedPrefix = this.baseURL.replace(/\/api\/v1$/, '/api');
            res = altRes;
          }
        } catch {
          // Keep original response if alternate fetch fails
        }
      }

      const contentType = res.headers.get('content-type') || '';
      let data: any = null;
      if (contentType.includes('application/json')) {
        try {
          data = await res.json();
        } catch {
          data = null;
        }
      } else {
        const text = await res.text();
        data = text ? { message: text } : null;
      }

      if (!res.ok) {
        throw new ApiError(res.status, data);
      }

      // Successful request resets offline status
      this.offlineUntil = 0;
      return data as T;
    } catch (error) {
      if (isNetworkError(error)) {
        this.markOffline();
      }
      throw error;
    }
  }

  async get<T>(endpoint: string, params?: Record<string, any>, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'GET', params });
  }

  async post<T>(endpoint: string, body?: any, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'POST', body });
  }

  async patch<T>(endpoint: string, body?: any, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'PATCH', body });
  }

  async put<T>(endpoint: string, body?: any, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'PUT', body });
  }

  async delete<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    return this.request<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

export const httpClient = new HttpClient();
