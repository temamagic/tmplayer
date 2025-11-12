<template>
  <section>
    <ul class="bg-[#0b0b0b] rounded p-2 divide-y divide-gray-900">
      <SongItem
        v-for="(song, i) in songs"
        :key="song.id"
        :song="song"
        :active="currentIndex === i"
        @play="$emit('play', i)"
      />
    </ul>

    <div class="text-center text-gray-400 py-3" v-if="loading">
      ⏳ Loading...
    </div>
    <div
      class="text-center text-gray-400 py-3"
      v-if="allLoaded && songs.length === 0"
    >
      No tracks here.
    </div>
  </section>
</template>

<script setup>
import SongItem from "./SongItem.vue";
const props = defineProps({
  songs: Array,
  currentIndex: Number,
  loading: Boolean,
  allLoaded: Boolean,
});
</script>
