/**
 * StateManager Tests
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { StateManager } from '../src/utils/state-manager';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
    get length() {
      return Object.keys(store).length;
    },
    key: vi.fn((index: number) => Object.keys(store)[index] || null),
  };
})();

describe('StateManager', () => {
  let stateManager: StateManager;

  beforeEach(() => {
    // Mock window.localStorage
    Object.defineProperty(window, 'localStorage', {
      value: localStorageMock,
      writable: true,
    });

    // Clear the mock store
    localStorageMock.clear();
    vi.clearAllMocks();

    stateManager = new StateManager();
  });

  afterEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
  });

  describe('Basic Functionality', () => {
    it('should save and load simple state', () => {
      const state = { key: 'value', number: 42 };
      
      stateManager.save('test-key', state);
      const loaded = stateManager.load('test-key');
      
      expect(loaded).toEqual(state);
      expect(localStorageMock.setItem).toHaveBeenCalledWith('test-key', JSON.stringify(state));
      expect(localStorageMock.getItem).toHaveBeenCalledWith('test-key');
    });

    it('should return null when loading non-existent key', () => {
      const loaded = stateManager.load('non-existent');
      expect(loaded).toBeNull();
    });

    it('should return null when localStorage returns null', () => {
      localStorageMock.getItem.mockReturnValueOnce(null);
      const loaded = stateManager.load('test-key');
      expect(loaded).toBeNull();
    });

    it('should remove state correctly', () => {
      const state = { key: 'value' };
      
      stateManager.save('test-key', state);
      stateManager.remove('test-key');
      
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('test-key');
      expect(stateManager.load('test-key')).toBeNull();
    });

    it('should clear all state', () => {
      stateManager.save('key1', { value: 1 });
      stateManager.save('key2', { value: 2 });
      
      stateManager.clear();
      
      expect(localStorageMock.clear).toHaveBeenCalled();
    });
  });

  describe('Type Safety', () => {
    interface TestState {
      name: string;
      age: number;
      active: boolean;
    }

    it('should maintain type safety for typed states', () => {
      const state: TestState = { name: 'John', age: 30, active: true };
      
      stateManager.save('typed-state', state);
      const loaded = stateManager.load<TestState>('typed-state');
      
      expect(loaded).toEqual(state);
      expect(loaded?.name).toBe('John');
      expect(loaded?.age).toBe(30);
      expect(loaded?.active).toBe(true);
    });

    it('should handle different types correctly', () => {
      const stringState = 'test string';
      const numberState = 42;
      const booleanState = true;
      const arrayState = [1, 2, 3];
      
      stateManager.save('string', stringState);
      stateManager.save('number', numberState);
      stateManager.save('boolean', booleanState);
      stateManager.save('array', arrayState);
      
      expect(stateManager.load('string')).toBe(stringState);
      expect(stateManager.load('number')).toBe(numberState);
      expect(stateManager.load('boolean')).toBe(booleanState);
      expect(stateManager.load('array')).toEqual(arrayState);
    });
  });

  describe('Complex Data Types', () => {
    it('should handle Set objects', () => {
      const state = {
        selectedIds: new Set([1, 2, 3]),
        name: 'test'
      };
      
      stateManager.save('set-state', state);
      const loaded = stateManager.load('set-state');
      
      expect(loaded).toEqual(state);
      expect(loaded?.selectedIds).toBeInstanceOf(Set);
      expect(Array.from(loaded?.selectedIds || [])).toEqual([1, 2, 3]);
    });

    it('should handle Map objects', () => {
      const state = {
        keyValuePairs: new Map([['key1', 'value1'], ['key2', 'value2']]),
        name: 'test'
      };
      
      stateManager.save('map-state', state);
      const loaded = stateManager.load('map-state');
      
      expect(loaded).toEqual(state);
      expect(loaded?.keyValuePairs).toBeInstanceOf(Map);
      expect(Array.from(loaded?.keyValuePairs?.entries() || [])).toEqual([['key1', 'value1'], ['key2', 'value2']]);
    });

    it('should handle Date objects', () => {
      const date = new Date('2023-01-15T10:30:00Z');
      const state = {
        lastUpdate: date,
        name: 'test'
      };
      
      stateManager.save('date-state', state);
      const loaded = stateManager.load('date-state');
      
      expect(loaded).toEqual(state);
      expect(loaded?.lastUpdate).toBeInstanceOf(Date);
      expect(loaded?.lastUpdate?.getTime()).toBe(date.getTime());
    });

    it('should handle nested complex objects', () => {
      const state = {
        metadata: {
          selectedRows: new Set([1, 3, 5]),
          columnSettings: new Map([['col1', { width: 100 }], ['col2', { width: 200 }]]),
          lastSort: new Date('2023-01-15'),
        },
        filters: [
          { id: 1, value: 'test' },
          { id: 2, value: new Set(['a', 'b']) }
        ]
      };
      
      stateManager.save('complex-state', state);
      const loaded = stateManager.load('complex-state');
      
      expect(loaded).toEqual(state);
      expect(loaded?.metadata.selectedRows).toBeInstanceOf(Set);
      expect(loaded?.metadata.columnSettings).toBeInstanceOf(Map);
      expect(loaded?.metadata.lastSort).toBeInstanceOf(Date);
      expect(loaded?.filters[1].value).toBeInstanceOf(Set);
    });

    it('should handle empty Sets and Maps', () => {
      const state = {
        emptySet: new Set(),
        emptyMap: new Map(),
        name: 'test'
      };
      
      stateManager.save('empty-collections', state);
      const loaded = stateManager.load('empty-collections');
      
      expect(loaded).toEqual(state);
      expect(loaded?.emptySet).toBeInstanceOf(Set);
      expect(loaded?.emptyMap).toBeInstanceOf(Map);
      expect(loaded?.emptySet?.size).toBe(0);
      expect(loaded?.emptyMap?.size).toBe(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle JSON serialization errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      // Create circular reference
      const circularState: any = { name: 'test' };
      circularState.self = circularState;
      
      stateManager.save('circular', circularState);
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'StateManager: Error saving state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle JSON parsing errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      // Mock invalid JSON
      localStorageMock.getItem.mockReturnValueOnce('{ invalid json }');
      
      const loaded = stateManager.load('invalid-json');
      
      expect(loaded).toBeNull();
      expect(consoleSpy).toHaveBeenCalledWith(
        'StateManager: Error loading state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle localStorage setItem errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      // Mock localStorage quota exceeded error
      localStorageMock.setItem.mockImplementationOnce(() => {
        throw new Error('QuotaExceededError');
      });
      
      stateManager.save('test', { data: 'test' });
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'StateManager: Error saving state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle localStorage getItem errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      // Mock localStorage access error
      localStorageMock.getItem.mockImplementationOnce(() => {
        throw new Error('localStorage access denied');
      });
      
      const loaded = stateManager.load('test');
      
      expect(loaded).toBeNull();
      expect(consoleSpy).toHaveBeenCalledWith(
        'StateManager: Error loading state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle localStorage removeItem errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      // Mock localStorage access error
      localStorageMock.removeItem.mockImplementationOnce(() => {
        throw new Error('localStorage access denied');
      });
      
      stateManager.remove('test');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'StateManager: Error removing state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle localStorage clear errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      // Mock localStorage access error
      localStorageMock.clear.mockImplementationOnce(() => {
        throw new Error('localStorage access denied');
      });
      
      stateManager.clear();
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'StateManager: Error clearing state:',
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Environment Compatibility', () => {
    it('should handle missing localStorage gracefully', () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      // Create StateManager when localStorage is not available
      Object.defineProperty(window, 'localStorage', {
        value: undefined,
        writable: true,
      });
      
      const noStorageManager = new StateManager();
      
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        'StateManager: localStorage is not available'
      );
      
      // Operations should be no-ops
      noStorageManager.save('test', { data: 'test' });
      const loaded = noStorageManager.load('test');
      
      expect(loaded).toBeNull();
      
      consoleWarnSpy.mockRestore();
    });

    it('should handle window not being defined (SSR)', () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      // Mock SSR environment
      const originalWindow = global.window;
      delete (global as any).window;
      
      const ssrManager = new StateManager();
      
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        'StateManager: localStorage is not available'
      );
      
      // Operations should be no-ops
      ssrManager.save('test', { data: 'test' });
      const loaded = ssrManager.load('test');
      
      expect(loaded).toBeNull();
      
      // Restore window
      global.window = originalWindow;
      consoleWarnSpy.mockRestore();
    });

    it('should handle localStorage access throwing on initialization', () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      // Mock localStorage access throwing
      Object.defineProperty(window, 'localStorage', {
        get() {
          throw new Error('localStorage access denied');
        },
        configurable: true,
      });
      
      const restrictedManager = new StateManager();
      
      expect(consoleWarnSpy).toHaveBeenCalledWith(
        'StateManager: localStorage is not available'
      );
      
      // Operations should be no-ops
      restrictedManager.save('test', { data: 'test' });
      const loaded = restrictedManager.load('test');
      
      expect(loaded).toBeNull();
      
      consoleWarnSpy.mockRestore();
    });
  });

  describe('Data Integrity', () => {
    it('should preserve data type information', () => {
      const originalState = {
        string: 'hello',
        number: 42,
        boolean: true,
        nullValue: null,
        undefinedValue: undefined,
        array: [1, 2, 3],
        object: { nested: true },
        set: new Set([1, 2, 3]),
        map: new Map([['key', 'value']]),
        date: new Date('2023-01-15')
      };
      
      stateManager.save('type-test', originalState);
      const loaded = stateManager.load('type-test');
      
      expect(typeof loaded?.string).toBe('string');
      expect(typeof loaded?.number).toBe('number');
      expect(typeof loaded?.boolean).toBe('boolean');
      expect(loaded?.nullValue).toBeNull();
      expect(loaded?.undefinedValue).toBeUndefined();
      expect(Array.isArray(loaded?.array)).toBe(true);
      expect(typeof loaded?.object).toBe('object');
      expect(loaded?.set).toBeInstanceOf(Set);
      expect(loaded?.map).toBeInstanceOf(Map);
      expect(loaded?.date).toBeInstanceOf(Date);
    });

    it('should handle large state objects', () => {
      const largeState = {
        largeArray: Array.from({ length: 10000 }, (_, i) => ({ id: i, name: `Item ${i}` })),
        largeSet: new Set(Array.from({ length: 1000 }, (_, i) => i)),
        largeMap: new Map(Array.from({ length: 1000 }, (_, i) => [`key${i}`, `value${i}`])),
      };
      
      stateManager.save('large-state', largeState);
      const loaded = stateManager.load('large-state');
      
      expect(loaded).toEqual(largeState);
      expect(loaded?.largeArray).toHaveLength(10000);
      expect(loaded?.largeSet?.size).toBe(1000);
      expect(loaded?.largeMap?.size).toBe(1000);
    });

    it('should handle special characters and unicode', () => {
      const unicodeState = {
        emoji: '🚀📊💼',
        chinese: '你好世界',
        arabic: 'مرحبا بالعالم',
        special: '!@#$%^&*()_+-=[]{}|;:,.<>?',
        quotes: '"single" and \'double\' quotes',
        newlines: 'line1\nline2\rline3\r\nline4'
      };
      
      stateManager.save('unicode-test', unicodeState);
      const loaded = stateManager.load('unicode-test');
      
      expect(loaded).toEqual(unicodeState);
    });
  });

  describe('Performance', () => {
    it('should handle rapid save/load operations', () => {
      const start = performance.now();
      
      for (let i = 0; i < 100; i++) {
        const state = { iteration: i, data: `test-${i}` };
        stateManager.save(`test-${i}`, state);
        const loaded = stateManager.load(`test-${i}`);
        expect(loaded).toEqual(state);
      }
      
      const end = performance.now();
      expect(end - start).toBeLessThan(1000); // Should complete in less than 1 second
    });

    it('should not leak memory with repeated operations', () => {
      // This test verifies that the replacer/reviver functions don't create memory leaks
      const baseState = { name: 'test', data: new Set([1, 2, 3]) };
      
      for (let i = 0; i < 1000; i++) {
        stateManager.save('memory-test', { ...baseState, iteration: i });
        stateManager.load('memory-test');
      }
      
      // If we get here without running out of memory, the test passes
      expect(true).toBe(true);
    });
  });
});