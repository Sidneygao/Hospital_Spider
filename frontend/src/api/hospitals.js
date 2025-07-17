import axios from 'axios';

const API_BASE =
  process.env.REACT_APP_API_BASE_URL || '/api';

export async function searchHospitals({ lat, lng, radius = 10, limit = 10 }) {
  const res = await axios.get(`${API_BASE}/places/hospitals`, {
    params: { lat, lng, radius: radius * 1000, limit },
  });
  if (res.data && Array.isArray(res.data.results)) {
    // Google API风格
    return res.data.results;
  } else if (Array.isArray(res.data)) {
    // 兜底
    return res.data;
  } else if (res.data && Array.isArray(res.data.data)) {
    // 兼容后端自定义结构
    return res.data.data;
  }
  return [];
}

export async function placesSearch(query) {
  const res = await fetch(`/api/google/places/search?query=${encodeURIComponent(query)}`);
  return res.json();
}

export async function getStaticMap(center, zoom, size, marker) {
  let url = `/api/google/staticmap?center=${encodeURIComponent(center)}&zoom=${zoom}&size=${size}`;
  if (marker) {
    // 使用Google官方蓝色marker
    url += `&markers=color:blue|label:P|${marker.lat},${marker.lng}`;
  }
  const res = await fetch(url);
  return res.blob();
}
