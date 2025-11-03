/**
 * EventEmitter Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { EventEmitter } from '../src/utils/event-emitter';

interface TestEvents {
  'test:event': (data: string) => void;
  'test:number': (num: number) => void;
  'test:object': (obj: { id: number; name: string }) => void;
  'test:nodata': () => void;
}

describe('EventEmitter', () => {
  let emitter: EventEmitter<TestEvents>;

  beforeEach(() => {
    emitter = new EventEmitter<TestEvents>();
  });

  describe('Event Registration', () => {
    it('should register event listener', () => {
      const handler = vi.fn();
      emitter.on('test:event', handler);
      
      emitter.emit('test:event', 'hello');
      expect(handler).toHaveBeenCalledWith('hello');
    });

    it('should register multiple listeners for same event', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      emitter.on('test:event', handler1);
      emitter.on('test:event', handler2);
      
      emitter.emit('test:event', 'hello');
      
      expect(handler1).toHaveBeenCalledWith('hello');
      expect(handler2).toHaveBeenCalledWith('hello');
    });

    it('should register listeners for different events', () => {
      const stringHandler = vi.fn();
      const numberHandler = vi.fn();
      
      emitter.on('test:event', stringHandler);
      emitter.on('test:number', numberHandler);
      
      emitter.emit('test:event', 'hello');
      emitter.emit('test:number', 42);
      
      expect(stringHandler).toHaveBeenCalledWith('hello');
      expect(numberHandler).toHaveBeenCalledWith(42);
      expect(stringHandler).not.toHaveBeenCalledWith(42);
      expect(numberHandler).not.toHaveBeenCalledWith('hello');
    });

    it('should return unsubscribe function', () => {
      const handler = vi.fn();
      const unsubscribe = emitter.on('test:event', handler);
      
      expect(typeof unsubscribe).toBe('function');
      
      emitter.emit('test:event', 'before unsubscribe');
      expect(handler).toHaveBeenCalledTimes(1);
      
      unsubscribe();
      emitter.emit('test:event', 'after unsubscribe');
      expect(handler).toHaveBeenCalledTimes(1); // Should not be called again
    });
  });

  describe('Event Emission', () => {
    it('should emit event with string data', () => {
      const handler = vi.fn();
      emitter.on('test:event', handler);
      
      emitter.emit('test:event', 'test data');
      expect(handler).toHaveBeenCalledWith('test data');
    });

    it('should emit event with number data', () => {
      const handler = vi.fn();
      emitter.on('test:number', handler);
      
      emitter.emit('test:number', 123);
      expect(handler).toHaveBeenCalledWith(123);
    });

    it('should emit event with object data', () => {
      const handler = vi.fn();
      emitter.on('test:object', handler);
      
      const testObj = { id: 1, name: 'test' };
      emitter.emit('test:object', testObj);
      expect(handler).toHaveBeenCalledWith(testObj);
    });

    it('should emit event without data', () => {
      const handler = vi.fn();
      emitter.on('test:nodata', handler);
      
      emitter.emit('test:nodata');
      expect(handler).toHaveBeenCalledWith(undefined);
    });

    it('should not throw when emitting event with no listeners', () => {
      expect(() => {
        emitter.emit('test:event', 'no listeners');
      }).not.toThrow();
    });

    it('should handle errors in event handlers gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const errorHandler = vi.fn(() => {
        throw new Error('Handler error');
      });
      const normalHandler = vi.fn();
      
      emitter.on('test:event', errorHandler);
      emitter.on('test:event', normalHandler);
      
      emitter.emit('test:event', 'test');
      
      expect(errorHandler).toHaveBeenCalled();
      expect(normalHandler).toHaveBeenCalled(); // Should still be called despite error
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('EventEmitter: Error in test:event handler:'),
        expect.any(Error)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Event Removal', () => {
    it('should remove specific event listener', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      emitter.on('test:event', handler1);
      emitter.on('test:event', handler2);
      
      emitter.off('test:event', handler1);
      emitter.emit('test:event', 'test');
      
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).toHaveBeenCalledWith('test');
    });

    it('should handle removing non-existent listener gracefully', () => {
      const handler = vi.fn();
      
      expect(() => {
        emitter.off('test:event', handler);
      }).not.toThrow();
    });

    it('should remove all listeners for specific event', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      const handler3 = vi.fn();
      
      emitter.on('test:event', handler1);
      emitter.on('test:event', handler2);
      emitter.on('test:number', handler3);
      
      emitter.removeAllListeners('test:event');
      
      emitter.emit('test:event', 'test');
      emitter.emit('test:number', 42);
      
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).not.toHaveBeenCalled();
      expect(handler3).toHaveBeenCalledWith(42); // Different event should still work
    });

    it('should remove all listeners for all events', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      emitter.on('test:event', handler1);
      emitter.on('test:number', handler2);
      
      emitter.removeAllListeners();
      
      emitter.emit('test:event', 'test');
      emitter.emit('test:number', 42);
      
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).not.toHaveBeenCalled();
    });
  });

  describe('Clear All Events', () => {
    it('should clear all event listeners', () => {
      const handler1 = vi.fn();
      const handler2 = vi.fn();
      
      emitter.on('test:event', handler1);
      emitter.on('test:number', handler2);
      
      emitter.clear();
      
      emitter.emit('test:event', 'test');
      emitter.emit('test:number', 42);
      
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).not.toHaveBeenCalled();
    });
  });

  describe('Memory Management', () => {
    it('should not leak memory when adding and removing many listeners', () => {
      const handlers: Array<() => void> = [];
      
      // Add many listeners
      for (let i = 0; i < 1000; i++) {
        const handler = vi.fn();
        handlers.push(handler);
        emitter.on('test:nodata', handler);
      }
      
      // Remove all listeners
      handlers.forEach(handler => {
        emitter.off('test:nodata', handler);
      });
      
      // Should not have any listeners
      emitter.emit('test:nodata');
      handlers.forEach(handler => {
        expect(handler).not.toHaveBeenCalled();
      });
    });

    it('should handle rapid event emission', () => {
      const handler = vi.fn();
      emitter.on('test:number', handler);
      
      // Emit many events rapidly
      for (let i = 0; i < 1000; i++) {
        emitter.emit('test:number', i);
      }
      
      expect(handler).toHaveBeenCalledTimes(1000);
      expect(handler).toHaveBeenLastCalledWith(999);
    });
  });

  describe('Type Safety', () => {
    it('should maintain type safety for event data', () => {
      const stringHandler = vi.fn((data: string) => {
        expect(typeof data).toBe('string');
      });
      
      const numberHandler = vi.fn((data: number) => {
        expect(typeof data).toBe('number');
      });
      
      const objectHandler = vi.fn((data: { id: number; name: string }) => {
        expect(data).toHaveProperty('id');
        expect(data).toHaveProperty('name');
        expect(typeof data.id).toBe('number');
        expect(typeof data.name).toBe('string');
      });
      
      emitter.on('test:event', stringHandler);
      emitter.on('test:number', numberHandler);
      emitter.on('test:object', objectHandler);
      
      emitter.emit('test:event', 'hello');
      emitter.emit('test:number', 42);
      emitter.emit('test:object', { id: 1, name: 'test' });
      
      expect(stringHandler).toHaveBeenCalledWith('hello');
      expect(numberHandler).toHaveBeenCalledWith(42);
      expect(objectHandler).toHaveBeenCalledWith({ id: 1, name: 'test' });
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty event names', () => {
      // TypeScript should prevent this, but test runtime behavior
      const emptyEmitter = new EventEmitter<Record<string, any>>();
      const handler = vi.fn();
      
      emptyEmitter.on('', handler);
      emptyEmitter.emit('', 'test');
      
      expect(handler).toHaveBeenCalledWith('test');
    });

    it('should handle null/undefined data', () => {
      const handler = vi.fn();
      emitter.on('test:event', handler);
      
      emitter.emit('test:event', null as any);
      emitter.emit('test:event', undefined as any);
      
      expect(handler).toHaveBeenCalledWith(null);
      expect(handler).toHaveBeenCalledWith(undefined);
    });

    it('should handle same handler registered multiple times', () => {
      const handler = vi.fn();
      
      emitter.on('test:event', handler);
      emitter.on('test:event', handler); // Same handler again
      
      emitter.emit('test:event', 'test');
      
      // Handler should only be called once (Set behavior)
      expect(handler).toHaveBeenCalledTimes(1);
    });
  });
});