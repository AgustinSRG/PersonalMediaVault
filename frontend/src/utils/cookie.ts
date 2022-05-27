/* Cookies */

"use strict"

export function getCookie(name: string): string {
    let value = "; " + document.cookie;
    let parts = value.split("; " + name + "=");
    if (parts.length == 2) return parts.pop().split(";").shift();
}

export function setCookie(name: string, value: string) {
    let date = new Date();
    date.setFullYear(date.getFullYear() + 1);
    document.cookie = name + "=" + value + ";expires=" + date.toString() + ";path=/";
}

export function getParameterByName(name: string, url?: string): string {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, '\\$&');
    var regex = new RegExp('[?&]' + name + '(=([^&#]*)|&|#|$)'),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, ' '));
}
