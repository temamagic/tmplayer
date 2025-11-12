// адаптируй базовый путь, если нужно
export async function fetchSongsApi(offset = 0, limit = 5) {
  try {
    const res = await fetch(`/api/tracks?offset=${offset}&limit=${limit}`);
    if (!res.ok) return [];
    const data = await res.json();
    // data should be array of { id, title, artist, src, cover }
    return data;
  } catch (e) {
    console.error("fetchSongsApi error", e);
    return [];
  }
}
