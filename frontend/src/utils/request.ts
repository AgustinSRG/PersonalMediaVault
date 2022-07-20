/* Request utility */

"use strict";

import { AuthController } from "@/control/auth";
import axios, { AxiosError } from "axios";

export function GetAPIURL(path: string): string {
    if (process.env.NODE_ENV === 'development') {
        return (process.env.DEV_TEST_HOST || "http://localhost") + path;
    } else {
        return "." + path;
    }
}

export function GetAssetURL(path: string): string {
    if (process.env.NODE_ENV === 'development') {
        return (process.env.DEV_TEST_HOST || "http://localhost") + path;
    } else {
        return "." + path;
    }
}

export function GenerateURIQuery(params: any): string {
    const keys = Object.keys(params);
    if (keys.length === 0) {
        return "";
    }

    let result = "";

    for (const key of keys) {
        if (!params[key]) {
            continue;
        }

        if (result !== "") {
            result += "&";
        } else {
            result += "?";
        }

        result += encodeURIComponent(key) + "=" + encodeURIComponent(params[key]);
    }

    return result;
}

export interface RequestParams {
    method: "GET" | "POST";
    url: string;
    json?: any;
    form?: FormData;
}

export class Request {

    public static pending: { [key: string]: Request } = {};

    public static ErrorHandler(): RequestErrorHandler {
        return new RequestErrorHandler();
    }

    public static Do(params: RequestParams): Request {
        if (params.method == "POST") {
            if (params.form) {
                return Request.PostFormData(null, params.url, params.form);
            } else {
                return Request.PostJSON(null, params.url, params.json || {});
            }
        } else {
            return Request.Get(null, params.url);
        }
    }

    public static Pending(key: string, params: RequestParams): Request {
        if (params.method == "POST") {
            if (params.form) {
                return Request.PostFormData(key, params.url, params.form);
            } else {
                return Request.PostJSON(key, params.url, params.json || {});
            }
        } else {
            return Request.Get(key, params.url);
        }
    }

    public static Get(key: string, url: string): Request {
        Request.Abort(key); // Abort any other request
        const controller = new AbortController();

        const r = new Request(url, controller);

        if (key) {
            Request.pending[key] = r;
        }

        const authToken = AuthController.Session;

        axios.get(url, {
            signal: controller.signal,
            headers: {
                "x-session-token": authToken,
            },
        }).then(response => {
            if (key) {
                delete Request.pending[key];
            }

            r._onSuccess(response.data)
        }).catch(thrown => {
            if (axios.isCancel(thrown)) {
                r._onCancel();
            } else if (axios.isAxiosError(thrown)) {
                if (key) {
                    delete Request.pending[key];
                }
                r._onRequestError(thrown);
            } else {
                if (key) {
                    delete Request.pending[key];
                }
                r._onUnexpectedError(thrown);
            }
        })

        return r;
    }

    public static PostJSON(key: string, url: string, json: any): Request {
        Request.Abort(key); // Abort any other request
        const controller = new AbortController();

        const r = new Request(url, controller);

        if (key) {
            Request.pending[key] = r;
        }

        const authToken = AuthController.Session;

        axios.post(url, json, {
            signal: controller.signal,
            headers: {
                "Content-Type": "application/json",
                "x-session-token": authToken,
            },
        }).then(response => {
            if (key) {
                delete Request.pending[key];
            }

            r._onSuccess(response.data)
        }).catch(thrown => {
            if (axios.isCancel(thrown)) {
                r._onCancel();
            } else if (axios.isAxiosError(thrown)) {
                if (key) {
                    delete Request.pending[key];
                }
                r._onRequestError(thrown);
            } else {
                if (key) {
                    delete Request.pending[key];
                }
                r._onUnexpectedError(thrown);
            }
        })

        return r;
    }

    public static PostFormData(key: string, url: string, form: FormData): Request {
        Request.Abort(key); // Abort any other request
        const controller = new AbortController();

        const r = new Request(url, controller);

        if (key) {
            Request.pending[key] = r;
        }

        const authToken = AuthController.Session;

        axios.post(url, form, {
            signal: controller.signal,
            headers: {
                "Content-Type": "multipart/form-data",
                "x-session-token": authToken,
            },
            onUploadProgress: progressEvent => {
                r._onUploadProgress(progressEvent.loaded || 0, progressEvent.total || 0);
            }
        }).then(response => {
            if (key) {
                delete Request.pending[key];
            }

            r._onSuccess(response.data)
        }).catch(thrown => {
            if (axios.isCancel(thrown)) {
                r._onCancel();
            } else if (axios.isAxiosError(thrown)) {
                if (key) {
                    delete Request.pending[key];
                }
                r._onRequestError(thrown);
            } else {
                if (key) {
                    delete Request.pending[key];
                }
                r._onUnexpectedError(thrown);
            }
        })

        return r;
    }

    public static Abort(key: string) {
        if (!key) {
            return;
        }
        if (Request.pending[key]) {
            Request.pending[key].abortController.abort();
            delete Request.pending[key];
        }
    }

    public url: string;
    public abortController: AbortController;

    private _onSuccess: (response: any) => void;
    private _onCancel: () => void;
    private _onRequestError: (error: AxiosError) => void;
    private _onUnexpectedError: (error: Error) => void;

    private _onUploadProgress: (loaded: number, total: number) => void;

    constructor(url: string, abortController: AbortController) {
        this.url = url;
        this.abortController = abortController;
        this._onSuccess = function () { };
        this._onCancel = function () { };
        this._onRequestError = function () { };
        this._onUnexpectedError = function () { };
        this._onUploadProgress = function () { };
    }

    public onSuccess(fn: (response: any) => void): Request {
        this._onSuccess = fn;
        return this;
    }

    public onCancel(fn: () => void): Request {
        this._onCancel = fn;
        return this;
    }

    public onRequestError(fn: (err: AxiosError) => void): Request {
        this._onRequestError = fn;
        return this;
    }

    public onUnexpectedError(fn: (err: Error) => void): Request {
        this._onUnexpectedError = fn;
        return this;
    }

    public onUploadProgress(fn: (loaded: number, total: number) => void): Request {
        this._onUploadProgress = fn;
        return this;
    }
}

interface RequestErrorHandlerCallback {
    status: number | string;
    code: string;
    callback: () => void;
}

export class RequestErrorHandler {
    private callbacks: RequestErrorHandlerCallback[];

    constructor() {
        this.callbacks = [];
    }

    public add(status: number | string, code: string, callback: () => void): RequestErrorHandler {
        this.callbacks.push({
            status: status,
            code: code,
            callback: callback,
        });

        return this;
    }

    public handle(error: AxiosError) {
        if (!error.response) {
            return;
        }
        if (error.response.status === 0 && error.response.statusText === "abort") {
            return;
        }

        // Get error code from body

        let errorCode = "";
        const data = error.response.data;
        if (data) {
            try {
                let parsedData: any;

                if (typeof data === "string") {
                    parsedData = JSON.parse(data);
                } else {
                    parsedData = data;
                }

                errorCode = parsedData.code || "";
            } catch (err) {
                errorCode = "";
            }
        }

        for (const callback of this.callbacks) {
            if (callback.status === "*" || callback.status === error.response.status) {
                if (callback.code === "*" || errorCode === callback.code) {
                    callback.callback();
                    return;
                }
            }
        }
    }
}
