<template>
  <header class="header">
    <div class="brand">
      <div class="logo">A</div>
      <div>
        <div style="font-weight:600">TMPlayer</div>
        <div class="small">tracks/ — folder</div>
      </div>
    </div>

    <div style="display:flex; gap:8px; align-items:center;">
      <button class="btn" @click="onRefresh">🔄 Refresh</button>
      <button class="btn" v-if="!allLoaded && !loading" @click="$emit('loadMore')">Load more</button>
    </div>
  </header>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue';
const props = defineProps({
  loading: { type: Boolean, default: false },
  allLoaded: { type: Boolean, default: false }
});
const emit = defineEmits(['refresh','loadMore']);

const onRefresh = () => emit('refresh');
</script>
