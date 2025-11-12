<template>
  <header class="fixed top-0 left-0 w-full z-10">
    <div class="bg-[#111111]/70 backdrop-blur-sm px-4 py-3 shadow">
      <div class="max-w-5xl mx-auto flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 bg-gradient-to-br from-indigo-600 to-cyan-400 rounded flex items-center justify-center text-white font-bold"
          >
            A
          </div>
          <div>
            <div class="font-semibold">TMPlayer</div>
            <div class="text-xs text-gray-400">tracks/ — folder</div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button class="btn" @click="refreshTracks">🔄 Refresh</button>
          <button
            class="btn"
            v-if="!allLoaded && !loading"
            @click="$emit('loadMore')"
          >
            Load more
          </button>
          <button class="btn" @click="router.push('/upload')">📂 Upload</button>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useRouter } from "vue-router";
import { defineProps, defineEmits } from "vue";
import { refreshTracksApi } from "../services/api.js";

const router = useRouter();
const props = defineProps({
  loading: { type: Boolean, default: false },
  allLoaded: { type: Boolean, default: false },
});
const emit = defineEmits(["refresh", "loadMore"]);

const refreshTracks = async () => {
  const result = await refreshTracksApi();
  if (result?.status === "ok") {
    emit("refresh");
  } else {
    console.error("refreshTracks failed");
  }
};

const onRefresh = () => emit("refresh");
</script>
