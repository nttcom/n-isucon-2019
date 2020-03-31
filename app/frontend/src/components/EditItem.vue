<template>
  <div class="new-item">
    <h1>記事を編集</h1>
    <label for="title">タイトル</label>
    <input id="title" v-model="eventData.title" type="text" />
    <br />
    <label for="body">本文</label>
    <textarea v-model="eventData.body" name id cols="30" rows="10"></textarea>
    <br />
    <button v-on:click="register">更新</button>
    <div v-if="loading">記事を更新中…</div>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      eventData: {
        title: "",
        body: ""
      },
      loading: false
    };
  },
  methods: {
    register: function(e) {
      // Patch
      this.loading = true;
      axios.patch(`${this.$apiUrl}/items/${this.$route.params.id}`, {
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
  },
  async created() {
    try {
      const resItem = await axios.get(
        `${this.$apiUrl}/items/${this.$route.params.id}`
      );
      this.eventData.title = resItem.data.title;
      this.eventData.body = resItem.data.body;
    } catch (e) {
      console.error(e);
    }
  }
};
</script>