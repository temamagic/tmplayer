<template>
  <li
    @click="$emit('play')"
    :class="[
      'song flex items-center gap-4 px-3 py-2 cursor-pointer',
      active ? 'active' : '',
    ]"
  >
    <img
      :src="song.cover || '/cover.jpg'"
      alt="cover"
      class="w-12 h-12 rounded object-cover"
    />
    <div class="flex-1 min-w-0">
      <div style="font-weight: 500" class="font-medium truncate">
        {{ song.title }}
      </div>
      <div class="text-xs text-gray-400 truncate">
        {{ song.artist || "Unknown artist" }}
      </div>
    </div>
    <div class="text-sm text-gray-400">{{ song.id }}</div>
  </li>
</template>

<script setup>
const props = defineProps({
  song: { type: Object, required: true },
  active: { type: Boolean, default: false },
});
</script>
