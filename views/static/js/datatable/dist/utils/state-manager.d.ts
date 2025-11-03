export declare class StateManager<T = any> {
    private storage;
    constructor();
    save(key: string, state: T): void;
    load<U = T>(key: string): U | null;
    remove(key: string): void;
    clear(): void;
    private replacer;
    private reviver;
}
//# sourceMappingURL=state-manager.d.ts.map