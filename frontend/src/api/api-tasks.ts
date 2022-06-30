// Tasks API

import { GetAPIURL, RequestParams } from "@/utils/request";

export interface TaskStatus {
    id: number;
    running: boolean;
    media_id: number;
    type: number;
    resolution: {
        width: number,
        height: number,
        fps: number,
    },
    stage: number;
    stage_start: number;
    time_now: number;
    stage_progress: number;
}

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
