<template>
  <div class="min-h-screen flex flex-col">
    <Header :loading="loading" :allLoaded="allLoaded" @refresh="onRefresh" @loadMore="loadMore" />
    <main class="flex-1 overflow-auto p-4 space-y-4">
      <Visualizer
        :analyser="analyserNode"
        v-model:enabled="vizEnabled"
        v-model:volume="volume"
        v-model:fftSize="fftSize"
        v-model:smoothing="smoothing"
      />

      <SongList
        :songs="songs"
        :currentIndex="currentIndex"
        :loading="loading"
        :allLoaded="allLoaded"
        @play="onPlayFromList"
      />
    </main>

    <PlayerFooter
      :song="currentSong"
      :progress="progress"
      :currentTime="currentTimeDisplay"
      :duration="durationDisplay"
      :isPlaying="isPlaying"
      @togglePlay="togglePlay"
      @prev="prevSong"
      @next="nextSong"
      @seek="onSeek"
    />
  </div>
</template>

<script setup>
import { onMounted, ref, computed, watch } from 'vue';
import Header from './components/Header.vue';
import SongList from './components/SongList.vue';
import Visualizer from './components/Visualizer.vue';
import PlayerFooter from './components/PlayerFooter.vue';
import usePlayer from './composables/usePlayer.js';
import { fetchSongsApi } from './services/api.js';

// init composable with API function
const player = usePlayer(fetchSongsApi);

// expose local refs
const songs = player.songs;
const loading = player.loading;
const allLoaded = player.allLoaded;
const currentIndex = player.currentIndex;
const currentSong = player.currentSong;
const isPlaying = player.isPlaying;
const progress = player.progress;
const currentTimeDisplay = player.currentTimeDisplay;
const durationDisplay = player.durationDisplay;

// viz props
const vizEnabled = player.vizEnabled;
const volume = player.volume;
const fftSize = player.fftSize;
const smoothing = player.smoothing;

// analyser node for Visualizer component
const analyserNode = player.analyser;

// actions
const loadMore = player.loadMore;
const refreshSongs = player.refreshSongs;
const playSong = player.playSong;
const togglePlay = player.togglePlay;
const nextSong = player.nextSong;
const prevSong = player.prevSong;
const seek = player.seek;

// wiring UI events
const onPlayFromList = async (index) => { await playSong(index); };
const onRefresh = async () => { await refreshSongs(); };

const onSeek = (pct) => seek(pct);

// load initial chunk
onMounted(async () => {
  await refreshSongs();
});

const updateFavicon = (icon) => {
  let link = document.querySelector("link[rel~='icon']");
  if (!link) {
    link = document.createElement('link');
    link.rel = 'icon';
    document.head.appendChild(link);
  }
  link.href = icon;
};

watch(
  () => ({
    song: currentSong.value,
    playing: isPlaying.value
  }),
  ({ song, playing }) => {
    const icon = !song ? '/music.svg' : playing ? '/play.svg' : '/pause.svg';
    let link = document.querySelector("link[rel~='icon']");
    if (!link) {
      link = document.createElement('link');
      link.rel = 'icon';
      document.head.appendChild(link);
    }
    link.href = icon;
  },
  { immediate: true, deep: false }
);
</script>
