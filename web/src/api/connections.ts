export interface ConnectionStatus {
  success: boolean;
  connection?: {
    provider: string;
    enabled: boolean;
    connected_at?: string;
  };
  message?: string;
}

export interface ConfigureResponse {
  success: boolean;
  message?: string;
  auth_url?: string;  // ← OAuth URL comes in JSON body
  provider?: string;
}

export async function getSlackStatus(): Promise<ConnectionStatus> {
  const response = await fetch('/api/connections/slack', {
    credentials: 'include',
  });
  return response.json();
}

export async function configureSlack(enabled: boolean): Promise<ConfigureResponse> {
  const response = await fetch('/api/connections/slack', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ enabled }),
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to configure Slack');
  }
  
  // ✅ OAuth URL is in JSON body, not headers
  const data = await response.json();
  return data;  // Contains: { success, auth_url, provider }
}

export async function getGithubStatus(): Promise<ConnectionStatus> {
  const response = await fetch('/api/connections/github', {
    credentials: 'include',
  });
  return response.json();
}

export async function configureGithub(enabled: boolean): Promise<ConfigureResponse> {
  const response = await fetch('/api/connections/github', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ enabled }),
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to configure GitHub');
  }
  
  // ✅ OAuth URL is in JSON body
  const data = await response.json();
  return data;  // Contains: { success, auth_url, provider }
}