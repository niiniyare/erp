## 19. ⚡ Performance Guidelines

### 19.1 Performance Philosophy

Performance is a feature. Fast interfaces feel responsive and professional.

**Core Web Vitals:**
- **LCP** (Largest Contentful Paint): < 2.5s
- **FID** (First Input Delay): < 100ms
- **CLS** (Cumulative Layout Shift): < 0.1

### 19.2 CSS Performance

```css
/* ✅ Good: Use transform and opacity */
.element {
  transform: translateX(100px);
  opacity: 0.5;
  transition: transform 200ms, opacity 200ms;
}

/* ❌ Bad: Avoid layout-triggering properties */
.element {
  left: 100px;
  margin-left: 20px;
  width: 500px;
}

/* Use will-change sparingly */
.animating-element {
  will-change: transform;
}

.animating-element.animation-complete {
  will-change: auto;
}

/* Contain layout calculations */
.card {
  contain: layout style paint;
}
```

### 19.3 Image Optimization

```html
<!-- Responsive images -->
<img
  src="image-800w.jpg"
  srcset="
    image-400w.jpg 400w,
    image-800w.jpg 800w,
    image-1200w.jpg 1200w
  "
  sizes="(max-width: 640px) 100vw, 800px"
  alt="Description"
  loading="lazy"
  decoding="async"
/>

<!-- Modern formats with fallback -->
<picture>
  <source type="image/avif" srcset="image.avif" />
  <source type="image/webp" srcset="image.webp" />
  <img src="image.jpg" alt="Description" />
</picture>
```

### 19.4 Code Splitting

```typescript
// Lazy load components
const Dashboard = React.lazy(() => import('./Dashboard'));
const Settings = React.lazy(() => import('./Settings'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Suspense>
  );
}

// Dynamic imports
async function loadModule() {
  const module = await import('./heavy-module');
  module.init();
}
```

### 19.5 Memoization

```typescript
// Memo component
const ExpensiveComponent = React.memo(({ data }) => {
  return <div>{/* Render data */}</div>;
});

// useMemo for expensive calculations
function DataList({ items }) {
  const sortedItems = React.useMemo(() => {
    return items.sort((a, b) => a.name.localeCompare(b.name));
  }, [items]);
  
  return <>{/* Render sortedItems */}</>;
}

// useCallback for stable references
function Parent() {
  const handleClick = React.useCallback(() => {
    console.log('Clicked');
  }, []);
  
  return <Child onClick={handleClick} />;
}
```

### 19.6 Bundle Size Optimization

```javascript
// Import only what you need
// ❌ Bad
import _ from 'lodash';

// ✅ Good
import debounce from 'lodash/debounce';

// Tree-shakeable imports
// ✅ Good
import { Button } from '@/components/ui/button';

// Analyze bundle
// package.json
{
  "scripts": {
    "analyze": "ANALYZE=true next build"
  }
}
```

---

