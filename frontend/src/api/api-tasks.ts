// Tasks API

import { GetApiURL, RequestParams } from "@/utils/request";

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
