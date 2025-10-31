## 11. 🎨 Theming Strategy

### 11.1 Theming Philosophy

Theming enables customization while maintaining consistency. A good theming system:

1. **Supports brand identity** — Easy to apply brand colors
2. **Maintains accessibility** — Enforces contrast requirements
3. **Scales effortlessly** — Works across all components
4. **Enables experimentation** — Quick theme switching
5. **Preserves semantics** — Color meanings remain clear

### 11.2 Theme Architecture

```typescript
interface Theme {
  id: string;
  name: string;
  colors: ColorPalette;
  typography: Typography;
  spacing: SpacingScale;
  radius: RadiusScale;
  shadows: ShadowScale;
}

interface ColorPalette {
  // Base colors
  background: string;
  foreground: string;
  
  // Component colors
  card: string;
  cardForeground: string;
  popover: string;
  popoverForeground: string;
  
  // Brand colors
  primary: string;
  primaryForeground: string;
  secondary: string;
  secondaryForeground: string;
  
  // Semantic colors
  muted: string;
  mutedForeground: string;
  accent: string;
  accentForeground: string;
  destructive: string;
  destructiveForeground: string;
  
  // UI elements
  border: string;
  input: string;
  ring: string;
}
```

### 11.3 Pre-built Themes

```typescript
// Default theme
const defaultTheme: Theme = {
  id: 'default',
  name: 'Default',
  colors: {
    background: '0 0% 100%',
    foreground: '222.2 84% 4.9%',
    primary: '222.2 47.4% 11.2%',
    primaryForeground: '210 40% 98%',
    // ... rest of colors
  }
};

// Slate theme
const slateTheme: Theme = {
  id: 'slate',
  name: 'Slate',
  colors: {
    background: '0 0% 100%',
    foreground: '222.2 84% 4.9%',
    primary: '215.4 16.3% 46.9%',
    primaryForeground: '210 40% 98%',
    // ... rest of colors
  }
};

// Blue theme
const blueTheme: Theme = {
  id: 'blue',
  name: 'Blue',
  colors: {
    background: '0 0% 100%',
    foreground: '222.2 84% 4.9%',
    primary: '221.2 83.2% 53.3%',
    primaryForeground: '210 40% 98%',
    // ... rest of colors
  }
};
```

### 11.4 Theme Application

```typescript
class ThemeSystem {
  private currentTheme: Theme;
  
  applyTheme(theme: Theme) {
    this.currentTheme = theme;
    const root = document.documentElement;
    
    // Apply colors
    Object.entries(theme.colors).forEach(([key, value]) => {
      const cssVar = `--${this.camelToKebab(key)}`;
      root.style.setProperty(cssVar, value);
    });
    
    // Apply typography
    Object.entries(theme.typography).forEach(([key, value]) => {
      const cssVar = `--font-${this.camelToKebab(key)}`;
      root.style.setProperty(cssVar, value);
    });
    
    // Save preference
    localStorage.setItem('theme-id', theme.id);
  }
  
  private camelToKebab(str: string): string {
    return str.replace(/[A-Z]/g, letter => `-${letter.toLowerCase()}`);
  }
}
```

### 11.5 Theme Generator

```typescript
// Generate theme from primary color
function generateTheme(primaryColor: string): Theme {
  // Parse HSL values
  const [h, s, l] = parseHSL(primaryColor);
  
  return {
    id: `custom-${Date.now()}`,
    name: 'Custom Theme',
    colors: {
      background: '0 0% 100%',
      foreground: '222.2 84% 4.9%',
      
      // Generate variations of primary
      primary: `${h} ${s}% ${l}%`,
      primaryForeground: `${h} ${s}% ${l > 50 ? 10 : 98}%`,
      
      // Generate complementary colors
      secondary: `${h} ${s * 0.5}% ${l + 10}%`,
      accent: `${(h + 30) % 360} ${s}% ${l}%`,
      
      // ... generate rest of palette
    }
  };
}
```

### 11.6 Brand Theme Example

```typescript
// Company brand theme
const companyTheme: Theme = {
  id: 'acme-corp',
  name: 'ACME Corp',
  colors: {
    // Brand colors from style guide
    primary: '210 100% 50%',        // Brand Blue
    primaryForeground: '0 0% 100%',
    
    secondary: '150 80% 40%',       // Brand Green
    secondaryForeground: '0 0% 100%',
    
    accent: '30 100% 50%',          // Accent Orange
    accentForeground: '0 0% 100%',
    
    // Neutral colors
    background: '0 0% 100%',
    foreground: '210 15% 20%',
    muted: '210 15% 96%',
    mutedForeground: '210 15% 45%',
    
    // Semantic colors
    destructive: '0 80% 50%',
    destructiveForeground: '0 0% 100%',
    
    // UI elements
    border: '210 15% 90%',
    input: '210 15% 95%',
    ring: '210 100% 50%',
  },
  typography: {
    sans: '"Roboto", sans-serif',
    mono: '"Roboto Mono", monospace',
  },
  radius: {
    sm: '0.25rem',
    md: '0.5rem',
    lg: '0.75rem',
  }
};
```

### 11.7 Theme Validation

```typescript
function validateTheme(theme: Theme): ValidationResult {
  const errors: string[] = [];
  
  // Check contrast ratios
  const bgFgContrast = calculateContrast(
    theme.colors.background,
    theme.colors.foreground
  );
  
  if (bgFgContrast < 4.5) {
    errors.push('Foreground/background contrast too low (< 4.5:1)');
  }
  
  const primaryContrast = calculateContrast(
    theme.colors.primary,
    theme.colors.primaryForeground
  );
  
  if (primaryContrast < 4.5) {
    errors.push('Primary color contrast too low');
  }
  
  // Check all required properties
  const requiredColors = [
    'background', 'foreground', 'primary', 'primaryForeground'
  ];
  
  for (const color of requiredColors) {
    if (!theme.colors[color]) {
      errors.push(`Missing required color: ${color}`);
    }
  }
  
  return {
    valid: errors.length === 0,
    errors
  };
}
```

### 11.8 Theme Switcher UI

```typescript
// Theme picker component
<div class="theme-picker">
  <button 
    data-theme="default"
    class="theme-option"
    aria-label="Default theme"
  >
    <div class="theme-preview">
      <span style="background: hsl(222.2 47.4% 11.2%)"></span>
      <span style="background: hsl(210 40% 96.1%)"></span>
      <span style="background: hsl(0 84.2% 60.2%)"></span>
    </div>
    <span>Default</span>
  </button>
  
  <button 
    data-theme="slate"
    class="theme-option"
    aria-label="Slate theme"
  >
    <div class="theme-preview">
      <span style="background: hsl(215.4 16.3% 46.9%)"></span>
      <span style="background: hsl(210 40% 96.1%)"></span>
      <span style="background: hsl(0 84.2% 60.2%)"></span>
    </div>
    <span>Slate</span>
  </button>
  
  <!-- More themes... -->
</div>

<style>
.theme-picker {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: var(--space-4);
}

.theme-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4);
  border: 2px solid var(--border);
  border-radius: var(--radius-md);
  cursor: pointer;
}

.theme-option[data-active="true"] {
  border-color: var(--primary);
}

.theme-preview {
  display: flex;
  gap: var(--space-1);
}

.theme-preview span {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-sm);
}
</style>
```

---

