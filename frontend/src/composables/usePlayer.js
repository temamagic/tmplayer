import { ref, reactive, computed, watch, nextTick } from 'vue';

/**
 * usePlayer(fetchSongsFn)
 * - encapsulates Audio element, AudioContext, AnalyserNode, GainNode
 * - exposes methods for play/pause/next/prev/seek, reactive state and analyser
 */
export default function usePlayer(fetchSongsFn) {
  // state
  const songs = ref([]);
  const offset = ref(0);
  const limit = ref(5);
  const allLoaded = ref(false);
  const loading = ref(false);

  const currentIndex = ref(-1);
  const currentSong = computed(() => (currentIndex.value >= 0 ? songs.value[currentIndex.value] : null));

  // audio element
  const audio = ref(new Audio());
  audio.value.preload = 'metadata';
  audio.value.crossOrigin = 'anonymous';

  const isPlaying = ref(false);
  const progress = ref(0);
  const currentTimeDisplay = ref('0:00');
  const durationDisplay = ref('0:00');

  // WebAudio
  const audioCtx = ref(null);
  const analyser = ref(null);
  const gainNode = ref(null);
  const sourceNode = ref(null);
  const analyserData = ref(null);

  // viz / settings
  const vizEnabled = ref(true);
  const volume = ref(0.9);
  const fftSize = ref(1024);
  const smoothing = ref(0.6);

  // helpers
  const formatTime = (s) => {
    if (!s && s !== 0) return '0:00';
    const m = Math.floor(s / 60);
    const sec = Math.floor(s % 60).toString().padStart(2, '0');
    return `${m}:${sec}`;
  };

  // init AudioContext and nodes (call after a user gesture)
  const initAudioGraph = () => {
    if (audioCtx.value) return;
    const AC = window.AudioContext || window.webkitAudioContext;
    audioCtx.value = new AC();
    analyser.value = audioCtx.value.createAnalyser();
    gainNode.value = audioCtx.value.createGain();

    analyser.value.fftSize = fftSize.value;
    analyser.value.smoothingTimeConstant = smoothing.value;
    gainNode.value.gain.value = volume.value;

    analyser.value.connect(gainNode.value);
    gainNode.value.connect(audioCtx.value.destination);

    analyserData.value = new Uint8Array(analyser.value.frequencyBinCount);
  };

  const connectSource = () => {
    if (!audioCtx.value) return;
    try {
      if (sourceNode.value) {
        try { sourceNode.value.disconnect(); } catch (e) {}
        sourceNode.value = null;
      }
      // createMediaElementSource can only be used once per element in some browsers;
      // but we recreate safely after disconnect.
      sourceNode.value = audioCtx.value.createMediaElementSource(audio.value);
      sourceNode.value.connect(analyser.value);
    } catch (e) {
      console.warn('connectSource failed', e);
    }
  };

  // Fade helpers (smooth fade in/out)
  const fadeTo = (target, time = 0.4) => {
    if (!gainNode.value || !audioCtx.value) return;
    const g = gainNode.value.gain;
    const now = audioCtx.value.currentTime;
    g.cancelScheduledValues(now);
    g.setValueAtTime(g.value, now);
    g.linearRampToValueAtTime(target, now + time);
  };

  const fadeIn = (to = volume.value, time = 0.4) => fadeTo(to, time);
  const fadeOut = (time = 0.45) => fadeTo(0, time);

  // playback
  const playSong = async (index) => {
    if (!songs.value[index]) return;
    currentIndex.value = index;
    const s = songs.value[index];

    // ensure absolute/correct URL — backend returns src like /tracks/...
    audio.value.src = s.src;
    try {
      if (!audioCtx.value) initAudioGraph();
      if (audioCtx.value.state === 'suspended') await audioCtx.value.resume();
      connectSource();
      gainNode.value.gain.value = 0; // start from 0 then fadeIn
      await audio.value.play();
      fadeIn();
      isPlaying.value = true;
    } catch (e) {
      console.warn('playSong error', e);
    }
  };

  const togglePlay = async () => {
    if (!audioCtx.value) initAudioGraph();
    if (!currentSong.value && songs.value.length > 0) {
      await playSong(0);
      return;
    }
    if (audio.value.paused) {
      // resume
      if (audioCtx.value.state === 'suspended') await audioCtx.value.resume();
      fadeIn();
      await audio.value.play();
      isPlaying.value = true;
    } else {
      // pause with fade
      fadeOut();
      setTimeout(() => {
        try { audio.value.pause(); } catch (e) {}
      }, 460);
      isPlaying.value = false;
    }
  };

  const nextSong = async () => {
    if (songs.value.length === 0) return;
    let next = currentIndex.value + 1;
    if (next >= songs.value.length) {
      // attempt to load more if not allLoaded
      if (!allLoaded.value) await loadMore();
      if (next >= songs.value.length) next = 0;
    }
    await playSong(next % songs.value.length);
  };

  const prevSong = async () => {
    if (songs.value.length === 0) return;
    let prev = currentIndex.value - 1;
    if (prev < 0) prev = songs.value.length - 1;
    await playSong(prev);
  };

  // progress / seek
  const updateProgress = () => {
    const a = audio.value;
    if (!a || !a.duration || isNaN(a.duration)) return;
    progress.value = (a.currentTime / a.duration) * 100;
    currentTimeDisplay.value = formatTime(a.currentTime);
    durationDisplay.value = formatTime(a.duration);
  };

  const seek = (pct) => {
    if (!audio.value.duration) return;
    audio.value.currentTime = Math.max(0, Math.min(1, pct)) * audio.value.duration;
    updateProgress();
  };

  // viz data fetching for Visualizer component
  const getAnalyserData = () => {
    if (!analyser.value || !analyserData.value) return null;
    analyser.value.getByteFrequencyData(analyserData.value);
    return analyserData.value;
  };

  // songs loading
  const refreshSongs = async () => {
    offset.value = 0;
    allLoaded.value = false;
    songs.value = [];
    await loadMore();
  };

  const loadMore = async () => {
    if (loading.value || allLoaded.value) return;
    loading.value = true;
    try {
      const data = await fetchSongsFn(offset.value, limit.value);
      if (!Array.isArray(data) || data.length === 0) {
        allLoaded.value = true;
      } else {
        songs.value.push(...data);
        offset.value += data.length;
      }
    } catch (e) {
      console.error('loadMore error', e);
    } finally {
      loading.value = false;
    }
  };

  // event wiring
  audio.value.addEventListener('timeupdate', updateProgress);
  audio.value.addEventListener('loadedmetadata', () => {
    durationDisplay.value = formatTime(audio.value.duration || 0);
  });
  audio.value.addEventListener('play', () => { isPlaying.value = true; });
  audio.value.addEventListener('pause', () => { isPlaying.value = false; });
  audio.value.addEventListener('ended', nextSong);

  // watchers for settings
  watch(volume, (v) => {
    if (gainNode.value) gainNode.value.gain.value = v;
    // also update default fade-in target
  });
  watch(fftSize, (v) => {
    if (analyser.value) {
      analyser.value.fftSize = v;
      analyserData.value = new Uint8Array(analyser.value.frequencyBinCount);
    }
  });
  watch(smoothing, (v) => {
    if (analyser.value) analyser.value.smoothingTimeConstant = v;
  });

  return {
    // state
    songs,
    allLoaded,
    loading,
    offset,
    limit,

    currentIndex,
    currentSong,
    isPlaying,
    progress,
    currentTimeDisplay,
    durationDisplay,

    // audio & viz
    audio,
    audioCtx,
    analyser,
    getAnalyserData,

    // settings
    vizEnabled,
    volume,
    fftSize,
    smoothing,

    // actions
    refreshSongs,
    loadMore,
    playSong,
    togglePlay,
    nextSong,
    prevSong,
    seek,

    // internal helpers (exposed if needed)
    initAudioGraph,
    connectSource,
  };
}
