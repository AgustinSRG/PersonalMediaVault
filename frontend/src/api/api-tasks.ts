// Tasks API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class TasksAPI {
    public static GetTasks(): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/tasks"),
        };
    }

    public static GetTask(id: number): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/tasks/" + encodeURIComponent(id)),
        };
    }

    public static KillTask(id: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/tasks/" + encodeURIComponent(id) + "/kill"),
        };
    }
}
