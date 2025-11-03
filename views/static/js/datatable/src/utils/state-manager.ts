/**
 * State Manager
 * Handles state persistence to localStorage
 */

export class StateManager<T = any> {
  private storage: Storage | null = null;

  constructor() {
    // Check if localStorage is available
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        this.storage = window.localStorage;
      }
    } catch (e) {
      console.warn('StateManager: localStorage is not available');
    }
  }

  save(key: string, state: T): void {
    if (!this.storage) return;

    try {
      const serialized = JSON.stringify(state, this.replacer);
      this.storage.setItem(key, serialized);
    } catch (error) {
      console.error('StateManager: Error saving state:', error);
    }
  }

  load<U = T>(key: string): U | null {
    if (!this.storage) return null;

    try {
      const serialized = this.storage.getItem(key);
      if (!serialized) return null;

      return JSON.parse(serialized, this.reviver) as U;
    } catch (error) {
      console.error('StateManager: Error loading state:', error);
      return null;
    }
  }

  remove(key: string): void {
    if (!this.storage) return;

    try {
      this.storage.removeItem(key);
    } catch (error) {
      console.error('StateManager: Error removing state:', error);
    }
  }

  clear(): void {
    if (!this.storage) return;

    try {
      this.storage.clear();
    } catch (error) {
      console.error('StateManager: Error clearing state:', error);
    }
  }

  // Custom replacer to handle Sets, Maps, etc.
  private replacer(_key: string, value: any): any {
    if (value instanceof Set) {
      return {
        __type: 'Set',
        __value: Array.from(value),
      };
    }
    if (value instanceof Map) {
      return {
        __type: 'Map',
        __value: Array.from(value.entries()),
      };
    }
    if (value instanceof Date) {
      return {
        __type: 'Date',
        __value: value.toISOString(),
      };
    }
    return value;
  }

  // Custom reviver to restore Sets, Maps, etc.
  private reviver(_key: string, value: any): any {
    if (value && typeof value === 'object' && value.__type) {
      switch (value.__type) {
        case 'Set':
          return new Set(value.__value);
        case 'Map':
          return new Map(value.__value);
        case 'Date':
          return new Date(value.__value);
        default:
          return value;
      }
    }
    return value;
  }
}
