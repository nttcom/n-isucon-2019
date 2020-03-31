import Vue from 'vue'
import Router from 'vue-router'
import Items from './components/Items.vue'
import Item from './components/Item.vue'
import User from './components/User.vue'
import EditUser from './components/EditUser.vue'
import Registration from '@/components/Registration'
import Login from '@/components/Login'
import NewItem from '@/components/NewItem'
import EditItem from '@/components/EditItem'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Items',
      component: Items
    },
    {
      path: '/items/:id',
      name: 'Item',
      component: Item
    },
    {
      path: '/users/:username',
      name: 'user',
      component: User
    },
    {
      path: '/users/:username/edit',
      component: EditUser
    },
    {
      path: '/items/:id/edit',
      component: EditItem
    },
    {
      path: '/registration',
      component: Registration
    },
    {
      path: '/login',
      component: Login 
    },
    {
      path: '/draft/new',
      component: NewItem
    },
  ]
})