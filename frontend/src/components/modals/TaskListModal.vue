<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <div
      class="modal-dialog modal-xl"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Tasks") }}</div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="loading" class="modal-body">
        <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
      </div>
      <div v-if="!loading" class="">
        <div class="table-responsive">
          <table class="table">
            <thead>
              <tr>
                <th class="text-left" colspan="2">
                  {{ $t("List of active tasks") }}
                </th>
                <th class="text-left">{{ $t("Type") }}</th>
                <th class="text-left">{{ $t("Media") }}</th>
                <th class="text-left">{{ $t("Status") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="tasks.length === 0">
                <td colspan="5">
                  {{
                    $t("There are no active tasks in the vault at the moment.")
                  }}
                </td>
              </tr>
              <tr v-for="t in tasks" :key="t.id">
                <td class="one-line td-shrink">
                  <i
                    class="fas fa-circle task-status"
                    :class="{ 'task-running': t.running }"
                  ></i>
                </td>
                <td class="task-pbar-td one-line td-shrink">
                  <div class="task-pbar-container" v-if="t.running">
                    <div
                      class="task-pbar-current"
                      :style="{ width: getProgressPercent(t.stage_progress) }"
                    ></div>
                  </div>
                </td>
                <td class="bold one-line td-shrink">{{ renderType(t.type) }}</td>
                <td class="bold one-line td-shrink">
                  <a
                    @click="goToMedia(t.media_id, $event)"
                    :href="getMediaURL(t.media_id)"
                    target="_blank"
                    rel="noopener noreferrer"
                    >{{ t.media_id }}</a
                  >
                </td>
                <td class="one-line">
                  {{
                    renderStatus(
                      t.running,
                      t.stage,
                      t.stage_progress,
                      t.stage_start,
                      t.time_now
                    )
                  }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { TasksAPI } from "@/api/api-tasks";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { GenerateURIQuery, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time-utils";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "TaskListModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      tasks: [],

      loading: true,
    };
  },
  methods: {
    setTasks: function (tasks) {
      this.tasks = tasks.sort((a, b) => {
        if (a.running && !b.running) {
          return -1;
        } else if (!a.running && b.running) {
          return 1;
        } else if (a.type < b.type) {
          return -1;
        } else if (a.type > b.type) {
          return 1;
        } else if (a.id < b.id) {
          return -1;
        } else {
          return 1;
        }
      });
    },

    load: function () {
      Timeouts.Abort("admin-tasks");
      Request.Abort("admin-tasks");
      Timeouts.Abort("admin-tasks-update");
      Timeouts.Abort("admin-tasks-update");

      if (!this.display) {
        return;
      }

      this.loading = true;

      Request.Pending("admin-tasks", TasksAPI.GetTasks())
        .onSuccess((tasks) => {
          this.setTasks(tasks);
          this.loading = false;
          Timeouts.Set("admin-tasks-update", 500, this.updateTasks.bind(this));
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.displayStatus = false;
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("admin-tasks", 1500, this.load.bind(this));
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("admin-tasks", 1500, this.load.bind(this));
        });
    },

    updateTasks: function () {
      Timeouts.Abort("admin-tasks-update");
      Timeouts.Abort("admin-tasks-update");

      if (!this.display) {
        return;
      }

      Request.Pending("admin-tasks-update", TasksAPI.GetTasks())
        .onSuccess((tasks) => {
          this.setTasks(tasks);
          Timeouts.Set("admin-tasks-update", 500, this.updateTasks.bind(this));
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.displayStatus = false;
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set(
                "admin-tasks-update",
                1500,
                this.updateTasks.bind(this)
              );
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("admin-tasks-update", 1500, this.updateTasks.bind(this));
        });
    },

    getProgressPercent: function (p: number) {
      return Math.floor(p * 100) / 100 + "%";
    },

    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    renderType: function (t: number) {
      switch (t) {
        case 0:
          return this.$t("Encode media");
        case 1:
          return this.$t("Encode extra resolution");
        case 2:
          return this.$t("Generate video previews");
        default:
          return "???";
      }
    },

    renderStatus: function (
      running: boolean,
      stage: string,
      p: number,
      ts: number,
      now: number
    ) {
      if (running) {
        let stageNumber = 0;

        switch (stage) {
          case "PREPARE":
            stageNumber = 0;
            break;
          case "COPY":
            stageNumber = 1;
            break;
          case "PROBE":
            stageNumber = 2;
            break;
          case "ENCODE":
            stageNumber = 3;
            break;
          case "ENCRYPT":
            stageNumber = 4;
            break;
          case "UPDATE":
            stageNumber = 5;
            break;
          case "FINISH":
            stageNumber = 6;
            break;
        }

        const progressPercent = this.getProgressPercent(p);

        const estimatedReaminingTime =
          (((now - ts) / p) * 100 - (now - ts)) / 1000;

        let txt =
          this.$t("Stage") +
          ": " +
          stageNumber +
          " / 6 | " +
          this.$t("Stage progress") +
          ": " +
          progressPercent;

        if (estimatedReaminingTime > 0) {
          txt +=
            " | " +
            this.$t("Remaining time (estimated)") +
            ": " +
            renderTimeSeconds(estimatedReaminingTime);
        }

        return txt;
      } else {
        return this.$t("Task is in queue");
      }
    },

    goToMedia: function (mid, e) {
      if (e) {
        e.preventDefault();
      }
      AppStatus.ClickOnMedia(mid, true);
      this.close();
    },

    getMediaURL: function (mid: number): string {
      return (
        window.location.protocol +
        "//" +
        window.location.host +
        window.location.pathname +
        GenerateURIQuery({
          media: mid + "",
        })
      );
    },
  },
  mounted: function () {
    this.load();
  },
  beforeUnmount: function () {
    Timeouts.Abort("admin-tasks");
    Request.Abort("admin-tasks");
    Timeouts.Abort("admin-tasks-update");
    Timeouts.Abort("admin-tasks-update");
  },
  watch: {
    display: function () {
      this.load();
    },
  },
});
</script>

<style>
.task-status {
  color: gray;
}

.task-status.task-running {
  color: lime;
}

.task-pbar-container {
  width: 120px;
  height: 1rem;
  display: flex;
  border: solid 1px var(--theme-border-color);
}

.task-pbar-current {
  height: 100%;
  background: lime;
}
</style>