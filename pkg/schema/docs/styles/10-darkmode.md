## 10. 🌓 Dark Mode

### 10.1 Dark Mode Philosophy

Dark mode isn't just inverted colors—it's a carefully designed alternative color scheme that:

1. **Reduces eye strain** in low-light environments
2. **Saves battery** on OLED displays
3. **Provides user choice** and personalization
4. **Maintains accessibility** and readability

### 10.2 Implementation Strategy

```css
/* Strategy 1: CSS Classes */
:root {
  /* Light mode (default) */
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
}

.dark {
  /* Dark mode */
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
}

/* Strategy 2: Data Attribute */
[data-theme="light"] {
  --background: 0 0% 100%;
}

[data-theme="dark"] {
  --background: 222.2 84% 4.9%;
}

/* Strategy 3: Media Query */
@media (prefers-color-scheme: dark) {
  :root {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
  }
}
```

### 10.3 Complete Color Tokens

```css
:root {
  /* Light mode colors */
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  
  --card: 0 0% 100%;
  --card-foreground: 222.2 84% 4.9%;
  
  --popover: 0 0% 100%;
  --popover-foreground: 222.2 84% 4.9%;
  
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  
  --border: 214.3 31.8% 91.4%;
  --input: 214.3 31.8% 91.4%;
  --ring: 222.2 84% 4.9%;
  
  --radius: 0.5rem;
}

.dark {
  /* Dark mode colors */
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  
  --card: 222.2 84% 4.9%;
  --card-foreground: 210 40% 98%;
  
  --popover: 222.2 84% 4.9%;
  --popover-foreground: 210 40% 98%;
  
  --primary: 210 40% 98%;
  --primary-foreground: 222.2 47.4% 11.2%;
  
  --secondary: 217.2 32.6% 17.5%;
  --secondary-foreground: 210 40% 98%;
  
  --muted: 217.2 32.6% 17.5%;
  --muted-foreground: 215 20.2% 65.1%;
  
  --accent: 217.2 32.6% 17.5%;
  --accent-foreground: 210 40% 98%;
  
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 210 40% 98%;
  
  --border: 217.2 32.6% 17.5%;
  --input: 217.2 32.6% 17.5%;
  --ring: 212.7 26.8% 83.9%;
}
```

### 10.4 Dark Mode JavaScript

```typescript
// Theme management
class ThemeManager {
  private theme: 'light' | 'dark' | 'system' = 'system';
  
  constructor() {
    this.init();
  }
  
  init() {
    // Load saved preference
    const saved = localStorage.getItem('theme');
    if (saved) {
      this.theme = saved as 'light' | 'dark' | 'system';
    }
    
    // Apply theme
    this.apply();
    
    // Listen for system changes
    window.matchMedia('(prefers-color-scheme: dark)')
      .addEventListener('change', () => {
        if (this.theme === 'system') {
          this.apply();
        }
      });
  }
  
  setTheme(theme: 'light' | 'dark' | 'system') {
    this.theme = theme;
    localStorage.setItem('theme', theme);
    this.apply();
  }
  
  apply() {
    const root = document.documentElement;
    
    if (this.theme === 'system') {
      const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      root.classList.toggle('dark', isDark);
    } else {
      root.classList.toggle('dark', this.theme === 'dark');
    }
  }
  
  getTheme() {
    return this.theme;
  }
}

// Usage
const themeManager = new ThemeManager();

// Toggle theme
document.getElementById('theme-toggle')?.addEventListener('click', () => {
  const current = themeManager.getTheme();
  const next = current === 'dark' ? 'light' : 'dark';
  themeManager.setTheme(next);
});
```

### 10.5 Dark Mode Considerations

#### **Avoid Pure Black**

```css
/* ❌ Bad: Pure black is harsh */
.dark {
  --background: 0 0% 0%;
}

/* ✅ Good: Slightly lighter for better readability */
.dark {
  --background: 222.2 84% 4.9%;  /* ~#0a0a0f */
}
```

#### **Reduce White Contrast**

```css
/* ❌ Bad: Pure white on dark */
.dark {
  --text: 0 0% 100%;
}

/* ✅ Good: Slightly dimmed white */
.dark {
  --text: 210 40% 98%;  /* ~#f9fafb */
}
```

#### **Adjust Shadows**

```css
/* Light mode shadows */
.card {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

/* Dark mode shadows */
.dark .card {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.5);
  /* Stronger shadows for depth */
}
```

#### **Elevation System**

```css
:root {
  --elevation-1: hsl(var(--background));
  --elevation-2: hsl(var(--muted) / 0.3);
  --elevation-3: hsl(var(--muted) / 0.5);
}

.dark {
  --elevation-1: hsl(222.2 84% 4.9%);    /* Base */
  --elevation-2: hsl(222.2 84% 6%);      /* Card */
  --elevation-3: hsl(222.2 84% 8%);      /* Modal */
}
```

### 10.6 Image Handling

```css
/* Reduce brightness of images in dark mode */
.dark img {
  opacity: 0.9;
}

/* Invert logos/icons */
.dark .logo {
  filter: invert(1) hue-rotate(180deg);
}

/* Use different images for dark mode */
<picture>
  <source srcset="logo-dark.svg" media="(prefers-color-scheme: dark)" />
  <img src="logo-light.svg" alt="Logo" />
</picture>
```

### 10.7 Color Semantic Consistency

Ensure colors maintain meaning across themes:

```css
/* Success color */
:root {
  --success: 142.1 76.2% 36.3%;  /* Green in light */
}

.dark {
  --success: 142.1 70.6% 45.3%;  /* Lighter green in dark */
}

/* Error color */
:root {
  --error: 0 84.2% 60.2%;  /* Red in light */
}

.dark {
  --error: 0 62.8% 50.6%;  /* Adjusted red in dark */
}
```

### 10.8 Testing Dark Mode

**Checklist:**
- [ ] All text meets contrast requirements (4.5:1)
- [ ] Interactive elements are visible
- [ ] Focus indicators are visible
- [ ] Shadows provide adequate depth
- [ ] Images don't look washed out
- [ ] Color meanings are preserved
- [ ] Test in actual dark environments
- [ ] Verify on OLED displays

---

