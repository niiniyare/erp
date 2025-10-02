// Global type definitions for web hooks

// Alpine.js types
interface AlpineComponent {
    $watch?: (property: string, callback: () => void, options?: { deep?: boolean }) => void;
    $dispatch?: (event: string, detail?: any) => void;
    $el?: HTMLElement;
}

// Node.js types for timeout
declare namespace NodeJS {
    interface Timeout {}
}

// Window extensions
declare global {
    interface Window {
        Alpine?: {
            store: (name: string) => any;
        };
        APP_CONFIG: {
            api: {
                baseURL: string;
                timeout: number;
                headers: Record<string, string>;
            };
            app: {
                name: string;
                version: string;
                description: string;
            };
            theme: string;
            features: Record<string, boolean>;
            menu: Array<{
                label: string;
                icon: string;
                url: string;
                children: Array<{
                    label: string;
                    url: string;
                    children: any[];
                }>;
            }>;
            defaultTenant: string | null;
            pagination: {
                pageSize: number;
                showSizeChanger: boolean;
                showQuickJumper: boolean;
            };
        };
        API: {
            getHeaders: (tenantId?: string | null) => Record<string, string>;
            url: (path: string) => string;
            handleResponse: (response: Response) => Promise<any>;
            fetch: (url: string, options?: RequestInit) => Promise<any>;
        };
        ERPAdmin: any;
        useDataTable: (config?: any) => any;
        DataTableConfigs: any;
        DataTableVersion: string;
        createConsoleDataTable: (config?: any) => any;
        createWorkspaceDataTable: (config?: any) => any;
        createPortalDataTable: (config?: any) => any;
    }
}

export {};