import { createRouter, createWebHistory } from "vue-router";
import Home from "../views/Home.vue";
import UploadTrack from "../views/UploadTrack.vue";

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home,
  },
  {
    path: "/upload",
    name: "UploadTrack",
    component: UploadTrack,
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
