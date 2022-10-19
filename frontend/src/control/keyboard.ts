// Keyboard events manager

type KeyboardEventHandler = (event: KeyboardEvent) => boolean;

export class KeyboardManager {
    public static handlers: { priority: number, fn: KeyboardEventHandler }[] = [];

    public static Initialize() {
        document.addEventListener("keydown", KeyboardManager.Handle);
    }

    public static Handle(event: KeyboardEvent) {
        for (const handler of KeyboardManager.handlers) {
            if (handler.fn(event)) {
                event.preventDefault();
                return;
            }
        }
    }

    private static Sort() {
        KeyboardManager.handlers = KeyboardManager.handlers.sort((a, b) => {
            if (a.priority > b.priority) {
                return -1
            } else if (a.priority < b.priority) {
                return 1;
            } else {
                return 0;
            }
        });
    }

    public static AddHandler(handler: KeyboardEventHandler, priority?: number) {
        KeyboardManager.handlers.push({
            priority: priority || 0,
            fn: handler,
        });
        KeyboardManager.Sort();
    }

    public static RemoveHandler(handler: KeyboardEventHandler) {
        for (let i = 0; i < KeyboardManager.handlers.length; i++) {
            if (KeyboardManager.handlers[i].fn === handler) {
                KeyboardManager.handlers.splice(i, 1);
                return;
            }
        }
    }
}
