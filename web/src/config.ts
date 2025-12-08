export const config = {
  backendUrl: import.meta.env.VITE_BACKEND_URL || window.location.origin,
  scalekitEnvironmentUrl: import.meta.env.VITE_SCALEKIT_ENVIRONMENT_URL || import.meta.env.SCALEKIT_ENVIRONMENT_URL
};

// API base URL - now served from same origin
export const API_BASE_URL = window.location.origin + '/api';
