---
name: "vite-css-entry-import"
description: "CSS files must be imported in Vite entry point to be processed by PostCSS/Tailwind"
domain: "frontend-build"
confidence: "high"
source: "earned"
---

## Context

Vite (and other modern bundlers) only process files that are part of the module dependency graph. CSS files—even those with critical directives like Tailwind's `@tailwind`—won't be compiled unless they're imported somewhere in the JavaScript module tree.

## Problem

Setting up Tailwind config correctly (`tailwind.config.js`, `postcss.config.js`) is not enough. If the CSS file containing `@tailwind` directives isn't imported, Vite won't run it through PostCSS, and the browser receives zero styles.

**Symptom:** UI renders structure but has no styling. Build completes successfully but generates minimal or no CSS output.

## Solution

Import the CSS file in the application entry point (`main.tsx`, `main.jsx`, or equivalent) **before** importing the root component.

## Examples

```typescript
// ❌ WRONG - CSS file exists but isn't imported
// src/main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
```

```typescript
// ✅ CORRECT - CSS imported in entry point
// src/main.tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import './App.css';  // <-- Critical: import CSS before App
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
```

```css
/* src/App.css - Tailwind directives */
@tailwind base;
@tailwind components;
@tailwind utilities;
```

## Verification

1. **Build output:** Check that CSS bundle size > 0
   ```bash
   npm run build
   # Look for: dist/assets/index-[hash].css with size > 0 KB
   ```

2. **Inspect generated CSS:** Verify Tailwind utilities are present
   ```bash
   head dist/assets/index-*.css
   # Should see: .flex, .bg-gray-50, .text-3xl, etc.
   ```

3. **Browser DevTools:** Check that `<link>` tag loads CSS with Tailwind classes

## Anti-Patterns

- Don't import CSS in child components only—Vite needs it in the entry file
- Don't rely on HTML `<link>` tags in `index.html`—Vite won't process those files
- Don't assume PostCSS/Tailwind config alone is sufficient—import is mandatory
- Don't import CSS conditionally—it must be statically imported for build-time processing

## Related Patterns

- **PostCSS Plugin Chain:** Vite → PostCSS → Tailwind (or other processors) → Browser
- **Module Graph:** Only imported files are processed by Vite's build pipeline
- **Static Imports:** Dynamic imports (`import()`) won't work for CSS processing
