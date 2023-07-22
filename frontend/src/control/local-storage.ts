// Local storage management

export class LocalStorage {
    public static Get(key: string, defaultVal: any): any {
        try {
            const v = localStorage.getItem(key);

            if (v === undefined || v === null) {
                return defaultVal;
            }

            return JSON.parse(v);
        } catch (ex) {
            console.error(ex);
            return defaultVal;
        }
    }

    public static Set(key: string, val: any) {
        try {
            localStorage.setItem(key, JSON.stringify(val));
        } catch (ex) {
            console.error(ex);
        }
    }
}
