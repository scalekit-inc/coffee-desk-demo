export const config = {
  backendUrl: import.meta.env.VITE_BACKEND_URL || window.location.origin
};

// API base URL - now served from same origin
export const API_BASE_URL = window.location.origin + '/api';
