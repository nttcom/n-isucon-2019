<template>
  <div class="registration">
    <h1>ユーザ登録</h1>
    <label for="username">ユーザ名</label>
    <input id="username" v-model="eventData.username" type="text">
    <br>
    <label for="password">パスワード</label>
    <input v-model="eventData.password" type="password">
    <br>
    <button v-on:click="register">登録</button>
    <div v-if="loading">ユーザ登録中…</div>
  </div>
</template>

<script>
import axios from "axios";

export default {
    data () {
        return {
            eventData: {
                username: "sample",
                password: "pass"
            },
            loading: false,
        }
    },
    methods: {
        register: function(e) {
            this.loading = true;
            axios.post(`${this.$apiUrl}/users`, {
                username: this.eventData.username,
                password: this.eventData.password,
            }).then(res => {
                this.loading = false;
                this.$router.push('/login')
            }).catch(e => {
                this.loading = false;
                console.error(e);
            });
        }
    }
}
</script>