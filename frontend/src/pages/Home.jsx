import { Modal } from 'antd';
import React, { useState, useEffect } from 'react';
import GoogleMapReact from 'google-map-react';
import { Input, Button, Spin, message, Tooltip } from 'antd';
import {
  EnvironmentOutlined,
  ReloadOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import './Home.css';
import HospitalCard from '../components/HospitalCard';

const AERO_COLOR = 'rgba(180, 220, 255, 0.7)';
const CARD_SHADOW = '0 4px 24px 0 rgba(0, 40, 120, 0.12)';

const defaultCenter = { lat: 1.2887, lng: 103.8007 }; // 新加坡Alexandra Hospital
const [center, setCenter] = useState(defaultCenter);
const defaultZoom = 14;

const GOOGLE_GEOCODE_API = 'https://maps.googleapis.com/maps/api/geocode/json';
const GOOGLE_MAP_KEY = 'AIzaSyBAEb0pUFf624vcKQ0LDzhKZ_ntc2WfMwM';

// 红十字SVG组件
const RedCross = ({ size = 28 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 28 28"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect x="11" y="3" width="6" height="22" rx="3" fill="#e53935" />
    <rect x="3" y="11" width="22" height="6" rx="3" fill="#e53935" />
    <circle cx="14" cy="14" r="13" stroke="#fff" strokeWidth="2" fill="none" />
  </svg>
);

// 医院热点标记，悬停时浮现名称
const HospitalMarker = ({
  text,
  active,
  onMouseEnter,
  onMouseLeave,
  onClick,
}) => (
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
    <svg
      width={active ? 36 : 28}
      height={active ? 36 : 28}
      viewBox="0 0 28 28"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect x="11" y="3" width="6" height="22" rx="3" fill="#e53935" />
      <rect x="3" y="11" width="22" height="6" rx="3" fill="#e53935" />
      <circle
        cx="14"
        cy="14"
        r="13"
        stroke="#fff"
        strokeWidth="2"
        fill="none"
      />
    </svg>
    {active && (
      <div
        style={{
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
        }}
      >
        {text}
      </div>
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

function getMapBounds(hospitals) {
  if (!hospitals.length) return null;
  let minLat = hospitals[0].lat,
    maxLat = hospitals[0].lat,
    minLng = hospitals[0].lng,
    maxLng = hospitals[0].lng;
  hospitals.forEach((h) => {
    if (h.lat < minLat) minLat = h.lat;
    if (h.lat > maxLat) maxLat = h.lat;
    if (h.lng < minLng) minLng = h.lng;
    if (h.lng > maxLng) maxLng = h.lng;
  });
  return { minLat, maxLat, minLng, maxLng };
}
function getMapCenterZoom(hospitals) {
  if (!hospitals.length) return { center: defaultCenter, zoom: defaultZoom };
  const bounds = getMapBounds(hospitals);
  const center = {
    lat: (bounds.minLat + bounds.maxLat) / 2,
    lng: (bounds.minLng + bounds.maxLng) / 2,
  };
  // 计算最大距离，确保5km内热点都能显示
  const R = 6371; // 地球半径km
  function haversine(lat1, lng1, lat2, lng2) {
    const toRad = (x) => (x * Math.PI) / 180;
    const dLat = toRad(lat2 - lat1);
    const dLng = toRad(lng2 - lng1);
    const a =
      Math.sin(dLat / 2) ** 2 +
      Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) * Math.sin(dLng / 2) ** 2;
    return 2 * R * Math.asin(Math.sqrt(a));
  }
  let maxDist = 0;
  for (let i = 0; i < hospitals.length; i++) {
    for (let j = i + 1; j < hospitals.length; j++) {
      const d = haversine(
        hospitals[i].lat,
        hospitals[i].lng,
        hospitals[j].lat,
        hospitals[j].lng,
      );
      if (d > maxDist) maxDist = d;
    }
  }
  // 5km内热点，zoom要能容纳maxDist+1km缓冲
  const dist = Math.max(maxDist, 5);
  // 经验公式，适配GoogleMapReact，zoom越大越近
  let zoom = 14;
  if (dist > 0) zoom = Math.floor(14 - Math.log2(dist / 2));
  if (zoom < 10) zoom = 10;
  if (zoom > 18) zoom = 18;
  return { center, zoom };
}
export default function Home() {
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const [hospitals, setHospitals] = useState([]);
  const [center, setCenter] = useState({ lat: 13.7466, lng: 100.5396 });
  const [zoom, setZoom] = useState(14);
  const [activeId, setActiveId] = useState(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailHospital, setDetailHospital] = useState(null);

  // 页面初始自动加载医院（默认中心点）
  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          const loc = { lat: pos.coords.latitude, lng: pos.coords.longitude };
          setCenter(loc);
          fetchHospitalsByLatLng(loc.lat, loc.lng);
        },
        () => {
          // 定位失败，使用默认中心点
          fetchHospitalsByLatLng(defaultCenter.lat, defaultCenter.lng);
        },
      );
    } else {
      fetchHospitalsByLatLng(defaultCenter.lat, defaultCenter.lng);
    }
    // eslint-disable-next-line
  }, []);

  // 通过经纬度请求后端API
 
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
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          const loc = { lat: pos.coords.latitude, lng: pos.coords.longitude };
          setCenter(loc); // 先更新center
          fetchHospitalsByLatLng(loc.lat, loc.lng);
          setLoading(false);
        },
        () => {
          setLoading(false);
          message.error('定位失败');
        },
      );
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
          onChange={(e) => setSearch(e.target.value)}
          onPressEnter={handleSearch}
          allowClear
        />
        <Button
          className="home-search-btn"
          type="primary"
          icon={<SearchOutlined />}
          onClick={handleSearch}
        >
          搜索
        </Button>
        <Button
          className="home-locate-btn"
          icon={<EnvironmentOutlined />}
          onClick={handleLocate}
        >
          定位
        </Button>
      </div>
      <div className="home-map-section">
        <div className="home-map-container square-map">
          <GoogleMapReact
            bootstrapURLKeys={{ key: GOOGLE_MAP_KEY }}
            center={center}
            zoom={zoom}
            yesIWantToUseGoogleMapApiInternals
          >
            {hospitals.map((h) => (
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
        {loading ? (
          <Spin size="large" />
        ) : (
          hospitals.map((h) => (
            <div
              key={h.id}
              onMouseEnter={() => setActiveId(h.id)}
              onMouseLeave={() => setActiveId(null)}
              style={{
                transition: 'box-shadow 0.2s',
                boxShadow: activeId === h.id ? '0 0 0 3px #2196f3' : 'none',
                borderRadius: 20,
              }}
              onClick={() => {
                setDetailHospital(h);
                setDetailOpen(true);
              }}
            >
              <HospitalCard hospital={h} />
            </div>
          ))
        )}
      </div>
      {/* ... existing code ...*/}
      <Modal
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={600}
        centered
        bodyStyle={{ padding: 0, borderRadius: 18, overflow: 'hidden' }}
        destroyOnClose
      >
        {detailHospital && (
          <div style={{ background: '#f8f6f1', minHeight: 360 }}>
            <div style={{ padding: 24 }}>
              <div
                style={{
                  fontSize: 22,
                  fontWeight: 700,
                  color: '#0d47a1',
                  marginBottom: 8,
                }}
              >
                {detailHospital.name}
              </div>
              <div style={{ color: '#1976d2', marginBottom: 8 }}>
                {detailHospital.address}
              </div>
              <div style={{ marginBottom: 8 }}>
                <span>类型：{detailHospital.type}</span> &nbsp;|&nbsp;
                <span>评分：{detailHospital.rating}</span> &nbsp;|&nbsp;
                <span>距离：{detailHospital.distance}km</span>
              </div>
              {detailHospital.phone && (
                <div style={{ marginBottom: 8 }}>
                  电话：
                  <a href={`tel:${detailHospital.phone}`}>
                    {detailHospital.phone}
                  </a>
                </div>
              )}
              {detailHospital.website && (
                <div style={{ marginBottom: 8 }}>
                  官网：
                  <a
                    href={detailHospital.website}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {detailHospital.website}
                  </a>
                </div>
              )}
            </div>
            <div style={{ width: '100%', height: 320 }}>
              <GoogleMapReact
                bootstrapURLKeys={{ key: GOOGLE_MAP_KEY }}
                center={{ lat: detailHospital.lat, lng: detailHospital.lng }}
                zoom={16}
                yesIWantToUseGoogleMapApiInternals
              >
                <HospitalMarker
                  lat={detailHospital.lat}
                  lng={detailHospital.lng}
                  text={detailHospital.name}
                  active={true}
                />
              </GoogleMapReact>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
// ... existing code ...
