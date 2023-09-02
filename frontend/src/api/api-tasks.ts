// Tasks API

import { GetApiURL, RequestParams } from "@/utils/request";
import { TaskStatus } from "./models";

export class TasksAPI {
    public static GetTasks(): RequestParams<TaskStatus[]> {
        return {
            method: "GET",
            url: GetApiURL("/api/tasks"),
        };
    }

    public static GetTask(id: number): RequestParams<TaskStatus> {
        return {
            method: "GET",
            url: GetApiURL("/api/tasks/" + encodeURIComponent(id)),
        };
    }
}
