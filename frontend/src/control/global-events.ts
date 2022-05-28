// Global events manager

export class GlobalEvents {
    public static events: {[key: string]: Function[]} = {};

    public static AddEventListener(eventName: string, handler: Function) {
        if (!this.events[eventName]) {
            this.events[eventName] = [];
        }
        this.events[eventName].push(handler);
    }

    public static Emit(eventName: string, ...args: string[]) {
        if (this.events[eventName]) {
            for (let handler of this.events[eventName]) {
                try {
                    handler(...args);
                } catch (ex) {
                    console.error(ex);
                }
            }
        }
    }

    public static RemoveEventListener(eventName: string, handler: Function) {
        if (!this.events[eventName]) {
            return;
        }
        let i = this.events[eventName].indexOf(handler);
        if (i >= 0) {
            this.events[eventName].splice(i, 1);
            if (this.events[eventName].length === 0) {
                delete this.events[eventName];
            }
        }
    }
}
