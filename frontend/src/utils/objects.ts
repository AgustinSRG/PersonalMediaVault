// Object utils

export function copyObject(o: any): any {
    return JSON.parse(JSON.stringify(o));
}
