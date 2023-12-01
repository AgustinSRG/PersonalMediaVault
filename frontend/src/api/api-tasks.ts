// Tasks API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { TaskStatus } from "./models";
import { getApiURL } from "@/utils/api";

/**
 * Gets tasks list
 * @returns The request parameters
 */
export function apiTasksGetTasks(): RequestParams<TaskStatus[], CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL("/api/tasks"),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for get task API
 */
export type GetTaskErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Gets task status
 * @param id The task ID
 * @returns The request parameters
 */
export function apiTasksGetTask(id: number): RequestParams<TaskStatus, GetTaskErrorHandler> {
    return {
        method: "GET",
        url: getApiURL("/api/tasks/" + encodeURIComponent(id)),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
