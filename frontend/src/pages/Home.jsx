import React, { useEffect, useState, useRef } from 'react';
import { Input, Button, message, Spin } from 'antd';
import { EnvironmentOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import HospitalCard from '../components/HospitalCard';
import { searchHospitals } from '../api/hospitals';
import './Home.css';
// import GoogleMapReact from 'google-map-react';
import HospitalDetailModal from '../components/HospitalDetailModal';
import '../components/HospitalDetailModal.css';

// 动态加载高德地图JS API，Key通过.env注入
function useAmapLoader() {
  useEffect(() => {
    if (window.AMap) return;
    const jsKey = process.env.REACT_APP_AMAP_JS_KEY;
    if (!jsKey) {
      console.error('高德JS API Key未配置');
      return;
    }
    const script = document.createElement('script');
    script.src = `https://webapi.amap.com/maps?v=2.0&key=${jsKey}&plugin=AMap.Marker,AMap.InfoWindow`;
    script.async = true;
    document.body.appendChild(script);
    return () => { document.body.removeChild(script); };
  }, []);
}

const DEFAULT_RADIUS = 10;
const MAP_CACHE_KEY = 'mapCenterCache';
const CACHE_EXPIRE_DAYS = 30;
const CACHE_DISTANCE_METERS = 500;
const typesMedical = '090100,090200,090300,090400,090500,090600';

// Haversine 公式计算两点距离（米）
function getDistanceMeters(lat1, lng1, lat2, lng2) {
  const R = 6371000; // 地球半径（米）
  const toRad = deg => deg * Math.PI / 180;
  const dLat = toRad(lat2 - lat1);
  const dLng = toRad(lng2 - lng1);
  const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) *
    Math.sin(dLng / 2) * Math.sin(dLng / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

// 医院数据缓存相关
const HOSPITAL_CACHE_KEY = 'hospitalDataCache';
const HOSPITAL_CACHE_RADIUS = 5000; // 5km
const HOSPITAL_CACHE_EXPIRE = 30 * 24 * 3600 * 1000; // 30天

function getCachedHospitals(center) {
  const cache = JSON.parse(localStorage.getItem(HOSPITAL_CACHE_KEY) || '[]');
  const now = Date.now();
  const validCache = cache.filter(item => now - item.timestamp < HOSPITAL_CACHE_EXPIRE);
  for (const item of validCache) {
    if (getDistanceMeters(center.lat, center.lng, item.center.lat, item.center.lng) < HOSPITAL_CACHE_RADIUS) {
      return item.data;
    }
  }
  return null;
}

function setCachedHospitals(center, data) {
  const cache = JSON.parse(localStorage.getItem(HOSPITAL_CACHE_KEY) || '[]');
  const now = Date.now();
  const validCache = cache.filter(item => now - item.timestamp < HOSPITAL_CACHE_EXPIRE);
  validCache.push({ center, radius: HOSPITAL_CACHE_RADIUS, data, timestamp: now });
  localStorage.setItem(HOSPITAL_CACHE_KEY, JSON.stringify(validCache));
}

// 红十字SVG组件
const RedCrossMarker = ({ name, onHover, onLeave, onClick, focused }) => (
  <div
    onMouseEnter={onHover}
    onMouseLeave={onLeave}
    onClick={onClick}
    style={{ cursor: 'pointer', position: 'relative', zIndex: focused ? 2 : 1 }}
  >
    <svg width="32" height="32" viewBox="0 0 32 32">
      <circle cx="16" cy="16" r="14" fill="#fff" stroke="#d32f2f" strokeWidth="2" />
      <rect x="13" y="7" width="6" height="18" rx="2" fill="#d32f2f" />
      <rect x="7" y="13" width="18" height="6" rx="2" fill="#d32f2f" />
    </svg>
    {focused && (
      <div style={{
        position: 'absolute',
        left: '50%',
        top: '-36px',
        transform: 'translateX(-50%)',
        background: 'rgba(255,255,255,0.95)',
        color: '#d32f2f',
        border: '1px solid #d32f2f',
        borderRadius: 6,
        padding: '4px 12px',
        fontSize: 14,
        whiteSpace: 'nowrap',
        boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
        zIndex: 10,
      }}>{name}</div>
    )}
  </div>
);

// 小蓝人+方向锥SVG定位点
const LocationMarker = () => (
  <div style={{ position: 'relative', width: 0, height: 0 }}>
    <svg width="40" height="40" viewBox="0 0 40 40" style={{ position: 'absolute', left: -20, top: -20 }}>
      {/* 方向锥 */}
      <path d="M20 20 L20 5 A15 15 0 0 1 35 20 Z" fill="rgba(30,144,255,0.25)" />
      {/* 小蓝人 */}
      <circle cx="20" cy="28" r="6" fill="#2196f3" stroke="#1565c0" strokeWidth="2" />
      <rect x="17" y="28" width="6" height="8" rx="3" fill="#2196f3" />
      <ellipse cx="20" cy="24" rx="4" ry="3" fill="#90caf9" />
    </svg>
  </div>
);

function getOffsetLatLng(lat, lng, distanceKm = 0.65, angleDeg = 25) {
  const R = 6371;
  const rad = (deg) => deg * Math.PI / 180;
  const dLat = (distanceKm / R) * Math.sin(rad(angleDeg));
  const dLng = (distanceKm / (R * Math.cos(rad(lat)))) * Math.cos(rad(angleDeg));
  return {
    lat: lat + dLat * 180 / Math.PI,
    lng: lng + dLng * 180 / Math.PI,
  };
}

// 兜底医院数据（仅含莱佛士和北医三院）
const DEFAULT_HOSPITALS = [
  {
    id: 'raffles',
    name: '北京莱佛士医院',
    address: '北京市朝阳区东直门南大街1号',
    latitude: 39.939073,
    longitude: 116.434564,
    type: '综合医院',
    rating: 4.5,
    distance: 0.0,
    phone: '010-64629111',
    website: 'https://www.rafflesmedical.com.cn/',
    tags: ['莱佛士', '国际医院', '综合医院'],
    intro: '北京莱佛士医院是一家国际化综合医院，提供高品质医疗服务。',
  },
  {
    id: 'puh3h',
    name: '北京大学第三医院',
    address: '北京市海淀区花园北路49号',
    latitude: 39.9753,
    longitude: 116.3541,
    type: '综合性三级甲等',
    rating: 4.6,
    distance: 7.5,
    phone: '010-82266699',
    website: 'https://www.puh3.net.cn/',
    tags: ['北医三院', '综合医院', '三级甲等'],
    intro: '北京大学第三医院（北医三院）是集医疗、教学、科研、预防、保健为一体的综合性三级甲等医院。',
  },
];

export default function Home() {
  const [address, setAddress] = useState('');
  const [location, setLocation] = useState({ lat: 39.9336, lng: 116.4402 }); // 默认北京东四十条
  const [searchLatLng, setSearchLatLng] = useState({ lng: 116.4402, lat: 39.9336 }); // 新增，默认与location一致
  const [hospitals, setHospitals] = useState([]);
  const [loading, setLoading] = useState(false);
  const [loadingLoc, setLoadingLoc] = useState(false);
  const mapContainerRef = useRef(null);
  const [focusedHospitalId, setFocusedHospitalId] = useState(null);
  const [isSample, setIsSample] = useState(false); // 新增
  const [sampleReason, setSampleReason] = useState(''); // 新增
  const [heatmap, setHeatmap] = useState([]); // 新增
  const [selectedHospital, setSelectedHospital] = useState(null);
  const [showDetail, setShowDetail] = useState(false);

  useAmapLoader();
  const mapRef = useRef(null);

  useEffect(() => {
    if (!window.AMap || !location || !mapRef.current) return;
    if (mapRef.current._amap_instance) {
      mapRef.current._amap_instance.destroy();
      mapRef.current._amap_instance = null;
    }
    const map = new window.AMap.Map(mapRef.current, {
      zoom: 14,
      center: [location.lng, location.lat],
      viewMode: '2D',
    });
    mapRef.current._amap_instance = map;
    // 渲染定位点
    new window.AMap.Marker({
      position: [location.lng, location.lat],
      map,
      icon: new window.AMap.Icon({
        image: 'https://webapi.amap.com/theme/v1.3/markers/n/loc.png',
        size: [32, 32],
      }),
      title: '当前位置',
    });
    // 渲染医院marker
    hospitals.forEach(hospital => {
      const marker = new window.AMap.Marker({
        position: [hospital.longitude, hospital.latitude],
        map,
        title: hospital.name,
      });
      const info = new window.AMap.InfoWindow({
        content: `<b>${hospital.name}</b><br/>${hospital.address}`,
      });
      marker.on('click', () => info.open(map, marker.getPosition()));
    });
    return () => { map && map.destroy(); };
  }, [location, hospitals]);
 

  // 进入页面时尝试读取缓存
  useEffect(() => {
    const cacheStr = localStorage.getItem(MAP_CACHE_KEY);
    if (cacheStr) {
      try {
        const cache = JSON.parse(cacheStr);
        const now = Date.now();
        if (cache.timestamp && now - cache.timestamp < CACHE_EXPIRE_DAYS * 24 * 3600 * 1000) {
          // 获取当前定位
          if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(
              pos => {
                const dist = getDistanceMeters(
                  cache.lat, cache.lng,
                  pos.coords.latitude, pos.coords.longitude,
                );
                if (dist < CACHE_DISTANCE_METERS) {
                  setLocation({ lat: cache.lat, lng: cache.lng });
                  return;
                } else {
                  setLocation({ lat: pos.coords.latitude, lng: pos.coords.longitude });
                }
              },
              () => {
                setLocation({ lat: cache.lat, lng: cache.lng });
              },
            );
          } else {
            setLocation({ lat: cache.lat, lng: cache.lng });
          }
        } else {
          localStorage.removeItem(MAP_CACHE_KEY);
        }
      } catch (e) {
        // 捕获异常但不做处理，避免空代码块
      }
    } else {
      handleLocate();
    }
    // eslint-disable-next-line
  }, []);

  // 地图中心变化时写入缓存
  useEffect(() => {
    if (location) {
      localStorage.setItem(MAP_CACHE_KEY, JSON.stringify({
        lat: location.lat,
        lng: location.lng,
        timestamp: Date.now(),
      }));
    }
  }, [location]);

  // 在useEffect中，地图中心变化或搜索时：
  useEffect(() => {
    if (!location) return;
    setLoading(true);
    const fetchHospitals = async () => {
      try {
        const amapKey = process.env.REACT_APP_AMAP_KEY;
        const poisRes = await fetch(`/api/amap/around?location=${searchLatLng.lng},${searchLatLng.lat}&radius=5000&types=${typesMedical}`);
        const poisData = await poisRes.json();
        if (poisData.status === '1' && poisData.pois && poisData.pois.length > 0) {
          const hospitals = poisData.pois.map((poi, idx) => ({
            id: poi.id || idx,
            name: poi.name,
            address: poi.address,
            latitude: parseFloat(poi.location.split(',')[1]),
            longitude: parseFloat(poi.location.split(',')[0]),
            type: poi.type,
            rating: 0,
            distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
            phone: poi.tel,
            website: '',
            tags: [poi.type],
            intro: poi.name + (poi.type ? `（${poi.type}）` : ''),
            _sample: false,
          }));
          setHospitals(hospitals);
          setIsSample(false);
        } else {
          setHospitals(DEFAULT_HOSPITALS);
          setIsSample(true);
          setSampleReason('高德API无结果');
        }
      } catch (e) {
        setHospitals(DEFAULT_HOSPITALS);
        setIsSample(true);
        setSampleReason('高德API异常');
      }
      setLoading(false);
    };
    fetchHospitals();
  }, [location, searchLatLng]); // 依赖加上searchLatLng

  // 通过浏览器API获取定位
  const handleLocate = () => {
    setLoadingLoc(true);
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        pos => {
          setLocation({ lat: pos.coords.latitude, lng: pos.coords.longitude });
          setLoadingLoc(false);
          message.success('定位成功');
        },
        err => {
          setLoadingLoc(false);
          message.error('定位失败，请手动输入地址');
        },
      );
    } else {
      setLoadingLoc(false);
      message.error('浏览器不支持定位');
    }
  };

  // 搜索医院
  const handleSearch = async () => {
    console.log('handleSearch triggered', address, location);
    if (!address && !location) {
      message.warning('请先定位或输入地址');
      return;
    }
    setLoading(true);
    let newSearchLatLng = location;
    // 若有address，先用高德地理编码API获取经纬度
    if (address) {
      try {
        const amapKey = process.env.REACT_APP_AMAP_KEY;
        console.log('fetching geo', address);
        const geoRes = await fetch(`/api/amap/geo?address=${encodeURIComponent(address)}`);
        const geoData = await geoRes.json();
        console.log('geoData', geoData);
        if (geoData.status === '1' && geoData.geocodes && geoData.geocodes.length > 0) {
          const loc = geoData.geocodes[0].location.split(',');
          newSearchLatLng = { lng: parseFloat(loc[0]), lat: parseFloat(loc[1]) };
          setLocation(newSearchLatLng);
        } else {
          message.error('地址定位失败，已用当前位置');
        }
      } catch (e) {
        message.error('高德地理编码失败，已用当前位置');
      }
    }
    setSearchLatLng(newSearchLatLng); // 新增，确保searchLatLng同步
    // 用高德周边搜索API查找5KM内医院
    try {
      const amapKey = process.env.REACT_APP_AMAP_KEY;
      console.log('fetching around', newSearchLatLng);
      const poisRes = await fetch(`/api/amap/around?location=${newSearchLatLng.lng},${newSearchLatLng.lat}&radius=5000&types=${typesMedical}`);
      const poisData = await poisRes.json();
      console.log('poisData', poisData);
      if (poisData.status === '1' && poisData.pois && poisData.pois.length > 0) {
        const hospitals = poisData.pois.map((poi, idx) => ({
          id: poi.id || idx,
          name: poi.name,
          address: poi.address,
          latitude: parseFloat(poi.location.split(',')[1]),
          longitude: parseFloat(poi.location.split(',')[0]),
          type: poi.type,
          rating: 0,
          distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
          phone: poi.tel,
          website: '',
          tags: [poi.type],
          intro: poi.name + (poi.type ? `（${poi.type}）` : ''),
          _sample: false,
        }));
        setHospitals(hospitals);
        setIsSample(false);
      } else {
        message.error('未获取到医院数据，已用SAMPLE兜底');
        setHospitals(DEFAULT_HOSPITALS);
        setIsSample(true);
        setSampleReason('高德API无结果');
      }
    } catch (e) {
      message.error('医院数据获取失败，已用SAMPLE兜底');
      setHospitals(DEFAULT_HOSPITALS);
      setIsSample(true);
      setSampleReason('高德API异常');
    }
    setLoading(false);
  };

  // 优化地图容器为正方形自适应
  // 修改renderMap函数，只显示定位信息和医院列表，不渲染GoogleMapReact
  const renderMap = () => (
    <div
      className="home-map-container"
      style={{ width: '100%', aspectRatio: '1/1', maxWidth: 600, margin: '0 auto', background: '#f5f5f5', borderRadius: 16, boxShadow: '0 4px 16px rgba(0,0,0,0.08)', overflow: 'hidden', minHeight: 320 }}
    >
      <div id="map-container" ref={mapRef} style={{ width: '100%', height: '100%' }} />
    </div>
  );

  return (
    <div className="home-aero-bg">
      <div className="home-title">医院智能推荐平台</div>
      {isSample && (
        <div style={{ textAlign: 'center', margin: '12px 0' }}>
          <span style={{ color: '#d32f2f', fontWeight: 'bold', fontSize: 16 }}>当前为兜底SAMPLE数据，仅供参考{sampleReason ? `（${sampleReason}）` : ''}</span>
        </div>
      )}
      <div className="home-search-bar">
        <Input
          className="home-search-input"
          placeholder="请输入地址或医院名"
          value={address}
          onChange={e => setAddress(e.target.value)}
          disabled={loadingLoc}
        />
        <Button
          className="home-locate-btn"
          icon={<EnvironmentOutlined />}
          loading={loadingLoc}
          onClick={handleLocate}
        >定位</Button>
        <Button
          className="home-search-btn"
          icon={<SearchOutlined />}
          type="primary"
          loading={loading}
          onClick={handleSearch}
        >搜索</Button>
        <Button
          className="home-search-btn"
          icon={<ReloadOutlined />}
          onClick={handleSearch}
        >刷新</Button>
      </div>
      <div className="home-map-section">
        <div className="home-map-container">
          {renderMap()}
        </div>
      </div>
      <div className="home-hospital-list">
        {loading ? (
          <Spin style={{ width: '100%', margin: '32px auto', display: 'block' }} />
        ) : (
          hospitals.length === 0 ? (
            <div style={{ textAlign: 'center', color: '#888', margin: '32px 0' }}>暂无医院数据</div>
          ) : (
            hospitals.map(hospital => (
              <HospitalCard
                key={hospital.id}
                hospital={hospital}
                onClick={() => {
                  setSelectedHospital(hospital);
                  setShowDetail(true);
                }}
              />
            ))
          )
        )}
      </div>
      <HospitalDetailModal
        visible={showDetail}
        hospital={selectedHospital}
        onClose={() => setShowDetail(false)}
      />
    </div>
  );
}
