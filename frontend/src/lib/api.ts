import axios from "axios";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
});

export interface UserPreferences {
  user_id?: string;
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
  neurotype: string;
  created_at: string;
  updated_at: string;
}

export async function getPreferences(slackUserID: string, teamID = "") {
  const { data } = await api.get<UserPreferences>(
    "/api/v1/preferences/by-slack",
    { params: { slack_user_id: slackUserID, team_id: teamID } }
  );
  return data;
}

export async function updatePreferences(
  slackUserID: string,
  teamID = "",
  prefs: Record<string, unknown>
) {
  const { data } = await api.put<UserPreferences>(
    "/api/v1/preferences/by-slack",
    prefs,
    { params: { slack_user_id: slackUserID, team_id: teamID } }
  );
  return data;
}

export async function updateUser(slackUserID: string, teamID = "", data: Record<string, unknown>) {
  const { data: user } = await api.put<User>(
    "/api/v1/user/by-slack",
    data,
    { params: { slack_user_id: slackUserID, team_id: teamID } }
  );
  return user;
}

export async function getUser(slackUserID: string, teamID = "") {
  const { data } = await api.get<User>(
    "/api/v1/user/by-slack",
    { params: { slack_user_id: slackUserID, team_id: teamID } }
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
