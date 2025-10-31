## 15. 🎨 Icon System

### 15.1 Icon Philosophy

Icons enhance UX by providing visual cues and saving space. A good icon system:

1. **Consistent style** — Same design language
2. **Clear meaning** — Immediately recognizable
3. **Scalable** — Works at multiple sizes
4. **Accessible** — Proper labels and fallbacks
5. **Performant** — Optimized delivery

### 15.2 Icon Implementation

```typescript
// Icon component
interface IconProps {
  name: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  className?: string;
  'aria-label'?: string;
}

function Icon({ name, size = 'md', className, 'aria-label': ariaLabel }: IconProps) {
  const sizeMap = {
    xs: 12,
    sm: 16,
    md: 20,
    lg: 24,
    xl: 32,
  };
  
  return (
    <svg
      className={`icon icon-${size} ${className}`}
      width={sizeMap[size]}
      height={sizeMap[size]}
      aria-label={ariaLabel}
      aria-hidden={!ariaLabel}
    >
      <use href={`/icons.svg#${name}`} />
    </svg>
  );
}

// Usage
<Icon name="check" size="sm" aria-label="Success" />
<Icon name="arrow-right" />
```

### 15.3 Icon Sizes

```css
:root {
  --icon-xs: 12px;
  --icon-sm: 16px;
  --icon-md: 20px;
  --icon-lg: 24px;
  --icon-xl: 32px;
  --icon-2xl: 40px;
}

.icon {
  display: inline-block;
  vertical-align: middle;
  fill: currentColor;
}

.icon-xs { width: var(--icon-xs); height: var(--icon-xs); }
.icon-sm { width: var(--icon-sm); height: var(--icon-sm); }
.icon-md { width: var(--icon-md); height: var(--icon-md); }
.icon-lg { width: var(--icon-lg); height: var(--icon-lg); }
.icon-xl { width: var(--icon-xl); height: var(--icon-xl); }
```

### 15.4 Icon Library Structure

```
icons/
├── actions/
│   ├── add.svg
│   ├── edit.svg
│   ├── delete.svg
│   └── search.svg
├── navigation/
│   ├── arrow-left.svg
│   ├── arrow-right.svg
│   ├── chevron-down.svg
│   └── menu.svg
├── status/
│   ├── check.svg
│   ├── error.svg
│   ├── warning.svg
│   └── info.svg
├── social/
│   ├── twitter.svg
│   ├── github.svg
│   └── linkedin.svg
└── misc/
    ├── user.svg
    ├── settings.svg
    └── help.svg
```

### 15.5 Icon Sprite Generation

```javascript
// Build script to generate sprite
const fs = require('fs');
const path = require('path');
const SVGO = require('svgo');

async function generateSprite() {
  const iconsDir = path.join(__dirname, 'icons');
  const outputPath = path.join(__dirname, 'public/icons.svg');
  
  const svgo = new SVGO({
    plugins: [
      { removeViewBox: false },
      { removeDimensions: true },
    ],
  });
  
  let sprite = '<svg xmlns="http://www.w3.org/2000/svg" style="display:none">';
  
  // Process all SVG files
  const files = getAllSvgFiles(iconsDir);
  
  for (const file of files) {
    const content = fs.readFileSync(file, 'utf-8');
    const optimized = await svgo.optimize(content);
    const id = path.basename(file, '.svg');
    
    // Wrap in symbol
    sprite += `<symbol id="${id}" viewBox="0 0 24 24">${optimized.data}</symbol>`;
  }
  
  sprite += '</svg>';
  
  fs.writeFileSync(outputPath, sprite);
  console.log(`Generated sprite with ${files.length} icons`);
}
```

### 15.6 Animated Icons

```css
/* Spinning icon */
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.icon-spin {
  animation: spin 1s linear infinite;
}

/* Pulsing icon */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.icon-pulse {
  animation: pulse 2s ease-in-out infinite;
}

/* Usage */
<Icon name="loader" className="icon-spin" />
<Icon name="notification" className="icon-pulse" />
```

---

