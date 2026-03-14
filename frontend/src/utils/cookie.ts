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

/**
 * Gets query parameter by name from the URL
 * @param name The parameter name
 * @param url The URL (if not provided, it will use the current location)
 * @returns The parameter value
 */
export function getParameterByName(name: string, url?: string): string {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    const regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return "";
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}
