<template>
  <section>
    <div class="flex items-center justify-between mb-2">
      <div class="flex items-center gap-3">
        <div class="font-medium">Track</div>
        <div class="text-xs text-gray-400">WebAudio API</div>
      </div>

      <div class="flex items-center gap-3">
        <label class="text-xs text-gray-400">Volume</label>
        <input
          type="range"
          min="0"
          max="1"
          step="0.01"
          v-model.number="volumeLocal"
          class="w-32"
        />
        <label class="text-xs text-gray-400 ml-2">FFT</label>
        <select
          v-model.number="fftLocal"
          class="bg-[#0f0f0f] px-2 py-1 rounded text-sm"
        >
          <option v-for="n in [256, 512, 1024, 2048]" :key="n" :value="n">
            {{ n }}
          </option>
        </select>
        <label class="text-xs text-gray-400 ml-2">Smoothing</label>
        <input
          type="range"
          min="0"
          max="0.99"
          step="0.01"
          v-model.number="smoothingLocal"
          class="w-28"
        />
        <button @click="toggle" class="btn">
          {{ enabled ? "Disable" : "Enable" }}
        </button>
      </div>
    </div>

    <div class="viz-wrap">
      <canvas
        ref="canvas"
        width="1600"
        height="300"
        style="width: 100%; height: 100%; display: block"
      ></canvas>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
const props = defineProps({
  analyser: Object, // AnalyserNode
  enabled: Boolean,
  volume: Number,
  fftSize: Number,
  smoothing: Number,
});
const emit = defineEmits([
  "update:enabled",
  "update:volume",
  "update:fftSize",
  "update:smoothing",
]);

const canvas = ref(null);
let animation = null;
let dataArray = null;

// mirrored local models (for v-model convenience)
const enabledLocal = ref(!!props.enabled);
const volumeLocal = ref(props.volume ?? 0.9);
const fftLocal = ref(props.fftSize ?? 1024);
const smoothingLocal = ref(props.smoothing ?? 0.6);

watch(
  () => props.enabled,
  (v) => (enabledLocal.value = v),
);
watch(
  () => props.volume,
  (v) => (volumeLocal.value = v),
);
watch(
  () => props.fftSize,
  (v) => (fftLocal.value = v),
);
watch(
  () => props.smoothing,
  (v) => (smoothingLocal.value = v),
);

watch(enabledLocal, (v) => emit("update:enabled", v));
watch(volumeLocal, (v) => emit("update:volume", v));
watch(fftLocal, (v) => emit("update:fftSize", v));
watch(smoothingLocal, (v) => emit("update:smoothing", v));

const enabled = enabledLocal;

function resizeCanvas() {
  const c = canvas.value;
  if (!c) return;
  const dpr = window.devicePixelRatio || 1;
  c.width = c.clientWidth * dpr;
  c.height = c.clientHeight * dpr;
}

function renderFrame() {
  if (!props.analyser || !enabled.value) {
    animation = requestAnimationFrame(renderFrame);
    return;
  }
  const ctx = canvas.value.getContext("2d");
  const bufferLength = props.analyser.frequencyBinCount;
  if (!dataArray || dataArray.length !== bufferLength) {
    dataArray = new Uint8Array(bufferLength);
  }
  props.analyser.getByteFrequencyData(dataArray);

  const w = canvas.value.width;
  const h = canvas.value.height;
  ctx.clearRect(0, 0, w, h);

  const bars = Math.floor(bufferLength / 2);
  const barWidth = w / bars;

  for (let i = 0; i < bars; i++) {
    const v = Math.pow(dataArray[i] / 255, 1.5);
    const barH = v * h * 0.9;
    const grd = ctx.createLinearGradient(0, h - barH, 0, h);
    grd.addColorStop(0, "#7c3aed");
    grd.addColorStop(1, "#06b6d4");
    ctx.fillStyle = grd;
    ctx.beginPath();
    const x = i * barWidth;
    const radius = barWidth * 0.3;
    ctx.moveTo(x + radius, h - barH);
    ctx.lineTo(x + barWidth - radius, h - barH);
    ctx.quadraticCurveTo(
      x + barWidth,
      h - barH,
      x + barWidth,
      h - barH + radius,
    );
    ctx.lineTo(x + barWidth, h);
    ctx.lineTo(x, h);
    ctx.closePath();
    ctx.fill();
  }

  animation = requestAnimationFrame(renderFrame);
}

onMounted(() => {
  resizeCanvas();
  window.addEventListener("resize", resizeCanvas);
  animation = requestAnimationFrame(renderFrame);
});

onBeforeUnmount(() => {
  cancelAnimationFrame(animation);
  window.removeEventListener("resize", resizeCanvas);
});

function toggle() {
  enabledLocal.value = !enabledLocal.value;
}
</script>
