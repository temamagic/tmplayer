<template>
  <footer
    class="bg-[#111111] px-4 py-3 shadow-inner"
    style="display: flex; gap: 12px; align-items: center"
  >
    <img
      v-if="song"
      :src="song.cover || '/cover.jpg'"
      class="w-14 h-14 rounded object-cover"
    />

    <div style="flex: 1; min-width: 0">
      <div
        style="
          display: flex;
          justify-content: space-between;
          align-items: center;
        "
      >
        <div style="min-width: 0">
          <div style="font-weight: 600" class="truncate">
            {{ song?.title || "No track" }}
          </div>
          <div class="small truncate">{{ song?.artist || "" }}</div>
        </div>
        <div class="small">{{ currentTime }} / {{ duration }}</div>
      </div>

      <div style="margin-top: 12px" @click="onSeekClick">
        <div class="progress-bar">
          <div class="progress" :style="{ width: progress + '%' }"></div>
        </div>
      </div>
    </div>

    <div style="display: flex; align-items: center; gap: 8px">
      <button class="btn" @click="$emit('prev')">⏮</button>
      <button class="btn" @click="$emit('togglePlay')">
        {{ isPlaying ? "⏸" : "▶️" }}
      </button>
      <button class="btn" @click="$emit('next')">⏭</button>
    </div>

    <!-- audio is managed in usePlayer; we don't include <audio> here -->
  </footer>
</template>

<script setup>
const props = defineProps({
  song: Object,
  progress: Number,
  currentTime: String,
  duration: String,
  isPlaying: Boolean,
});
const emit = defineEmits(["seek", "togglePlay", "prev", "next"]);

function onSeekClick(e) {
  const bar = e.currentTarget.querySelector(".progress-bar");
  if (!bar) return;
  const rect = bar.getBoundingClientRect();
  const pct = (e.clientX - rect.left) / rect.width;
  emit("seek", Math.max(0, Math.min(1, pct)));
}
</script>
