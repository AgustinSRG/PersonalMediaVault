<template>
  <div
    class="modal-container modal-container-login"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
  >
    <form @submit="submit" class="modal-dialog modal-md" role="document">
      <div class="modal-header">
        <div class="modal-title no-close">{{ $t("The media vault is locked") }}</div>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Username") }}:</label>
          <input
            type="text"
            name="username"
            v-model="username"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width auto-focus"
          />
        </div>
        <div class="form-group">
          <label>{{ $t("Password") }}:</label>
          <input
            type="password"
            name="password"
            v-model="password"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width"
          />
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer">
        <button v-if="!busy && mustWait <= 0" type="submit" class="modal-footer-btn">
          <i class="fas fa-unlock"></i> {{ $t("Unlock vault") }}
        </button>
        <button v-if="!busy && mustWait === 1" type="button" disabled class="modal-footer-btn">
          <i class="fas fa-hourglass"></i> {{ $t("You must wait 1 second to try again") }}
        </button>
        <button v-if="!busy && mustWait > 1" type="button" disabled class="modal-footer-btn">
          <i class="fas fa-hourglass"></i> {{ $t("You must wait $TIME seconds to try again").replace("$TIME", mustWait + "") }}
        </button>
        <button v-if="busy" type="button" disabled class="modal-footer-btn">
          <i class="fa fa-spinner fa-spin"></i> {{ $t("Unlocking vault") }}...
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AuthAPI } from "@/api/api-auth";
import { AuthController } from "@/control/auth";
import { Request } from "@/utils/request";
import { defineComponent } from "vue";

export default defineComponent({
  name: "LoginModal",
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      username: "",
      password: "",
      cooldown: 0,
      mustWait: 0,
      now: Date.now(),
      busy: false,
      error: "",
    };
  },
  methods: {
    autoFocus: function () {
      if (!this.display) {
        return;
      }
      const elem = this.$el.querySelector(".auto-focus");
      if (elem) {
        setTimeout(() => {
          elem.focus();
        }, 200);
      }
    },

    submit: function (e) {
      e.preventDefault();

      if (this.busy) {
        return;
      }

      this.busy = true;
      this.error = "";

      Request.Do(AuthAPI.Login(this.username, this.password))
        .onSuccess((response) => {
          this.busy = false;
          this.username = "";
          this.password = "";
          AuthController.SetSession(response.session_id, response.vault_fingerprint);
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              this.error = this.$t("Invalid username or password");
            })
            .add(403, "COOLDOWN", () => {
              this.error = this.$t("You must wait 5 seconds to try again");
            })
            .add(403, "*", () => {
              this.error = this.$t("Invalid username or password");
              this.cooldown = Date.now() + 5000;
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.busy = false;
        });
    },

    updateNow: function () {
      this.now = Date.now();
      if (this.now < this.cooldown) {
        this.mustWait  = Math.max(1, Math.round((this.cooldown - this.now) / 1000));
      } else {
        this.mustWait = 0;
      }
    },
  },
  mounted: function () {
    this.autoFocus();

    this.$options.timer = setInterval(this.updateNow.bind(this), 200);
  },
  watch: {
    display: function () {
      this.error = "";
      this.autoFocus();
    },
  },
  beforeUnmount: function () {
    clearInterval(this.$options.timer);
  },
});
</script>

<style>
.modal-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;

  background: rgba(0, 0, 0, 0.4);

  display: flex;
  flex-direction: column;

  padding: 1rem;

  transition: opacity 0.2s;
  opacity: 1;

  overflow: auto;
}

.modal-container-login {
  z-index: 301;
}

.modal-container.hidden {
  transition: opacity 0.2s, visibility 0.2s;
  pointer-events: none;
  opacity: 0;
  visibility: hidden;
}

.modal-dialog {
  display: flex;
  margin: auto;
  flex-direction: column;
  background: #212121;
  box-shadow: 0 16px 24px 2px rgb(0 0 0 / 14%), 0 6px 30px 5px rgb(0 0 0 / 12%),
    0 8px 10px -5px rgb(0 0 0 / 40%);
}

.modal-sm {
  width: 300px;
}

@media (max-width: 300px) {
  .modal-sm {
    width: calc(100% - 1rem);
  }
}

.modal-md {
  width: 500px;
}

@media (max-width: 600px) {
  .modal-md {
    width: calc(100% - 1rem);
  }
}

.modal-lg {
  width: 800px;
}

@media (max-width: 900px) {
  .modal-lg {
    width: calc(100% - 1rem);
  }
}

.modal-xl {
  width: 1140px;
}

@media (max-width: 1240px) {
  .modal-xl {
    width: calc(100% - 1rem);
  }
}

.modal-header {
  display: flex;
  flex-direction: row;
  align-items: center;
  border-bottom: solid 1px rgba(255, 255, 255, 0.1);
}

.modal-title {
  width: calc(100% - 48px);
  padding: 1rem;
  font-size: 24px;
  font-weight: bold;
}

.modal-title.no-close {
  width: 100%;
}

.modal-close-btn {
  display: block;
  width: 48px;
  height: 48px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
}

.modal-close-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.modal-close-btn:hover {
  color: white;
}

.modal-close-btn:disabled:hover {
  color: rgba(255, 255, 255, 0.75);
}

.modal-body {
  padding: 1rem;
}

.modal-footer {
  border-top: solid 1px rgba(255, 255, 255, 0.1);
}

.modal-footer-btn {
  display: block;
  width: 100%;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
  text-align: left;
  padding: 1rem;
  white-space: nowrap;
  font-weight: bold;
}

.modal-footer-btn i {
  width: 1.5rem;
}

.modal-footer-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.modal-footer-btn:hover {
  color: white;
}

.modal-footer-btn:disabled:hover {
  color: rgba(255, 255, 255, 0.75);
}
</style>
