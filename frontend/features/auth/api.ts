import { apiClient, setTokens, clearTokens } from "@/lib/api-client";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  preferred_locale: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

// ---------------------------------------------------------------------------
// API functions
// ---------------------------------------------------------------------------

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await apiClient<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: data,
  });
  setTokens(res.access_token, res.refresh_token);
  return res;
}

export async function register(data: RegisterRequest): Promise<AuthResponse> {
  const res = await apiClient<AuthResponse>("/api/v1/auth/register", {
    method: "POST",
    body: data,
  });
  setTokens(res.access_token, res.refresh_token);
  return res;
}

export async function refreshTokens(
  refreshToken: string
): Promise<TokenResponse> {
  const res = await apiClient<TokenResponse>("/api/v1/auth/refresh", {
    method: "POST",
    body: { refresh_token: refreshToken },
  });
  setTokens(res.access_token, res.refresh_token);
  return res;
}

export async function getCurrentUser(): Promise<User> {
  return apiClient<User>("/api/v1/auth/me");
}

export function logout(): void {
  clearTokens();
}
