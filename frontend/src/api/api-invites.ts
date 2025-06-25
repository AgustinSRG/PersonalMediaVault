// About API

"use strict";

import type { CommonAuthenticatedErrorHandler, CommonErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import { getApiURL, API_PREFIX } from "@/utils/api";
import type { LoginResult, SessionDuration } from "./api-auth";

const API_GROUP_PREFIX = "/invites";

/**
 * Invite code status
 */
export interface InviteCodeStatus {
    /**
     * True if the user has a generated invite code
     */
    has_code: boolean;

    /**
     * The invite code
     */
    code: string;

    /**
     * The session duration if the code is used (Milliseconds)
     */
    duration: number;

    /**
     * Remaining time until code expiration (Milliseconds) (If <=0, the code has expired)
     */
    expiration_remaining: number;
}

/**
 * Error handler for invites API
 */
export type InvitesApiErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;
};

/**
 * Gets invite code status for the current user
 * @returns The request parameters
 */
export function apiInvitesGetStatus(): RequestParams<InviteCodeStatus, InvitesApiErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Invite session
 */
export interface InviteSession {
    /**
     * Unique session index
     */
    index: number;

    /**
     * Session timestamp (Unix milliseconds)
     */
    timestamp: number;

    /**
     * Expiration timestamp (Unix Milliseconds)
     */
    expiration: number;
}

/**
 * Gets list of invited sessions
 * @returns The request parameters
 */
export function apiInvitesGetSessions(): RequestParams<InviteSession[], InvitesApiErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/sessions`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Deletes an invited session
 * @param index Session index
 * @returns The request parameters
 */
export function apiInvitesDeleteSession(index: number): RequestParams<void, InvitesApiErrorHandler> {
    return {
        method: "DELETE",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/sessions/${index}`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for invite code generation API
 */
export type GenerateInviteCodeErrorHandler = InvitesApiErrorHandler & {
    /**
     * Error: Limit reached
     */
    limitReached: () => void;
};

/**
 * Generates an invite code. If another code exits, it will be replaced by a new one.
 * @param duration Session duration
 * @returns The request parameters
 */
export function apiInvitesGenerateCode(duration: SessionDuration): RequestParams<InviteCodeStatus, GenerateInviteCodeErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/generate`),
        json: {
            duration: duration,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "*", handler.limitReached)
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Clears the current invite code
 * @returns The request parameters
 */
export function apiInvitesClearCode(): RequestParams<void, InvitesApiErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/clear`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for invite code login
 */
export type InviteCodeLoginErrorHandler = CommonErrorHandler & {
    /**
     * Error: Invalid code
     */
    invalidCode: () => void;

    /**
     * Error: Wring code
     */
    wrongCode: () => void;

    /**
     * Error: Cooldown
     */
    cooldown: () => void;
};

/**
 * Creates an active session from an invite code
 * @param code Invite code
 * @returns The request parameters
 */
export function apiInvitesLogin(code: string): RequestParams<LoginResult, InviteCodeLoginErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/login`),
        json: {
            code: code,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "*", handler.invalidCode)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "*", handler.wrongCode)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
