<template>
  <div>
    <ul id="commments">
      <li style="margin-bottom: 10px;" v-for="(comment, key, index) in comments" :key="index">
        コメント番号: {{ comment.comment_id }}
        <br />
        {{ comment.comment }}
        <br />
        by {{comment.username}}
        <button v-if="isMycomment(comment.username)" v-on:click="removeComment(comment.comment_id)">削除</button>
      </li>
    </ul>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "Comments",
  props: {},
  data() {
    return {
      comments: []
    };
  },
  computed: {

  },
  methods: {
    isMycomment(username){
      if(localStorage.username && localStorage.username == username){
        return true;
      }
      return false;
    },
    async removeComment(comment_id) {
      try {
        const resRemovedComment = await axios.delete(
          `${this.$apiUrl}/items/${this.$route.params.id}/comments/${comment_id}`
        );

        // If DELETE gets 204, let's reload comments
        const resComments = await axios.get(
          `${this.$apiUrl}/items/${this.$route.params.id}/comments`
        );
        this.comments = resComments.data.comments;
      } catch (e) {
        console.error(e);
      }
    }
  },
  async mounted() {
    try {
      const resComments = await axios.get(
        `${this.$apiUrl}/items/${this.$route.params.id}/comments`
      );
      this.comments = resComments.data.comments;
    } catch (e) {
      if (e.response.status == 404) {
        this.comments = [];
      } else {
        console.error(e);
      }
    }
  }
};
</script>

<style scoped>
</style>
