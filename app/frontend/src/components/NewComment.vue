<template>
  <div class="new-comment">
    <h3>コメントを投稿する</h3>
    <label for="body">本文</label>
    <br>
    <textarea v-model="eventData.body" name="" id="" cols="30" rows="10"></textarea>
    <br>
    <button v-on:click="register">投稿</button>
    <div v-if="loading">コメントを投稿中…</div>
  </div>
</template>

<script>
import axios from "axios";

export default {
    data () {
        return {
            eventData: {
                body: "なんて良いコメント"
            },
            loading: false,
        }
    },
    methods: {
        register: function(e) {
            this.loading = true;
            axios.post(`${this.$apiUrl}/items/${this.$route.params.id}/comments`, {
                comment: this.eventData.body,
            }).then(res => {
                this.loading = false;
                location.reload();
            }).catch(e => {
                this.loading = false;
                console.error(e);
            });
        }
    }
}
</script>