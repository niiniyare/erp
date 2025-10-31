## 9. 🎬 Animation & Motion

### 9.1 Motion Philosophy

Animation serves purpose—it's not decoration. Good motion:

1. **Guides attention** — Draws eye to important changes
2. **Provides feedback** — Confirms user actions
3. **Shows relationships** — Connects cause and effect
4. **Maintains context** — Helps users track state changes
5. **Adds personality** — Makes interface feel alive

### 9.2 Motion Principles

#### **Duration**

Animation speed affects perceived quality:

```css
:root {
  /* Micro-interactions */
  --duration-instant: 50ms;   /* Hover feedback */
  --duration-fast:    100ms;  /* Button press */
  --duration-normal:  200ms;  /* Standard transitions */
  --duration-slow:    300ms;  /* Drawers, modals */
  --duration-slower:  500ms;  /* Page transitions */
}

/* Usage */
.button {
  transition: background-color var(--duration-fast);
}

.modal {
  transition: opacity var(--duration-normal);
}

.drawer {
  transition: transform var(--duration-slow);
}
```

**Duration Guidelines:**

| Duration | Use Case | Example |
|----------|----------|---------|
| 50ms | Instant feedback | Hover state |
| 100ms | Quick interactions | Button press, checkbox |
| 200ms | Standard transitions | Color change, opacity |
| 300ms | Larger elements | Modal, dropdown |
| 500ms+ | Complex animations | Page transitions, loaders |

#### **Easing**

Easing curves create natural motion:

```css
:root {
  /* Standard easings */
  --ease-linear:     linear;
  --ease-in:         cubic-bezier(0.4, 0, 1, 1);
  --ease-out:        cubic-bezier(0, 0, 0.2, 1);
  --ease-in-out:     cubic-bezier(0.4, 0, 0.2, 1);
  
  /* Custom easings */
  --ease-smooth:     cubic-bezier(0.4, 0, 0.2, 1);      /* Material Design */
  --ease-bounce:     cubic-bezier(0.68, -0.55, 0.27, 1.55);
  --ease-elastic:    cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

/* Usage */
.slide-in {
  animation: slide 300ms var(--ease-out);
}

.bounce {
  animation: bounce 500ms var(--ease-bounce);
}
```

**Easing Selection:**

- **Ease-out** — Entering elements (fast start, slow end)
- **Ease-in** — Exiting elements (slow start, fast end)
- **Ease-in-out** — Persistent elements (smooth both ends)

### 9.3 Common Animation Patterns

#### **Fade**

```css
@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes fade-out {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}

.fade-in {
  animation: fade-in var(--duration-normal) var(--ease-out);
}
```

#### **Slide**

```css
@keyframes slide-in-bottom {
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes slide-out-top {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(-100%);
    opacity: 0;
  }
}

.slide-in {
  animation: slide-in-bottom var(--duration-slow) var(--ease-out);
}
```

#### **Scale**

```css
@keyframes scale-in {
  from {
    transform: scale(0.95);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes scale-out {
  from {
    transform: scale(1);
    opacity: 1;
  }
  to {
    transform: scale(0.95);
    opacity: 0;
  }
}

.scale-in {
  animation: scale-in var(--duration-normal) var(--ease-out);
}
```

#### **Spin**

```css
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.spinner {
  animation: spin 1s linear infinite;
}
```

### 9.4 Transition Properties

```css
/* Single property transition */
.button {
  transition: background-color var(--duration-fast) var(--ease-out);
}

/* Multiple properties */
.card {
  transition: 
    transform var(--duration-normal) var(--ease-out),
    box-shadow var(--duration-normal) var(--ease-out);
}

/* All properties (use sparingly) */
.element {
  transition: all var(--duration-normal) var(--ease-out);
}

/* Delayed transition */
.delayed {
  transition: opacity var(--duration-normal) var(--ease-out) 100ms;
}
```

### 9.5 Loading States

```css
/* Skeleton loader */
@keyframes skeleton {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.skeleton {
  background: linear-gradient(
    90deg,
    var(--gray-200) 0%,
    var(--gray-100) 50%,
    var(--gray-200) 100%
  );
  background-size: 200% 100%;
  animation: skeleton 1.5s ease-in-out infinite;
}

/* Pulse */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.loading-pulse {
  animation: pulse 2s ease-in-out infinite;
}

/* Progress bar */
@keyframes progress {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(100%);
  }
}

.progress-indeterminate::after {
  content: '';
  position: absolute;
  width: 50%;
  height: 100%;
  background: var(--primary);
  animation: progress 1.5s ease-in-out infinite;
}
```

### 9.6 Micro-interactions

```css
/* Button press */
.button:active {
  transform: scale(0.98);
  transition: transform var(--duration-instant);
}

/* Checkbox check */
@keyframes check {
  0% {
    stroke-dashoffset: 24;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

.checkbox-icon {
  stroke-dasharray: 24;
  stroke-dashoffset: 24;
  transition: stroke-dashoffset var(--duration-normal) var(--ease-out);
}

.checkbox:checked + .checkbox-icon {
  stroke-dashoffset: 0;
}

/* Toggle switch */
.toggle-knob {
  transform: translateX(0);
  transition: transform var(--duration-normal) var(--ease-out);
}

.toggle:checked + .toggle-knob {
  transform: translateX(20px);
}
```

### 9.7 Page Transitions

```css
/* View transition API */
@view-transition {
  navigation: auto;
}

/* Customize transition */
::view-transition-old(root) {
  animation: fade-out var(--duration-normal) var(--ease-in);
}

::view-transition-new(root) {
  animation: fade-in var(--duration-normal) var(--ease-out);
}

/* Fallback for non-supporting browsers */
@supports not (view-transition-name: none) {
  .page-enter {
    animation: fade-in var(--duration-slower) var(--ease-out);
  }
  
  .page-exit {
    animation: fade-out var(--duration-slower) var(--ease-in);
  }
}
```

### 9.8 Scroll Animations

```css
/* Scroll-triggered animations */
.fade-in-on-scroll {
  opacity: 0;
  transform: translateY(20px);
  transition: 
    opacity var(--duration-slow) var(--ease-out),
    transform var(--duration-slow) var(--ease-out);
}

.fade-in-on-scroll.visible {
  opacity: 1;
  transform: translateY(0);
}

/* Intersection Observer implementation */
<script>
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('visible');
    }
  });
}, {
  threshold: 0.1
});

document.querySelectorAll('.fade-in-on-scroll').forEach(el => {
  observer.observe(el);
});
</script>
```

### 9.9 Performance Optimization

```css
/* Use transform and opacity for best performance */
/* ✅ Good: GPU-accelerated */
.element {
  transform: translateX(100px);
  opacity: 0.5;
}

/* ❌ Bad: Triggers layout/paint */
.element {
  left: 100px;
  margin-left: 20px;
}

/* Force GPU acceleration */
.accelerated {
  will-change: transform, opacity;
  /* Remove after animation completes */
}

/* Contain layout shifts */
.animated-container {
  contain: layout style paint;
}
```

### 9.10 Motion Best Practices

**Do:**
- ✅ Keep animations under 500ms (most cases)
- ✅ Use `transform` and `opacity` for performance
- ✅ Respect `prefers-reduced-motion`
- ✅ Provide animation purpose (feedback, guidance)
- ✅ Test on slower devices
- ✅ Use will-change sparingly

**Don't:**
- ❌ Animate width, height, or position (use transform)
- ❌ Animate too many elements simultaneously
- ❌ Use animations without purpose
- ❌ Ignore motion sensitivity
- ❌ Create distracting animations
- ❌ Forget to remove will-change after animation

---

