<template>
  <div>
    <h2>ユーザ情報の編集</h2>
    <label for="username">ユーザ名</label>
    <input id="username" v-model="username" type="text" value="username" />
    <br />
    <label for="password">パスワード</label>
    <input v-model="password" type="password" />
    <br />
    <button v-on:click="editUser">更新</button>
    <div v-if="loading">ユーザ情報更新中…</div>
    <hr>
    <h2>アイコンの登録</h2>
    <input @change="selectedFile" type="file" name="file" />
    <button @click="upload" type="submit">アイコン(png)をアップロード</button>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "EditUser",
  props: {},
  components: {},
  data() {
    return {
      username: this.$route.params.username,
      password: "",
      loading: false,
      uploadfile: null
    };
  },
  methods: {
    selectedFile(e) {
      e.preventDefault();
      let files = e.target.files;
      this.uploadFile = files[0];
    },
    async upload() {
      const params = new FormData();
      params.append('iconimage', this.uploadFile);
      let config = {
        headers: {
          "content-type": "multipart/form-data",
        }
      };
      try {
        const res = await axios.post(`${this.$apiUrl}/users/${this.$route.params.username}/icon`, params, config)
        this.$router.push(`/users/${this.$route.params.username}`);
      } catch(e) {
        console.error(e)
      }
    },
    async editUser() {
      this.loading = true;

      let patchBody = {};
      if (this.username) {
        patchBody.username = this.username;
      }
      if (this.password) {
        patchBody.password = this.password;
      }
      axios
        .patch(
          `${this.$apiUrl}/users/${this.$route.params.username}`,
          patchBody
        )
        .then(res => {
          this.loading = false;
          this.$router.push(`/users/${this.$route.params.username}`);
        })
        .catch(e => {
          this.loading = false;
          console.error(e);
        });
    }
  },
  async created() {
    // check user is logged in or not
    if (localStorage.loggedIn) {
      this.loggedIn = true;
    } else {
      this.loggedIn = false;
    }
  }
};
</script>

<style scoped>
</style>
