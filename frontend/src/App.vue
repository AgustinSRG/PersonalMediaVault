<template>
  <MainLayout></MainLayout>
</template>

<script lang="ts">
import { Options, Vue } from "vue-class-component";

// Player components
import MainLayout from "./components/layout/MainLayout.vue";

@Options({
  components: {
    MainLayout,
  },
  data: function () {
    return {};
  },
})
export default class App extends Vue {}
</script>

<style>
/* Add here any global / shared styles */

.bold {
  font-weight: bold;
}

.text-right {
  text-align: right;
}

.text-center {
  text-align: center;
}

.text-left {
  text-align: left;
}

.one-line {
  white-space: nowrap;
}

.btn {
  display: inline-block;

  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: white;
  background: transparent;
  outline: none;
  border-radius: 100vw;
  padding: 12px 24px;
  white-space: nowrap;
}

.btn-sm {
  font-size: 16px;
}

.btn i {
  margin-right: 12px;
}

.btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.btn:hover {
  opacity: 1;
}

.btn:disabled:hover {
  opacity: 0.7;
}

.btn:focus {
  outline: none;
}

.btn-primary {
  border: solid 1px rgba(255, 255, 255, 0.1);
}

.btn-primary:not(disabled):hover {
  background: rgba(255, 255, 255, 0.1);
}

.form-group {
  padding-bottom: 0.75rem;
}

.form-group label {
  display: block;
  width: 100%;
  padding-bottom: 0.3rem;
}

.form-control {
  height: calc(1.5em + 0.75rem + 2px);
  padding: 0.375rem 0.75rem;
  font-size: 1rem;
  font-weight: 400;
  line-height: 1.5;
  background-clip: padding-box;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.25rem;
  transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;

  color: #fff;
  background: hsl(0, 0%, 7%);
}

.form-control:focus {
  color: #fff;
  background: hsl(0, 0%, 7%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  outline: 0;
  box-shadow: 0 0 0 0.2rem rgba(255, 255, 255, 0.1);
}

.form-control-full-width {
  width: 100%;
}

.form-error {
  color: red;
}
</style>
