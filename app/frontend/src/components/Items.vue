<template>
  <div>
    <button :disabled="sortedByUpdate" v-on:click="fetchItemsSortedByCreated">新着順</button>
    <button :disabled="!sortedByUpdate" v-on:click="fetchItemsSortedByLikes">いいね順</button>
    <p v-if="loading">データ取得中…</p>
    <ul id="items">
      <!-- todo: add icon of user -->
      <li v-for="(item, key, index) in items" :key="index">
        <router-link v-bind:to="'items/' + item.id">{{ item.title }}</router-link>
        <br />
        <img class="icon" v-bind:src="`${$apiUrl}/users/${item.username}/icon`"/>
        by {{item.username}} / {{item.created_at}} / &#x1f44d; {{ item.likes }}
      </li>
    </ul>
    <ul style="display: inline;">
      <button v-if="currentPage > 1" v-on:click="movePage(currentPage - 1)">前</button>
      <button v-if="currentPage > 2" v-on:click="movePage(currentPage - 1)">{{ currentPage - 1 }}</button>
      <button v-bind:class='{active:currentPage === currentPage}' v-on:click="movePage(currentPage)">{{ currentPage }}</button>
      <button v-if="currentPage < totalPages - 1" v-on:click="movePage(currentPage + 1)">{{ currentPage + 1 }}</button>
      <button v-if="currentPage < totalPages" v-on:click="movePage(currentPage + 1)">次</button>
    </ul>
  </div>
</template>

<script>
import axios from "axios";
axios.defaults.withCredentials = true;

export default {
  name: "Items",
  props: {},
  data() {
    return {
      items: [],
      currentPage: 1,
      itemPerPage: 10,
      totalPages: 0,
      totalItems: 0,
      sortedByUpdate: true,
      loading: false,
    };
  },
  methods: {
    async movePage(page){
      this.currentPage = page;
      if (this.sortedByUpdate) {
        this.fetchItemsSortedByCreated();
      } else {
        this.fetchItemsSortedByLikes();
      }
    },
    async fetchItemsSortedByLikes() {
      this.sortedByUpdate = false;
      try {
        this.loading = true;
        const resItems = await axios.get(
          `${this.$apiUrl}/items?sort=like&page=${this.currentPage - 1}`
        );
        this.totalItems = resItems.data.count;
        this.loading = false;

        const allItems = await Promise.all(
          resItems.data.items.map(async item => {
            return axios.get(`${this.$apiUrl}/items/${item.id}`);
          })
        );

        this.items = [];
        allItems.forEach(item => {
          item.data.likes = this.countLikes(item.data.likes);
          this.items.push(item.data);
        });
      } catch (e) {}
    },
    async fetchItemsSortedByCreated() {
      this.sortedByUpdate = true;
      try {
        this.loading = true;
        const resItems = await axios.get(
          `${this.$apiUrl}/items?page=${this.currentPage - 1}`
        );
        this.loading = false;
        this.totalItems = resItems.data.count;
        this.totalPages = Math.round(this.totalItems / this.itemPerPage);

        const allItems = await Promise.all(
          resItems.data.items.map(async item => {
            return axios.get(`${this.$apiUrl}/items/${item.id}`);
          })
        );

        this.items = [];
        allItems.forEach(item => {
          item.data.likes = this.countLikes(item.data.likes);
          this.items.push(item.data);
        });
      } catch (e) {}
    },
    countLikes(rawLikesString) {
      if (rawLikesString === "") {
        return 0;
      }
      return rawLikesString.split(",").length;
    }
  },
  async created() {
    this.fetchItemsSortedByCreated();
  }
};
</script>

<style scoped>
.pagenate {
  display: inline;
}
.active {
  background-color: darkblue;
  color: white;
}
.icon {
  width: 12px;
  height: 12px;
}
</style>
