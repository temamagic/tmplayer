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
            <div class="text-xs text-gray-400">upload</div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button class="btn" @click="router.back()">🔙 Back</button>
          <button class="btn" @click="router.push('/')">🏠 Home</button>
        </div>
      </div>
    </div>
  </header>

  <main class="flex-1 overflow-auto p-4 space-y-4 pt-20 pb-36">
    <div class="max-w-xl mx-auto p-6 space-y-6">
      <h1 class="text-2xl font-bold">Add track</h1>

      <form @submit.prevent="uploadTrack" class="space-y-4">
        <div>
          <label class="block mb-1 font-medium">Select track</label>
          <input
            type="file"
            @change="handleFile"
            accept="audio/*"
            required
            class="w-full"
          />
        </div>

        <button type="submit" :disabled="uploading" class="btn">
          {{ uploading ? "Uploading..." : "Upload" }}
        </button>

        <div
          v-if="uploadProgress > 0"
          class="w-full bg-gray-200 rounded h-2 mt-2"
        >
          <div
            class="bg-indigo-600 h-2 rounded"
            :style="{ width: uploadProgress + '%' }"
          ></div>
        </div>
      </form>

      <p v-if="message" class="mt-4 text-green-600 font-medium">
        {{ message }}
      </p>
      <p v-if="error" class="mt-4 text-red-600 font-medium">{{ error }}</p>
    </div>
  </main>
</template>

<script setup>
import { useRouter } from "vue-router";
import { ref } from "vue";
const router = useRouter();

const file = ref(null);
const uploading = ref(false);
const uploadProgress = ref(0);
const message = ref("");
const error = ref("");

const handleFile = (event) => {
  file.value = event.target.files[0];
};

const uploadTrack = async () => {
  if (!file.value) return;
  uploading.value = true;
  message.value = "";
  error.value = "";
  uploadProgress.value = 0;

  const formData = new FormData();
  formData.append("file", file.value);

  const xhr = new XMLHttpRequest();
  xhr.open("POST", "/api/tracks/add", true);

  xhr.upload.onprogress = (e) => {
    if (e.lengthComputable) {
      uploadProgress.value = Math.round((e.loaded / e.total) * 100);
    }
  };

  xhr.onload = () => {
    uploading.value = false;
    if (xhr.status >= 200 && xhr.status < 300) {
      message.value = "Track uploaded successful";
      file.value = null;
      uploadProgress.value = 0;
    } else {
      error.value = "Error: " + xhr.statusText;
    }
  };

  xhr.onerror = () => {
    uploading.value = false;
    error.value = "Net error";
  };

  xhr.send(formData);
};
</script>
