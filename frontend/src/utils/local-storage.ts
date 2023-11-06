// Local storage management

"use strict";

/**
 * Fetches a value from local storage
 * @param key The local storage key
 * @param defaultVal The default value
 * @returns The fetched value, or the default value
 */
export function fetchFromLocalStorage<T>(key: string, defaultVal: T): T {
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

/**
 * Saves a value in the local storage
 * @param key The local storage key
 * @param val The value to store
 */
export function saveIntoLocalStorage(key: string, val: any) {
    try {
        localStorage.setItem(key, JSON.stringify(val));
    } catch (ex) {
        console.error(ex);
    }
}
