<template>
  <div>
    <h2>ユーザ情報</h2>
    <img class="iconuserpage" v-bind:src="`${this.$apiUrl}/users/${username}/icon`"/>
    <p>ユーザ名: {{ username }}</p>
    <p>登録日: {{ created_at }}</p>
    <p>最終更新日: {{ updated_at }}</p>
    <button v-on:click="editUser">編集</button>
    <button v-on:click="removeUser">退会</button>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "User",
  props: {},
  components: {},
  data() {
    return {
      username: this.$route.params.username,
      password: "",
      created_at: "",
      updated_at: ""
    };
  },
  methods: {
    async editUser() {
      this.$router.push(`/users/${this.$route.params.username}/edit`);
    },
    async removeUser() {
      try {
        const res = await axios.delete(
          `${this.$apiUrl}/users/${this.$route.params.username}`
        );

        localStorage.removeItem("loggedIn");
        localStorage.removeItem("username");

        this.loggedIn = false;
        this.$router.push(`/`);
      } catch (e) {
        console.error(e);
      }
    }
  },
  async created() {
    // check user is logged in or not
    if (localStorage.loggedIn) {
      this.loggedIn = true;
    } else {
      this.loggedIn = false;
    }

    // fetch detail of user
    try {
      const resUser = await axios.get(
        `${this.$apiUrl}/users/${this.$route.params.username}`
      );
      this.created_at = resUser.data.created_at;
      this.updated_at = resUser.data.updated_at;
    } catch (e) {
      console.error(e);
    }
  }
};
</script>

<style scoped>
.iconuserpage {
  width: 100px;
  height: 100px;
}
</style>
