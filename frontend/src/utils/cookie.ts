/* Cookies */

"use strict";

export function setAssetsSessionCookie(name: string, value: string) {
    const date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    document.cookie = name + "=" + value + ";path=/assets/;SameSite=Strict";
}

export function getCookie(name: string): string {
    const value = "; " + document.cookie;
    const parts = value.split("; " + name + "=");
    if (parts.length == 2) return parts.pop().split(";").shift();
}

export function getParameterByName(name: string, url?: string): string {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    const regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return "";
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}
