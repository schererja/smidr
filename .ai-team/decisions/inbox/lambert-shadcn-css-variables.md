### 2026-02-11: shadcn/ui requires CSS variable definitions in App.css

**By:** Lambert

**What:** When using shadcn/ui components, you must define CSS custom properties for all semantic theme tokens in App.css within a `@layer base` block, AND extend the Tailwind config to map those variables to utility class names.

**Why:** 
shadcn/ui components use semantic class names like `bg-background`, `border-input`, `text-muted-foreground` that reference CSS variables via `hsl(var(--variable))`. If these variables aren't defined, all component styling is invalid and the UI renders completely unstyled. 

The two-part setup is required:
1. Define CSS variables in App.css: `--background: 0 0% 100%;` (HSL space-separated format)
2. Map to Tailwind utilities in config: `background: "hsl(var(--background))"`

This architecture enables runtime theme switching (light/dark) by changing CSS variable values without rebuilding. It's a common mistake to copy shadcn components without completing the theme setup - the `npx shadcn@latest init` CLI handles this automatically, but manual setup must include both pieces.

**Files affected:**
- `ui/src/App.css` - CSS variable definitions
- `ui/tailwind.config.js` - Theme color mappings
