<template>
  <div class="login">
    <h1>ログイン</h1>
    <label for="username">ユーザ名</label>
    <input id="username" v-model="eventData.username" type="text" />
    <br />
    <label for="password">パスワード</label>
    <input v-model="eventData.password" type="password" />
    <br />
    <button v-on:click="login">ログイン</button>
    <div v-if="loading">ログイン処理中…</div>
    <div v-if="failed">ユーザ名かパスワードが異なっています</div>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      eventData: {
        name: "",
        password: ""
      },
      loading: false,
      failed: false,
    };
  },
  methods: {
    async login() {
      axios.defaults.withCredentials = true;
      try {
        this.loading = true;
        const res = await axios.post(`${this.$apiUrl}/signin`, {
          username: this.eventData.username,
          password: this.eventData.password
        });
        this.failed = false;
        localStorage.loggedIn = true;
        localStorage.username = res.data.username;
        this.$router.push("/");
      } catch (e) {
        this.failed = true;
        console.error(e);
      } finally {
        this.loading = false;
      }
    },
  }
};
</script>