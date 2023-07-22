// Timeouts manager

export class Timeouts {
    public static pending: { [key: string]: number } = {};

    public static Set(key: string, ms: number, handler: () => void) {
        Timeouts.Abort(key);
        Timeouts.pending[key] = setTimeout(() => {
            delete Timeouts.pending[key];
            handler();
        }, ms);
    }

    public static Abort(key: string) {
        if (Timeouts.pending[key] !== undefined) {
            clearTimeout(Timeouts.pending[key]);
            delete Timeouts.pending[key];
        }
    }
}
