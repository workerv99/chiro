/** Shapes returned by the Chiro backend (pkg/model, pkg/app/subscription_handlers.go). */

export interface CloudUser {
  user_id: string;
  email: string;
  name: string;
  role?: string;
  status?: string;
}

export interface AuthResponse {
  token: string;
  user: CloudUser;
}

export interface Subscription {
  plan: string;
  status: string;
  /** Not returned by the backend today; honored when present (ISO string or epoch ms). */
  expires_at?: string | number | null;
  limits?: Record<string, number>;
  usage?: Record<string, number>;
}

export interface RegisterInput {
  email: string;
  password: string;
  name: string;
  terms_accepted: boolean;
  privacy_accepted: boolean;
}

export interface LoginInput {
  email: string;
  password: string;
}

/** Persisted session. Implemented per platform (tokenStore.native.ts / tokenStore.ts). */
export interface SessionStore {
  getToken(): Promise<string | null>;
  setToken(token: string): Promise<void>;
  getUser(): Promise<CloudUser | null>;
  setUser(user: CloudUser): Promise<void>;
  clear(): Promise<void>;
}
