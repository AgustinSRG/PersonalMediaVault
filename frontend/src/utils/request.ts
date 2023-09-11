/* Request utility */

"use strict";

import { AuthController } from "@/control/auth";

interface StatusCodeError {
    status: number;
    body: string;
}

export function GetApiURL(path: string): string {
    if (import.meta.env.DEV) {
        return (import.meta.env.DEV_TEST_HOST || "http://localhost") + path;
    } else {
        return location.protocol + "//" + location.host + path;
    }
}

export function GetAssetURL(path: string): string {
    if (import.meta.env.DEV) {
        return (import.meta.env.DEV_TEST_HOST || "http://localhost") + path;
    } else {
        return location.protocol + "//" + location.host + path;
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

/* eslint-disable @typescript-eslint/no-unused-vars */
export interface RequestParams<Return_Type = any> {
    method: "GET" | "POST";
    url: string;
    json?: any;
    form?: FormData;
}

export class Request<Return_Type = any> {
    public static pending: { [key: string]: Request } = {};

    public static ErrorHandler(): RequestErrorHandler {
        return new RequestErrorHandler();
    }

    public static Do<Return_Type = any>(params: RequestParams<Return_Type>): Request<Return_Type> {
        if (params.method == "POST") {
            if (params.form) {
                return Request.PostFormData<Return_Type>(null, params.url, params.form);
            } else {
                return Request.PostJSON<Return_Type>(null, params.url, params.json || {});
            }
        } else {
            return Request.Get<Return_Type>(null, params.url);
        }
    }

    public static Pending<Return_Type = any>(key: string, params: RequestParams<Return_Type>): Request<Return_Type> {
        if (params.method == "POST") {
            if (params.form) {
                return Request.PostFormData<Return_Type>(key, params.url, params.form);
            } else {
                return Request.PostJSON<Return_Type>(key, params.url, params.json || {});
            }
        } else {
            return Request.Get<Return_Type>(key, params.url);
        }
    }

    public static Get<Return_Type = any>(key: string, url: string): Request<Return_Type> {
        Request.Abort(key); // Abort any other request
        const controller = new AbortController();

        const r = new Request<Return_Type>(url, controller, null);

        if (key) {
            Request.pending[key] = r;
        }

        const authToken = AuthController.Session;
        AuthController.RefreshAuthStatus();

        fetch(url, {
            method: "GET",
            signal: controller.signal,
            headers: {
                "x-session-token": authToken,
            },
        })
            .then((response) => {
                if (response.status !== 200) {
                    if (key) {
                        delete Request.pending[key];
                    }
                    response
                        .text()
                        .then((txt) => {
                            r._onRequestError({
                                status: response.status,
                                body: txt,
                            });
                        })
                        .catch(() => {
                            r._onRequestError({
                                status: 0,
                                body: "",
                            });
                        });
                    return;
                }

                if (response.headers.get("content-type") === "application/json") {
                    response
                        .json()
                        .then((data) => {
                            if (key) {
                                delete Request.pending[key];
                            }

                            r._onSuccess(data);
                        })
                        .catch((err) => {
                            if (key) {
                                delete Request.pending[key];
                            }
                            r._onUnexpectedError(err);
                        });
                } else {
                    response
                        .text()
                        .then((txt) => {
                            if (key) {
                                delete Request.pending[key];
                            }

                            r._onSuccess(txt);
                        })
                        .catch((err) => {
                            if (key) {
                                delete Request.pending[key];
                            }
                            r._onUnexpectedError(err);
                        });
                }
            })
            .catch((err) => {
                if (err.name === "AbortError") {
                    r._onCancel();
                } else {
                    if (key) {
                        delete Request.pending[key];
                    }
                    r._onRequestError({
                        status: 0,
                        body: "",
                    });
                }
            });

        return r;
    }

    public static PostJSON<Return_Type = any>(key: string, url: string, json: any): Request<Return_Type> {
        Request.Abort(key); // Abort any other request
        const controller = new AbortController();

        const r = new Request<Return_Type>(url, controller, null);

        if (key) {
            Request.pending[key] = r;
        }

        const authToken = AuthController.Session;
        AuthController.RefreshAuthStatus();

        fetch(url, {
            method: "POST",
            signal: controller.signal,
            headers: {
                "Content-Type": "application/json",
                "x-session-token": authToken,
            },
            body: JSON.stringify(json),
        })
            .then((response) => {
                if (response.status !== 200) {
                    if (key) {
                        delete Request.pending[key];
                    }
                    response
                        .text()
                        .then((txt) => {
                            r._onRequestError({
                                status: response.status,
                                body: txt,
                            });
                        })
                        .catch(() => {
                            r._onRequestError({
                                status: 0,
                                body: "",
                            });
                        });
                    return;
                }

                if (response.headers.get("content-type") === "application/json") {
                    response
                        .json()
                        .then((data) => {
                            if (key) {
                                delete Request.pending[key];
                            }

                            r._onSuccess(data);
                        })
                        .catch((err) => {
                            if (key) {
                                delete Request.pending[key];
                            }
                            r._onUnexpectedError(err);
                        });
                } else {
                    response
                        .text()
                        .then((txt) => {
                            if (key) {
                                delete Request.pending[key];
                            }

                            r._onSuccess(txt);
                        })
                        .catch((err) => {
                            if (key) {
                                delete Request.pending[key];
                            }
                            r._onUnexpectedError(err);
                        });
                }
            })
            .catch((err) => {
                if (err.name === "AbortError") {
                    r._onCancel();
                } else {
                    if (key) {
                        delete Request.pending[key];
                    }
                    r._onRequestError({
                        status: 0,
                        body: "",
                    });
                }
            });

        return r;
    }

    public static PostFormData<Return_Type = any>(key: string, url: string, form: FormData): Request<Return_Type> {
        Request.Abort(key); // Abort any other request

        const request = new XMLHttpRequest();

        const r = new Request<Return_Type>(url, null, request);

        if (key) {
            Request.pending[key] = r;
        }

        const authToken = AuthController.Session;
        AuthController.RefreshAuthStatus();

        request.onabort = () => {
            r._onCancel();
        };

        request.upload.onprogress = (evt) => {
            if (!evt.lengthComputable) {
                return;
            }

            r._onUploadProgress(evt.loaded || 0, evt.total || 0);
        };

        request.onreadystatechange = () => {
            if (request.readyState === XMLHttpRequest.DONE) {
                if (key) {
                    delete Request.pending[key];
                }

                if (request.status !== 200) {
                    r._onRequestError({
                        status: request.status,
                        body: request.responseText,
                    });
                    return;
                }

                let data: any = request.responseText;

                if (request.getResponseHeader("content-type") === "application/json") {
                    try {
                        data = JSON.parse(data);
                    } catch (ex) {
                        r._onUnexpectedError(ex);
                        return;
                    }
                }

                r._onSuccess(data);
            }
        };

        request.onerror = () => {
            if (key) {
                delete Request.pending[key];
            }
            r._onRequestError({
                status: 0,
                body: "",
            });
        };

        request.responseType = "text";

        // Open
        request.open("POST", url);

        // Set headers
        request.setRequestHeader("x-session-token", authToken);

        // Send form data
        request.send(form);

        return r;
    }

    public static Abort(key: string) {
        if (!key) {
            return;
        }
        if (Request.pending[key]) {
            Request.pending[key].abort();
            delete Request.pending[key];
        }
    }

    public url: string;
    private abortController: AbortController | null;
    private xhr: XMLHttpRequest | null;

    private _onSuccess: (response: any) => void;
    private _onCancel: () => void;
    private _onRequestError: (error: StatusCodeError) => void;
    private _onUnexpectedError: (error: Error) => void;

    private _onUploadProgress: (loaded: number, total: number) => void;

    constructor(url: string, abortController: AbortController | null, xhr: XMLHttpRequest | null) {
        this.url = url;
        this.abortController = abortController;
        this.xhr = xhr;
        this._onSuccess = function () {};
        this._onCancel = function () {};
        this._onRequestError = function () {};
        this._onUnexpectedError = function () {};
        this._onUploadProgress = function () {};
    }

    public abort() {
        if (this.abortController) {
            this.abortController.abort();
        }

        if (this.xhr) {
            this.xhr.onreadystatechange = null;
            this.xhr.abort();
        }
    }

    public onSuccess(fn: (response: Return_Type) => void): Request {
        this._onSuccess = fn;
        return this;
    }

    public onCancel(fn: () => void): Request {
        this._onCancel = fn;
        return this;
    }

    public onRequestError(fn: (err: StatusCodeError) => void): Request {
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

    public handle(error: StatusCodeError) {
        // Get error code from body

        let errorCode = "";
        const data = error.body;
        if (data) {
            try {
                const parsedData = JSON.parse(data);
                errorCode = parsedData.code || "";
            } catch (err) {
                errorCode = "";
            }
        }

        for (const callback of this.callbacks) {
            if (callback.status === "*" || callback.status === error.status) {
                if (callback.code === "*" || errorCode === callback.code) {
                    callback.callback();
                    return;
                }
            }
        }
    }
}
