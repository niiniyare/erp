/**
 * Event Emitter
 * Simple pub/sub implementation for DataTable events
 */

export type EventCallback<T = any> = (data: T) => void;

export class EventEmitter<Events extends Record<string, (...args: any[]) => void>> {
  private events: Map<keyof Events, Set<EventCallback>> = new Map();

  on<K extends keyof Events>(event: K, callback: EventCallback<Parameters<Events[K]>[0]>): () => void {
    if (!this.events.has(event)) {
      this.events.set(event, new Set());
    }

    this.events.get(event)!.add(callback);

    // Return unsubscribe function
    return () => this.off(event, callback);
  }

  off<K extends keyof Events>(event: K, callback: EventCallback<Parameters<Events[K]>[0]>): void {
    const callbacks = this.events.get(event);
    if (callbacks) {
      callbacks.delete(callback);
    }
  }

  emit<K extends keyof Events>(event: K, data?: any): void {
    const callbacks = this.events.get(event);
    if (callbacks) {
      callbacks.forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error(`EventEmitter: Error in ${String(event)} handler:`, error);
        }
      });
    }
  }

  clear(): void {
    this.events.clear();
  }

  removeAllListeners(event?: keyof Events): void {
    if (event) {
      this.events.delete(event);
    } else {
      this.events.clear();
    }
  }
}
