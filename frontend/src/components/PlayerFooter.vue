<template>
  <footer class="fixed bottom-0 left-0 w-full z-10">
    <div class="bg-[#111111]/70 backdrop-blur-sm px-4 py-3 shadow-inner">
      <div class="max-w-5xl mx-auto flex gap-3 items-center">
        <img
          v-if="song"
          :src="song.cover || '/cover.jpg'"
          class="w-14 h-14 rounded object-cover"
        />

        <div class="flex-1 min-w-0">
          <div class="flex justify-between items-center">
            <div class="min-w-0">
              <div class="font-semibold truncate">
                {{ song?.title || "No track" }}
              </div>
              <div class="text-xs text-gray-400 truncate">
                {{ song?.artist || "" }}
              </div>
            </div>
            <div class="text-xs text-gray-400 time">
              {{ currentTime }} / {{ duration }}
            </div>
          </div>

          <div class="mt-3" @click="onSeekClick">
            <div class="progress-bar">
              <div class="progress" :style="{ width: progress + '%' }"></div>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button class="btn" @click="$emit('prev')">⏮</button>
          <button class="btn" @click="$emit('togglePlay')">
            {{ isPlaying ? "⏸" : "▶️" }}
          </button>
          <button class="btn" @click="$emit('next')">⏭</button>

          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            v-model.number="volumeLocal"
            class="w-24"
          />
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup>
import { ref, watch } from "vue";
const props = defineProps({
  song: Object,
  progress: Number,
  currentTime: String,
  duration: String,
  isPlaying: Boolean,
  volume: Number,
});
const emit = defineEmits([
  "seek",
  "togglePlay",
  "prev",
  "next",
  "update:volume",
]);

const volumeLocal = ref(props.volume ?? 0.9);

watch(volumeLocal, (v) => emit("update:volume", v));
watch(
  () => props.volume,
  (v) => (volumeLocal.value = v),
);

function onSeekClick(e) {
  const bar = e.currentTarget.querySelector(".progress-bar");
  if (!bar) return;
  const rect = bar.getBoundingClientRect();
  const pct = (e.clientX - rect.left) / rect.width;
  emit("seek", Math.max(0, Math.min(1, pct)));
}
</script>
