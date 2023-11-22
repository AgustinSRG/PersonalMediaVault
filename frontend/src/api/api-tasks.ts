// Tasks API

"use strict";

import { RequestParams } from "@asanrom/request-browser";
import { TaskStatus } from "./models";
import { getApiURL } from "@/utils/api";

export class TasksAPI {
    public static GetTasks(): RequestParams<TaskStatus[]> {
        return {
            method: "GET",
            url: getApiURL("/api/tasks"),
        };
    }

    public static GetTask(id: number): RequestParams<TaskStatus> {
        return {
            method: "GET",
            url: getApiURL("/api/tasks/" + encodeURIComponent(id)),
        };
    }
}
