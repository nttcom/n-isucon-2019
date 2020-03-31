<template>
  <div>
    <h2>{{ title }}</h2>
    <div>最終更新日時: {{ updated_at }}</div>
    <span><img v-if="username" class="iconuseritem" v-bind:src="`${$apiUrl}/users/${username}/icon`"/>{{ username }}</span>
    <p>{{ body }}</p>
    <button v-if="isMyItem" v-on:click="editItem">編集</button>
    <button v-if="isMyItem" v-on:click="removeItem">削除</button>
    <hr />
    <likes></likes>
    <hr />
    <comments></comments>
    <new-comment-form v-if="loggedIn"></new-comment-form>
  </div>
</template>

<script>
import axios from "axios";
import Comments from "@/components/Comments";
import NewComment from "@/components/NewComment";
import Likes from "@/components/Likes";

export default {
  name: "Item",
  props: {},
  components: {
    comments: Comments,
    "new-comment-form": NewComment,
    likes: Likes
  },
  computed: {
    isMyItem(){
      if(localStorage.username && localStorage.username == this.username){
        return true;
      }
      return false;
    },
  },
  data() {
    return {
      title: "",
      body: "",
      username: "",
      created_at: "",
      updated_at: "",
      loggedIn: false
    };
  },
  methods: {
    async removeItem() {
      try {
        const resRemovedItem = await axios.delete(
          `${this.$apiUrl}/items/${this.$route.params.id}`
        );
        this.$router.push('/');
      } catch (e) {
        console.error(e);
      }
    },
    async editItem() {
      // todo: 自分の記事であれば、localstorageと判別する
      this.$router.push(`/items/${this.$route.params.id}/edit`);
    },
  },
  async created() {
    // check user is logged in or not
    if (localStorage.loggedIn) {
      this.loggedIn = true;
    } else {
      this.loggedIn = false;
    }

    // fetch detail of item
    try {
      const resItem = await axios.get(
        `${this.$apiUrl}/items/${this.$route.params.id}`
      );
      this.title = resItem.data.title;
      this.body = resItem.data.body;
      this.username = resItem.data.username;
      this.created_at = resItem.data.created_at;
      this.updated_at = resItem.data.updated_at;
    } catch (e) {
      console.error(e);
    }
  }
};
</script>

<style scoped>
.iconuseritem {
  width: 16px;
  height: 16px;
}
</style>
