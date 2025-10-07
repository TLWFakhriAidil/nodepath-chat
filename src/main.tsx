// BASIC REACT TEST
console.log('🚨 MAIN.TSX: React main.tsx file loaded!');

import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'

console.log('🚨 MAIN.TSX: About to render React app');

try {
  createRoot(document.getElementById("root")!).render(<App />);
  console.log('🚨 MAIN.TSX: React app rendered successfully!');
} catch (error) {
  console.error('❌ MAIN.TSX: React render failed:', error);
}
