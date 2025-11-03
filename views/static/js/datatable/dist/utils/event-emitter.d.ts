export type EventCallback<T = any> = (data: T) => void;
export declare class EventEmitter<Events extends Record<string, (...args: any[]) => void>> {
    private events;
    on<K extends keyof Events>(event: K, callback: EventCallback<Parameters<Events[K]>[0]>): () => void;
    off<K extends keyof Events>(event: K, callback: EventCallback<Parameters<Events[K]>[0]>): void;
    emit<K extends keyof Events>(event: K, data?: any): void;
    clear(): void;
    removeAllListeners(event?: keyof Events): void;
}
//# sourceMappingURL=event-emitter.d.ts.map