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
  link?: string;      // ← Magic link from Scalekit
  expiry?: string;    // ← Link expiry timestamp
  success?: boolean;   // ← For disable operations
  message?: string;   // ← For disable operations
  provider?: string;
}

export async function getSlackStatus(): Promise<ConnectionStatus> {
  const response = await fetch('/api/connections?connector=slack', {
    credentials: 'include',
  });
  return response.json();
}

export async function configureSlack(enabled: boolean): Promise<ConfigureResponse> {
  const response = await fetch('/api/connections?connector=slack', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ enabled }),
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to configure Slack');
  }
  
  // ✅ Magic link response from Scalekit: { link, expiry }
  const data = await response.json();
  return data;  // Contains: { link, expiry }
}

export async function getGithubStatus(): Promise<ConnectionStatus> {
  const response = await fetch('/api/connections?connector=github-actions', {
    credentials: 'include',
  });
  return response.json();
}

export async function configureGithub(enabled: boolean): Promise<ConfigureResponse> {
  const response = await fetch('/api/connections?connector=github-actions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ enabled }),
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to configure GitHub');
  }
  
  // ✅ Magic link response from Scalekit: { link, expiry }
  const data = await response.json();
  return data;  // Contains: { link, expiry }
}