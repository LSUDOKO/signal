import axios from "axios";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
});

export interface UserPreferences {
  user_id: string;
  focus_mode_enabled: boolean;
  focus_threshold: number;
  translator_enabled: boolean;
  digest_enabled: boolean;
  digest_hour: number;
  deep_work_auto_detect: boolean;
  quiet_hours_start: string;
  quiet_hours_end: string;
}

export interface User {
  id: string;
  slack_user_id: string;
  slack_team_id: string;
  email: string;
  display_name: string;
  neurotype: Neurotype;
  created_at: string;
  updated_at: string;
}

export type Neurotype =
  | "adhd"
  | "autism"
  | "anxiety"
  | "unspecified"
  | "ally";

export async function getPreferences(userId: string) {
  const { data } = await api.get<UserPreferences>(
    `/api/v1/users/${userId}/preferences`
  );
  return data;
}

export async function updatePreferences(
  userId: string,
  prefs: Partial<UserPreferences>
) {
  const { data } = await api.put<UserPreferences>(
    `/api/v1/users/${userId}/preferences`,
    prefs
  );
  return data;
}

export async function getHealth() {
  const { data } = await api.get<{ status: string; version: string }>(
    "/health"
  );
  return data;
}

export default api;
