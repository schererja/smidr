---
name: "shadcn-ui-react"
description: "Integrating shadcn/ui component library with React and Tailwind CSS"
domain: "frontend"
confidence: "low"
source: "earned"
---

## Context
shadcn/ui is a collection of reusable React components built on Radix UI and Tailwind CSS. Unlike traditional component libraries, shadcn components are copied into your project (in `src/ui/` directory), giving you full control and customization. This skill documents the setup and usage patterns for integrating shadcn/ui with a React + Vite + TypeScript project.

## Setup Patterns

### Prerequisites
Install required dependencies:
```bash
npm install class-variance-authority clsx tailwind-merge lucide-react
npm install -D tailwindcss postcss @tailwindcss/postcss @types/node
```

### TypeScript Path Aliases
Configure `tsconfig.json` for `@/*` imports:
```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

Configure `vite.config.ts` to resolve aliases:
```typescript
import path from 'path'

export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

### Tailwind Configuration
For Tailwind 4.x with Vite, use `@tailwindcss/postcss` plugin:

`postcss.config.js`:
```javascript
export default {
  plugins: {
    '@tailwindcss/postcss': {},
  },
}
```

`src/App.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

`tailwind.config.js`:
```javascript
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

### Utility Function (cn)
Create `src/lib/utils.ts` for className merging:
```typescript
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

This utility properly merges Tailwind classes, ensuring correct precedence.

## Component Patterns

### Badge Component
`src/ui/badge.tsx`:
```typescript
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold",
  {
    variants: {
      variant: {
        default: "border-transparent bg-primary text-primary-foreground",
        outline: "text-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

export function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}
```

Usage with custom styling:
```typescript
<Badge 
  variant="outline" 
  className="bg-green-100 text-green-800 border-green-300"
>
  Healthy
</Badge>
```

### Card Component
`src/ui/card.tsx`:
```typescript
const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn("rounded-lg border bg-card text-card-foreground shadow-sm", className)}
      {...props}
    />
  )
)

const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("flex flex-col space-y-1.5 p-6", className)} {...props} />
  )
)
```

Usage:
```typescript
<Card>
  <CardHeader>
    <CardTitle>System Metrics</CardTitle>
  </CardHeader>
  <CardContent>
    <p>Content here</p>
  </CardContent>
</Card>
```

### Table Component (Data Grid)
`src/ui/table.tsx` provides responsive table components:
```typescript
const Table = React.forwardRef<HTMLTableElement, React.HTMLAttributes<HTMLTableElement>>(
  ({ className, ...props }, ref) => (
    <div className="relative w-full overflow-auto">
      <table ref={ref} className={cn("w-full caption-bottom text-sm", className)} {...props} />
    </div>
  )
)
```

Usage with search filtering:
```typescript
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/ui/table';
import { Input } from '@/ui/input';
import { Search } from 'lucide-react';

function DataGrid({ items }) {
  const [search, setSearch] = useState('');
  const filtered = items.filter(item => 
    item.name.toLowerCase().includes(search.toLowerCase())
  );
  
  return (
    <>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          placeholder="Search..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-10"
        />
      </div>
      
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {filtered.map((item) => (
            <TableRow 
              key={item.id}
              onClick={() => navigate(`/detail/${item.id}`)}
              className="cursor-pointer hover:bg-gray-50"
            >
              <TableCell>{item.name}</TableCell>
              <TableCell><Badge>{item.status}</Badge></TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </>
  );
}
```

### Input Component
`src/ui/input.tsx`:
```typescript
const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
```

## Common Patterns

### Composition with Custom Classes
shadcn components accept `className` prop for Tailwind customization:
```typescript
<Card className="shadow-lg hover:shadow-xl transition-shadow">
  <CardContent className="grid grid-cols-2 gap-4">
    {/* Content */}
  </CardContent>
</Card>
```

### Responsive Layouts
Use Tailwind responsive prefixes with shadcn components:
```typescript
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {items.map(item => (
    <Card key={item.id}>{/* ... */}</Card>
  ))}
</div>
```

### Icons with lucide-react
```typescript
import { Search, ChevronRight, Clock } from 'lucide-react';

<div className="flex items-center gap-2">
  <Clock className="h-4 w-4 text-gray-500" />
  <span>{timestamp}</span>
</div>
```

### TypeScript Types
shadcn components use `React.forwardRef` with proper typing:
```typescript
export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}
```

## Anti-Patterns

- **Don't use CSS Modules with shadcn** — Stick to Tailwind utility classes for consistency.
- **Don't modify base component files** — Extend with wrapper components or className prop.
- **Don't import from node_modules** — shadcn components live in your `src/ui/` directory.
- **Don't skip path aliases** — Always use `@/` imports for cleaner code organization.

## When to Apply

Use shadcn/ui when:
- Building React applications with Tailwind CSS
- You want full control over component code (no black-box dependencies)
- You need accessible components built on Radix UI primitives
- You prefer utility-first CSS over component-scoped styles
- You want to customize components without fighting framework abstractions

## Trade-offs

**Advantages:**
- Full ownership of component code (copy-paste into project)
- No version conflicts or breaking changes from npm updates
- Customizable at source level (no theme overrides or CSS specificity battles)
- Built on accessible Radix UI primitives
- Works seamlessly with Tailwind's utility classes
- Lightweight (only includes components you use)

**Disadvantages:**
- Manual updates (no `npm update` for bug fixes)
- Requires copying each component individually
- More initial setup than traditional component libraries
- Need to understand Tailwind and React patterns
- No official CLI for setup (requires manual configuration)

## References

- Components: `ui/src/ui/badge.tsx`, `ui/src/ui/card.tsx`, `ui/src/ui/table.tsx`, `ui/src/ui/input.tsx`
- Utils: `ui/src/lib/utils.ts`
- Config: `ui/tailwind.config.js`, `ui/postcss.config.js`, `ui/vite.config.ts`, `ui/tsconfig.json`
- Examples: `ui/src/pages/SystemList.tsx` (data grid with search), `ui/src/components/HealthBadge.tsx` (custom badge styling)
- Official docs: https://ui.shadcn.com/
