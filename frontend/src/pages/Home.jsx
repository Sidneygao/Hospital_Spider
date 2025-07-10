import React, { useState, useEffect } from 'react';
import GoogleMapReact from 'google-map-react';
import { Input, Button, Spin, message, Tooltip } from 'antd';
import { EnvironmentOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import './Home.css';
import HospitalCard from '../components/HospitalCard';

const AERO_COLOR = 'rgba(180, 220, 255, 0.7)';
const CARD_SHADOW = '0 4px 24px 0 rgba(0, 40, 120, 0.12)';

const defaultCenter = { lat: 13.7466, lng: 100.5396 }; // 曼谷四面佛
const defaultZoom = 14;

const GOOGLE_GEOCODE_API = 'https://maps.googleapis.com/maps/api/geocode/json';
const GOOGLE_MAP_KEY = 'AIzaSyBAEb0pUFf624vcKQ0LDzhKZ_ntc2WfMwM';

// 红十字SVG组件
const RedCross = ({ size = 28 }) => (
  <svg width={size} height={size} viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="11" y="3" width="6" height="22" rx="3" fill="#e53935" />
    <rect x="3" y="11" width="22" height="6" rx="3" fill="#e53935" />
    <circle cx="14" cy="14" r="13" stroke="#fff" strokeWidth="2" fill="none" />
  </svg>
);

// 医院热点标记，悬停时浮现名称
const HospitalMarker = ({ text, active, onMouseEnter, onMouseLeave, onClick }) => (
  <div
    onMouseEnter={onMouseEnter}
    onMouseLeave={onMouseLeave}
    onClick={onClick}
    style={{
      position: 'relative',
      cursor: 'pointer',
      zIndex: active ? 10 : 1,
      transition: 'all 0.2s',
    }}
  >
    <RedCross size={active ? 36 : 28} />
    {active && (
      <div style={{
        position: 'absolute',
        left: '50%',
        top: '-38px',
        transform: 'translateX(-50%)',
        background: 'rgba(255,255,255,0.95)',
        color: '#e53935',
        fontWeight: 700,
        fontSize: 14,
        borderRadius: 8,
        boxShadow: '0 2px 12px rgba(0,0,0,0.13)',
        padding: '6px 16px',
        whiteSpace: 'nowrap',
        pointerEvents: 'none',
        border: '1.5px solid #e53935',
      }}>{text}</div>
    )}
  </div>
);

async function geocodeAddress(address) {
  const url = `${GOOGLE_GEOCODE_API}?address=${encodeURIComponent(address)}&key=${GOOGLE_MAP_KEY}`;
  try {
    const resp = await fetch(url);
    const data = await resp.json();
    if (data.status === 'OK' && data.results.length > 0) {
      const loc = data.results[0].geometry.location;
      return { lat: loc.lat, lng: loc.lng };
    }
    throw new Error('地标解析失败');
  } catch (e) {
    throw new Error('地标解析失败，可能是网络或API Key问题');
  }
}

export default function Home() {
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const [hospitals, setHospitals] = useState([]);
  const [center, setCenter] = useState(defaultCenter);
  const [zoom, setZoom] = useState(defaultZoom);
  const [activeId, setActiveId] = useState(null);

  // 启动时自动加载周边医院
  useEffect(() => {
    fetchHospitalsByLatLng(center.lat, center.lng);
    // eslint-disable-next-line
  }, []);

  // 通过经纬度请求后端API
  const fetchHospitalsByLatLng = async (lat, lng) => {
    setLoading(true);
    try {
      const resp = await fetch(`/api/hospitals/search?lat=${lat}&lng=${lng}&limit=10`);
      if (!resp.ok) throw new Error('后端API请求失败');
      const data = await resp.json();
      if (data.status === 'success') {
        setHospitals(data.data);
        setCenter({ lat, lng });
      } else {
        throw new Error('后端API返回异常');
      }
    } catch (e) {
      // fallback DUMMY
      setHospitals([
        {
          id: 1,
          name: '曼谷BNH医院',
          address: '9/1 Convent Rd, Silom, Bang Rak, Bangkok',
          lat: 13.7266,
          lng: 100.5346,
          rating: 4.7,
          type: '综合医院',
          distance: 2.1,
        },
        {
          id: 2,
          name: '曼谷Bumrungrad国际医院',
          address: '33 Sukhumvit 3, Khlong Toei Nuea, Watthana, Bangkok',
          lat: 13.7499,
          lng: 100.5567,
          rating: 4.6,
          type: '国际医院',
          distance: 3.5,
        },
        {
          id: 3,
          name: '曼谷Samitivej医院',
          address: '133 Sukhumvit 49, Khlong Tan Nuea, Watthana, Bangkok',
          lat: 13.7382,
          lng: 100.5766,
          rating: 4.5,
          type: '综合医院',
          distance: 4.2,
        },
      ]);
      message.error('后端API不可用，已使用模拟数据');
    }
    setLoading(false);
  };

  // 搜索地标/地址
  const handleSearch = async () => {
    if (!search.trim()) return;
    setLoading(true);
    try {
      // 先用Google Geocoding API转经纬度
      const loc = await geocodeAddress(search);
      setCenter(loc);
      fetchHospitalsByLatLng(loc.lat, loc.lng);
    } catch (e) {
      message.error(e.message || '地标解析失败，已使用默认位置');
      fetchHospitalsByLatLng(defaultCenter.lat, defaultCenter.lng);
    }
    setLoading(false);
  };

  // 定位到当前位置
  const handleLocate = () => {
    if (navigator.geolocation) {
      setLoading(true);
      navigator.geolocation.getCurrentPosition((pos) => {
        const loc = { lat: pos.coords.latitude, lng: pos.coords.longitude };
        setCenter(loc);
        fetchHospitalsByLatLng(loc.lat, loc.lng);
        setLoading(false);
      }, () => {
        setLoading(false);
        message.error('定位失败');
      });
    }
  };

  const handleRefresh = () => {
    fetchHospitalsByLatLng(center.lat, center.lng);
  };

  return (
    <div className="home-aero-bg">
      <div className="home-title">🏥 医院智能推荐平台</div>
      <div className="home-search-bar">
        <Input
          className="home-search-input"
          placeholder="输入地标/医院/地址..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          onPressEnter={handleSearch}
          allowClear
        />
        <Button
          className="home-search-btn"
          type="primary"
          icon={<SearchOutlined />}
          onClick={handleSearch}
        >搜索</Button>
        <Button
          className="home-locate-btn"
          icon={<EnvironmentOutlined />}
          onClick={handleLocate}
        >定位</Button>
      </div>
      <div className="home-map-section">
        <div className="home-map-container">
          <GoogleMapReact
            bootstrapURLKeys={{ key: GOOGLE_MAP_KEY }}
            center={center}
            zoom={zoom}
            yesIWantToUseGoogleMapApiInternals
          >
            {hospitals.map(h => (
              <HospitalMarker
                key={h.id}
                lat={h.lat}
                lng={h.lng}
                text={h.name}
                active={activeId === h.id}
                onMouseEnter={() => setActiveId(h.id)}
                onMouseLeave={() => setActiveId(null)}
                onClick={() => setActiveId(h.id)}
              />
            ))}
          </GoogleMapReact>
          <Button
            className="home-map-refresh"
            shape="circle"
            icon={<ReloadOutlined />}
            onClick={handleRefresh}
          />
        </div>
      </div>
      <div className="home-hospital-list">
        {loading ? <Spin size="large" /> : (
          hospitals.map(h => (
            <div
              key={h.id}
              onMouseEnter={() => setActiveId(h.id)}
              onMouseLeave={() => setActiveId(null)}
              style={{
                transition: 'box-shadow 0.2s',
                boxShadow: activeId === h.id ? '0 0 0 3px #2196f3' : 'none',
                borderRadius: 20,
              }}
            >
              <HospitalCard hospital={h} />
            </div>
          ))
        )}
      </div>
    </div>
  );
} 