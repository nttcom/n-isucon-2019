<template>
  <div class="new-item">
    <h1>新規記事を投稿</h1>
    <label for="title">タイトル</label>
    <input id="title" v-model="eventData.title" type="text">
    <br>
    <label for="body">本文</label>
    <textarea v-model="eventData.body" name="" id="" cols="30" rows="10"></textarea>
    <br>
    <button v-on:click="register">Niitaに投稿</button>
    <div v-if="loading">記事を投稿中…</div>
  </div>
</template>

<script>
import axios from "axios";

export default {
    data () {
        return {
            eventData: {
                title: "いい感じのタイトルをつけよう",
                body: "なんでも書いて共有しよう"
            },
            loading: false,
        }
    },
    methods: {
        register: function(e) {
            this.loading = true;
            axios.post(`${this.$apiUrl}/items`, {
                title: this.eventData.title,
                body: this.eventData.body,
            }).then(res => {
                this.loading = false;
                this.$router.push('/items/' + res.data.id);
            }).catch(e => {
                this.loading = false;
                console.error(e);
            });
        }
    }
}
</script>