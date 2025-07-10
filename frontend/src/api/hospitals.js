import axios from 'axios';

const API_BASE = process.env.REACT_APP_API_BASE_URL || 'http://localhost:3000/api';

export async function searchHospitals({ lat, lng, radius = 10, limit = 10 }) {
  const res = await axios.get(`${API_BASE}/hospitals/search`, {
    params: { lat, lng, radius, limit },
  });
  return res.data;
}
