<template>
  <div>
    <p>&#x1f44d;総数: {{ like_count }}</p>
    <span>&#x1f44d;したユーザ一覧</span>
    <ul>
      <li id="likes" v-for="(like, key, index) in likes" :key="index">{{like}}</li>
    </ul>
    <button v-on:click="postLike"> &#x1f44d; </button>
    <button v-on:click="removeLike"> &#x1f44d; を取り消す</button>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "Likes",
  props: {},
  data() {
    return {
      likes: [],
      like_count: 0,
    };
  },
  methods: {
    async fetchLikes() {
      try {
        const resLikes = await axios.get(
          `${this.$apiUrl}/items/${this.$route.params.id}/likes`
        );
        // Fetched data should be `"likes": "4,6"`, or `"likes": ""` if there's no likes.
        if (resLikes.data.likes === "") {
          this.likes = [];
          this.like_count = 0;
        } else {
          this.likes = resLikes.data.likes.split(",");
          this.like_count = resLikes.data.like_count;
        }
      } catch (e) {
        console.error(e);
      }
    },
    async postLike() {
      try {
        const resLike = await axios.post(
          `${this.$apiUrl}/items/${this.$route.params.id}/likes`
        );
        await this.fetchLikes();
      } catch (e) {
        console.error(e);
      }
    },
    async removeLike() {
      try {
        console.log("いいねを取り消すよー");
        const resLike = await axios.delete(
          `${this.$apiUrl}/items/${this.$route.params.id}/likes`
        );
        await this.fetchLikes();
      } catch (e) {
        console.error(e);
      }
    }
  },
  async mounted() {
    try {
      await this.fetchLikes();
    } catch (e) {
      console.error(e);
    }
  }
};
</script>

<style scoped>
#likes {
  list-style: square;
}
</style>
