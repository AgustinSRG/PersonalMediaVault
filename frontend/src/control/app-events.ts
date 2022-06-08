// Global events manager

type CallbackFunctionVariadic = (...args: any[]) => void;

export class AppEvents {
    public static events: {[key: string]: CallbackFunctionVariadic[]} = {};

    public static AddEventListener(eventName: string, handler: CallbackFunctionVariadic) {
        if (!AppEvents.events[eventName]) {
            AppEvents.events[eventName] = [];
        }
        AppEvents.events[eventName].push(handler);
    }

    public static Emit(eventName: string, ...args: any[]) {
        if (AppEvents.events[eventName]) {
            for (const handler of AppEvents.events[eventName]) {
                try {
                    handler(...args);
                } catch (ex) {
                    console.error(ex);
                }
            }
        }
    }

    public static RemoveEventListener(eventName: string, handler: CallbackFunctionVariadic) {
        if (!AppEvents.events[eventName]) {
            return;
        }
        const i = AppEvents.events[eventName].indexOf(handler);
        if (i >= 0) {
            AppEvents.events[eventName].splice(i, 1);
            if (AppEvents.events[eventName].length === 0) {
                delete AppEvents.events[eventName];
            }
        }
    }
}
