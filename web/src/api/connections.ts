export interface ConnectionStatus {
  success: boolean;
  connection?: {
    provider: string;
    enabled: boolean;
    connected_at?: string;
  };
  message?: string;
}

export async function getSlackStatus(): Promise<ConnectionStatus> {
  const response = await fetch('/api/connections/slack', {
    credentials: 'include',
  });
  return response.json();
}

export async function configureSlack(enabled: boolean): Promise<{ authUrl?: string }> {
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
  
  // Get authorization URL from header if enabling
  const authUrl = response.headers.get('X-Authorization-URL');
  const data = await response.json();
  
  return { ...data, authUrl };
}

export async function getGithubStatus(): Promise<ConnectionStatus> {
  const response = await fetch('/api/connections/github', {
    credentials: 'include',
  });
  return response.json();
}

export async function configureGithub(enabled: boolean): Promise<{ authUrl?: string }> {
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
  
  const authUrl = response.headers.get('X-Authorization-URL');
  const data = await response.json();
  
  return { ...data, authUrl };
}