import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

// Access server-side data if available
declare global {
  interface Window {
    serverData?: any;
  }
}

createRoot(document.getElementById("root")!).render(<App />);
