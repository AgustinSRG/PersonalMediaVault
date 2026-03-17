/* Cookies */

"use strict";

/**
 * Sets assets session cookie
 * @param name The cookie name
 * @param value The cookie value
 */
export function setAssetsSessionCookie(name: string, value: string) {
    const date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    document.cookie = name + "=" + value + ";path=/assets/;SameSite=Strict";
}
