// API requests

"use strict";

import { RequestError } from "./request-error";

/**
 * Name of the header to send the session token into
 */
const SESSION_TOKEN_HEADER_NAME = "x-session-token";

/**
 * Authentication status for requests
 */
const RequestAuthentication = {
    /**
     * Session token
     */
    session: "",

    /**
     * Callback to refresh the authentication
     */
    refreshCallback: () => {},
};

/**
 * Sets authentication information for requests
 * @param session The session token
 * @param refreshCallback Callback to refresh the authentication
 */
export function setRequestAuthentication(session: string, refreshCallback: () => void) {
    RequestAuthentication.session = session;
    RequestAuthentication.refreshCallback = refreshCallback;
}

/**
 * API request method
 */
export type RequestMethod = "GET" | "POST" | "DELETE";

/**
 * Request parameters
 */
export interface RequestParams<Error_Handler> {
    /**
     * Request method
     */
    method: RequestMethod;

    /**
     * Request URL
     */
    url: string;

    /**
     * Request body (will be sent as JSON)
     */
    json?: any;

    /**
     * Request body as mutipart/form-data
     */
    form?: FormData;

    /**
     * Function to handle request errors
     */
    handleError: (err: RequestError, handler: Error_Handler) => void;
}

/**
 * API request
 */
export class Request<Return_Type, Error_Handler> {
    /**
     * Request method
     */
    public readonly method: RequestMethod;

    /**
     * Request URL
     */
    public readonly url: string;

    /**
     * Request body (will be sent as JSON)
     */
    public readonly json?: any;

    /**
     * Request body as multipart/form-data
     */
    public readonly form?: FormData;

    /**
     * Function to handle request errors
     */
    public readonly handleError: (err: RequestError, handler: Error_Handler) => void;

    /**
     * Abort controller
     */
    private abortController?: AbortController;

    /**
     * XMLHttpRequest
     */
    private xhr?: XMLHttpRequest;

    /**
     * Success callback
     */
    private _onSuccess: (response: Return_Type) => void;

    /**
     * Cancel callback
     */
    private _onCancel: () => void;

    /**
     * Request error callback
     */
    private _onRequestError: (error: RequestError, handleError: (err: RequestError, handler: Error_Handler) => void) => void;

    /**
     * Unexpected error callback
     */
    private _onUnexpectedError: (error: Error) => void;

    /**
     * Upload progress callback
     */
    private _onUploadProgress: (loaded: number, total: number) => void;

    /**
     * True if aborted
     */
    private aborted?: boolean;

    /**
     * Constructor
     * @param params Request parameters
     */
    constructor(params: RequestParams<Error_Handler>) {
        this.url = params.url;
        this.method = params.method;
        this.json = params.json;
        this.form = params.form;
        this.handleError = params.handleError;
    }

    /**
     * Sets the callback for success
     * @param fn Function to handle the data returned by the API
     * @returns Self
     */
    public onSuccess(fn: (response: Return_Type) => void): Request<Return_Type, Error_Handler> {
        this._onSuccess = fn;
        return this;
    }

    /**
     * Sets the callback for cancel
     * @param fn Function to handle the cancellation event
     * @returns Self
     */
    public onCancel(fn: () => void): Request<Return_Type, Error_Handler> {
        this._onCancel = fn;
        return this;
    }

    /**
     * Sets the callback for request error
     * @param fn Function to handle the error
     * @returns Self
     */
    public onRequestError(
        fn: (error: RequestError, handleError: (err: RequestError, handler: Error_Handler) => void) => void,
    ): Request<Return_Type, Error_Handler> {
        this._onRequestError = fn;
        return this;
    }

    /**
     * Sets the callback for unexpected error
     * @param fn Function to handle the error
     * @returns Self
     */
    public onUnexpectedError(fn: (err: Error) => void): Request<Return_Type, Error_Handler> {
        this._onUnexpectedError = fn;
        return this;
    }

    /**
     * Sets the callback for upload progress
     * @param fn Function to handle the upload progress
     * @returns Self
     */
    public onUploadProgress(fn: (loaded: number, total: number) => void): Request<Return_Type, Error_Handler> {
        this._onUploadProgress = fn;
        return this;
    }

    /**
     * Send the request
     * @param callback A function to call when the request is finished
     */
    public send(callback?: () => void) {
        if (this.method === "POST") {
            if (this.form) {
                this.sendAsMultipartFormData(callback);
            } else {
                this.sendAsFetch(callback);
            }
        } else {
            this.sendAsFetch(callback);
        }
    }

    /**
     * Aborts the request
     */
    public abort() {
        this.aborted = true;

        if (this.abortController) {
            this.abortController.abort();
        }

        if (this.xhr) {
            this.xhr.onreadystatechange = null;
            this.xhr.abort();
        }
    }

    /**
     * Sends the request with the 'fetch' method
     * @param callback A function to call when the request is finished
     */
    private sendAsFetch(callback?: () => void) {
        this.abortController = new AbortController();

        RequestAuthentication.refreshCallback();

        const headers: { [key: string]: string } = {};

        headers[SESSION_TOKEN_HEADER_NAME] = RequestAuthentication.session;

        if (this.json) {
            headers["Content-Type"] = "application/json";
        }

        return fetch(this.url, {
            method: this.method,
            signal: this.abortController.signal,
            headers,
            body: this.json ? JSON.stringify(this.json) : undefined,
        })
            .then((response) => {
                if (response.status !== 200) {
                    response
                        .text()
                        .then((txt) => {
                            if (!this.aborted) {
                                callback && callback();
                            }
                            this._onRequestError &&
                                this._onRequestError(
                                    {
                                        status: response.status,
                                        body: txt,
                                    },
                                    this.handleError,
                                );
                        })
                        .catch((err) => {
                            if (!this.aborted) {
                                callback && callback();
                            }

                            if (err.name === "AbortError") {
                                this._onCancel && this._onCancel();
                            } else {
                                this._onUnexpectedError && this._onUnexpectedError(err);
                            }
                        });
                    return;
                }

                if (response.headers.get("content-type") === "application/json") {
                    response
                        .json()
                        .then((data) => {
                            if (!this.aborted) {
                                callback && callback();
                            }

                            this._onSuccess && this._onSuccess(data);
                        })
                        .catch((err) => {
                            if (!this.aborted) {
                                callback && callback();
                            }

                            if (err.name === "AbortError") {
                                this._onCancel && this._onCancel();
                            } else {
                                this._onUnexpectedError && this._onUnexpectedError(err);
                            }
                        });
                } else {
                    response
                        .text()
                        .then((txt) => {
                            if (!this.aborted) {
                                callback && callback();
                            }

                            this._onSuccess && this._onSuccess(txt as Return_Type);
                        })
                        .catch((err) => {
                            if (!this.aborted) {
                                callback && callback();
                            }

                            if (err.name === "AbortError") {
                                this._onCancel && this._onCancel();
                            } else {
                                this._onUnexpectedError && this._onUnexpectedError(err);
                            }
                        });
                }
            })
            .catch((err) => {
                if (!this.aborted) {
                    callback && callback();
                }
                if (err.name === "AbortError") {
                    this._onCancel && this._onCancel();
                } else {
                    this._onRequestError &&
                        this._onRequestError(
                            {
                                status: 0,
                                body: "",
                            },
                            this.handleError,
                        );
                }
            });
    }

    /**
     * Sends the request with XMLHttpRequest for multipart requests
     * @param callback A function to call when the request is finished
     */
    private sendAsMultipartFormData(callback?: () => void) {
        const request = new XMLHttpRequest();

        RequestAuthentication.refreshCallback();

        request.onabort = () => {
            this._onCancel && this._onCancel();
        };

        request.upload.onprogress = (evt) => {
            if (!evt.lengthComputable) {
                return;
            }

            this._onUploadProgress && this._onUploadProgress(evt.loaded || 0, evt.total || 0);
        };

        request.onreadystatechange = () => {
            if (request.readyState === XMLHttpRequest.DONE) {
                if (!this.aborted) {
                    callback && callback();
                }

                if (request.status !== 200) {
                    this._onRequestError &&
                        this._onRequestError(
                            {
                                status: request.status,
                                body: request.responseText,
                            },
                            this.handleError,
                        );
                    return;
                }

                let data: any = request.responseText;

                if (request.getResponseHeader("content-type") === "application/json") {
                    try {
                        data = JSON.parse(data);
                    } catch (ex) {
                        this._onUnexpectedError && this._onUnexpectedError(ex);
                        return;
                    }
                }

                this._onSuccess && this._onSuccess(data);
            }
        };

        request.onerror = () => {
            if (!this.aborted) {
                callback && callback();
            }
            this._onRequestError &&
                this._onRequestError(
                    {
                        status: 0,
                        body: "",
                    },
                    this.handleError,
                );
        };

        request.responseType = "text";

        // Open
        request.open(this.method, this.url);

        // Set headers
        request.setRequestHeader(SESSION_TOKEN_HEADER_NAME, RequestAuthentication.session);

        // Send form data
        request.send(this.form);
    }
}

/**
 * Sends API request
 * @param r The request
 */
export function makeApiRequest(r: Request<any, any>) {
    r.send();
}
