import { apiClient } from "@/lib/api-client";

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  preferred_locale: string;
}

export interface AuthResponse {
  user: AuthUser;
  access_token: string;
  refresh_token: string;
}

export async function registerUser(data: {
  email: string;
  password: string;
  name: string;
}) {
  const res = await apiClient<AuthResponse>("/api/v1/auth/register", {
    method: "POST",
    body: data,
  });
  localStorage.setItem("access_token", res.access_token);
  localStorage.setItem("refresh_token", res.refresh_token);
  localStorage.setItem("user_name", res.user.name);
  localStorage.setItem("user_email", res.user.email);
  localStorage.setItem("user_id", res.user.id);
  return res;
}

export async function loginUser(data: { email: string; password: string }) {
  const res = await apiClient<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: data,
  });
  localStorage.setItem("access_token", res.access_token);
  localStorage.setItem("refresh_token", res.refresh_token);
  localStorage.setItem("user_name", res.user.name);
  localStorage.setItem("user_email", res.user.email);
  localStorage.setItem("user_id", res.user.id);
  return res;
}

export async function getMe() {
  return apiClient<AuthUser>("/api/v1/auth/me");
}

export function logout() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  localStorage.removeItem("user_name");
  localStorage.removeItem("user_email");
  localStorage.removeItem("user_id");
}
