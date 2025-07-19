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
// 高德API types参数，排除医美相关类别
const typesMedical = '090100,090300,090400,090500,090600'; // 排除090201-090205（医美/美容相关）
const excludeKeywords = ['美容', '医美', '整形', '美体', '美发', '美甲', 'SPA', '瘦身', '塑形', '康复', '药店', '药房', '大药房', '连锁药店'];
const clinicKeywords = ['社区', '健康中心', '保健中心', '卫生'];

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
        fontSize: 12,
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

// 删除所有与DUMMY医院、莱佛士医院、兜底医院相关的注释和变量
// 红十字icon SVG（BOLD/普通/小号）
const redCrossBold = `data:image/svg+xml;utf8,${encodeURIComponent(`
  <svg width="27" height="27" viewBox="0 0 27 27" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="13.5" cy="13.5" r="12" fill="#fff" stroke="#d32f2f" stroke-width="2.25"/>
    <rect x="11.25" y="5.25" width="4.5" height="16.5" rx="1.875" fill="#d32f2f"/>
    <rect x="5.25" y="11.25" width="16.5" height="4.5" rx="1.875" fill="#d32f2f"/>
  </svg>
`)}
`;
const redCrossNormal = `data:image/svg+xml;utf8,${encodeURIComponent(`
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="12" cy="12" r="10.5" fill="#fff" stroke="#d32f2f" stroke-width="1.5"/>
    <rect x="9.75" y="5.25" width="4.5" height="13.5" rx="1.5" fill="#d32f2f"/>
    <rect x="5.25" y="9.75" width="13.5" height="4.5" rx="1.5" fill="#d32f2f"/>
  </svg>
`)}
`;
const redCrossSmall = `data:image/svg+xml;utf8,${encodeURIComponent(`
  <svg width="18" height="18" viewBox="0 0 18 18" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="9" cy="9" r="7.5" fill="#fff" stroke="#d32f2f" stroke-width="1.125"/>
    <rect x="7.125" y="3.75" width="3" height="10.5" rx="1.125" fill="#d32f2f"/>
    <rect x="3.75" y="7.125" width="10.5" height="3" rx="1.125" fill="#d32f2f"/>
  </svg>
`)}
`;

// 牙齿SVG icon
const toothIcon = `data:image/svg+xml;utf8,${encodeURIComponent(`
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M4.5,12 Q3,6 7.5,4.5 Q12,3 16.5,4.5 Q21,6 19.5,12 Q18,21 12,21 Q6,21 4.5,12 Z" stroke="#1e90ff" stroke-width="1.9" fill="none"/>
    <path d="M9,18 Q12,15 15,18" stroke="#1e90ff" stroke-width="1.1" fill="none"/>
  </svg>
`)}
`;

// 医院类型过滤和icon选择
const hospitalTypeMap = [
  { key: '综合医院', icon: redCrossBold, size: [27, 27] },
  { key: '专科医院', icon: redCrossNormal, size: [24, 24] },
  { key: '诊所', icon: redCrossSmall, size: [18, 18] },
];
// 格式化和功能修正：
// 1. 社区医院只要name含“社区/健康中心/保健中心”优先归为诊所（小红十字），绝不归为综合医院
// 2. type含“美容”或“医美”彻底过滤
// 3. 牙科icon大小与诊所一致
// 4. 医院列表排序为综合医院>专科医院>牙科>诊所（社区医院和诊所合并为诊所）
const filterHospitals = (hospitals) => {
  return hospitals.filter(h => {
    const text = (h.name || '') + (h.type || '') + (h.address || '');
    // 医美/康复关键字过滤
    if (excludeKeywords.some(keyword => text.includes(keyword))) {
      return false;
    }
    // 只保留三类+社区/健康/保健
    if (
      (h.name && clinicKeywords.some(keyword => h.name.includes(keyword))) ||
      (h.type && (h.type.includes('综合医院') || h.type.includes('专科医院') || h.type.includes('诊所') || h.type.includes('社区卫生服务中心') || h.type.includes('卫生服务中心')))
    ) {
      return true;
    }
    return false;
  });
};
const sortHospitals = (hospitals) => {
  return hospitals.slice().sort((a, b) => {
    // 优先使用后端返回的排序字段
    const orderA = a.algo_display_order || 99;
    const orderB = b.algo_display_order || 99;
    
    // 如果排序字段相同，按距离排序（离观测点越近越靠前）
    if (orderA === orderB) {
      const distanceA = a.distance || 999;
      const distanceB = b.distance || 999;
      return distanceA - distanceB;
    }
    
    return orderA - orderB;
  });
};

// 牙齿PNG icon和动态定位GIF路径
const toothPng = '/tooth.png'; // 需将tooth.png放到public目录
const locateGif = '/locate_animated.gif'; // 需将locate_animated.gif放到public目录

// SVG小人+锥形icon转为DataURL（放大为48x48）
const locationIconDataUrl = `data:image/svg+xml;utf8,${encodeURIComponent(`
  <svg width="48" height="72" viewBox="0 0 48 72" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M24 36 L24 6 A27 27 0 0 1 42 36 Z" fill="rgba(30,144,255,0.25)"/>
    <circle cx="24" cy="58" r="7" fill="#2196f3" stroke="#1565c0" stroke-width="2"/>
    <rect x="20" y="58" width="8" height="10" rx="4" fill="#2196f3"/>
    <ellipse cx="24" cy="53" rx="5" ry="4" fill="#90caf9"/>
  </svg>
`)}
`;

// POI缓冲去重：经纬度差<100米且名称去除“北京”等公用字后重叠超4字，只保留名称最短的，其它加tag:OptOut
function normalizeName(name) {
  return (name || '').replace(/北京|市|区|县|医院|诊所|卫生院|社区|服务中心/g, '');
}
function isClose(lat1, lng1, lat2, lng2) {
  const R = 6371000;
  const toRad = deg => deg * Math.PI / 180;
  const dLat = toRad(lat2 - lat1);
  const dLng = toRad(lng2 - lng1);
  const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) + Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) * Math.sin(dLng / 2) * Math.sin(dLng / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c < 200;
}
function getNameOverlap(n1, n2) {
  let overlap = '';
  for (const c of n1) if (n2.includes(c) && !overlap.includes(c)) overlap += c;
  return overlap;
}
function dedupHospitals(hospitals) {
  const result = [];
  const used = new Set();
  function normalizeName(name) {
    return (name || '').replace(/医院|分院|门诊部|住院楼|本部|分部|大楼|楼|栋|（.*?）|\(.*?\)/g, '').replace(/\s/g, '').toLowerCase();
  }
  function getDistanceMeters(lat1, lng1, lat2, lng2) {
    const R = 6371000;
    const toRad = deg => deg * Math.PI / 180;
    const dLat = toRad(lat2 - lat1);
    const dLng = toRad(lng2 - lng1);
    const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) *
      Math.sin(dLng / 2) * Math.sin(dLng / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }
  for (let i = 0; i < hospitals.length; ++i) {
    if (used.has(i)) continue;
    const group = [hospitals[i]];
    const n1 = normalizeName(hospitals[i].name);
    const t1 = String(hospitals[i].typecode || '').slice(0, 5);
    for (let j = i + 1; j < hospitals.length; ++j) {
      if (used.has(j)) continue;
      const n2 = normalizeName(hospitals[j].name);
      const t2 = String(hospitals[j].typecode || '').slice(0, 5);
      const dist = getDistanceMeters(hospitals[i].latitude, hospitals[i].longitude, hospitals[j].latitude, hospitals[j].longitude);
      if (n1 === n2 && t1 === t2 && dist < 100) {
        group.push(hospitals[j]);
        used.add(j);
      }
    }
    // 合并策略：保留group[0]，如需更复杂可自定义
    result.push(group[0]);
  }
  return result;
}

// 将tryFitView函数移到useEffect外部
function tryFitView(map, lnglats, filtered) {
  let tryCount = 0;
  function inner() {
    tryCount++;
    const overlays = map.getAllOverlays('marker');
    if (overlays.length >= filtered.length || tryCount > 10) {
      map.setFitView(lnglats, false, [80, 80, 80, 80]);
    } else {
      requestAnimationFrame(inner);
    }
  }
  inner();
}

// 在文件顶部或组件内添加渲染函数
function renderChildType(childtype) {
  if (Array.isArray(childtype)) {
    return childtype.length === 0 ? '[]' : childtype.join(',');
  }
  if (typeof childtype === 'number') {
    return childtype.toString();
  }
  if (typeof childtype === 'string') {
    return childtype === '' ? '""' : childtype;
  }
  return String(childtype);
}
function renderTypeCode(typecode) {
  if (typeof typecode === 'string') {
    return typecode === '' ? '""' : typecode;
  }
  if (typeof typecode === 'number') {
    return typecode.toString();
  }
  return String(typecode);
}

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
  const [showLocateTip, setShowLocateTip] = useState(true);

  useAmapLoader();
  const mapRef = useRef(null);

  // 首页加载时自动定位，并高亮提示按钮
  useEffect(() => {
    handleLocate();
    setTimeout(() => setShowLocateTip(false), 5000); // 5秒后自动隐藏提示
  }, []);

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
    // 渲染定位点icon为SVG小蓝人+加长锥形
    const iconObj = new window.AMap.Icon({
      image: locationIconDataUrl,
      size: new window.AMap.Size(48, 72),
      imageSize: new window.AMap.Size(48, 72),
    });
    new window.AMap.Marker({
      position: [location.lng, location.lat],
      map,
      icon: iconObj,
      title: '当前位置',
      offset: new window.AMap.Pixel(-24, -36),
      zIndex: 100,
    });
    // 医院marker及InfoWindow（仅三类+牙科+社区医院归诊所）
    const filtered = filterHospitals(hospitals);
    const markers = [];
    filtered.forEach(hospital => {
      // 严格按hospital_category（小表第四列）适配icon
      let icon = redCrossNormal, size = [24, 24]; // Changed from 32 to 24
      const cat = hospital.algo_hospital_category || '';
      if (cat.includes('三级甲等')) { icon = redCrossBold; size = [27, 27]; } // Changed from 36 to 27
      else if (cat.includes('综合医院')) { icon = redCrossNormal; size = [27, 27]; } // Changed from 36 to 27
      else if (cat.includes('专科医院')) { icon = redCrossBold; size = [18, 18]; } // Changed from 24 to 18
      else if (cat.includes('卫生院') || cat.includes('社区')) { icon = redCrossNormal; size = [18, 18]; } // Changed from 24 to 18
      // 牙科特殊处理
      if (hospital.type && (hospital.type.includes('口腔') || hospital.type.includes('牙科'))) {
        icon = toothPng; size = [18, 18];
      }
      // 医院级别高亮
      let level = '';
      if (hospital.type.match(/(三级甲等|三甲|二级甲等|三级|二级|一级)/)) {
        level = hospital.type.match(/(三级甲等|三甲|二级甲等|三级|二级|一级)/)[0];
      }
      const marker = new window.AMap.Marker({
        position: [hospital.longitude, hospital.latitude],
        map,
        title: hospital.name,
        zIndex: 10,
        icon: new window.AMap.Icon({
          image: icon,
          size: new window.AMap.Size(...size),
          imageSize: new window.AMap.Size(...size),
        }),
        offset: new window.AMap.Pixel(-size[0]/2, -size[1]/2),
      });
      const info = new window.AMap.InfoWindow({
        content: `<b>${hospital.name}${level ? ` <span style='color:#d32f2f;font-weight:bold'>${level}</span>` : ''}</b><br/>${hospital.address}`,
        offset: new window.AMap.Pixel(0, -30),
      });
      marker.on('mouseover', () => info.open(map, marker.getPosition()));
      marker.on('mouseout', () => info.close());
      marker.on('click', () => info.open(map, marker.getPosition()));
      markers.push(marker);
    });
    // 综合医院识别更宽泛，type和name字段都用正则/(三甲|三级甲等|综合|协和)/i
    const isGeneralHospital = h =>
      /(三甲|三级甲等|综合|协和)/i.test(h.type || '') ||
      /(三甲|三级甲等|综合|协和)/i.test(h.name || '');
    const generalHospitals = filtered.filter(isGeneralHospital);
    if (generalHospitals.length === 3) {
      const lnglats = generalHospitals.map(h => new window.AMap.LngLat(h.longitude, h.latitude));
      tryFitView(map, lnglats, filtered);
    }
    map.on('zoomend', () => {
      if (map.getZoom() > 16) map.setZoom(16);
      if (map.getZoom() < 12) map.setZoom(12);
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
    console.log('useEffect被触发，location:', location, 'searchLatLng:', searchLatLng);
    if (!location) return;
    setLoading(true);
    const fetchHospitals = async () => {
      try {
        // 优先尝试获取合并后POI及TAG
        console.log('开始调用 /api/merged-pois');
        const mergedRes = await fetch('/api/merged-pois');
        console.log('API响应状态:', mergedRes.status);
        if (mergedRes.ok) {
          const mergedData = await mergedRes.json();
          console.log('API返回数据:', mergedData);
          if (mergedData.pois && mergedData.pois.length > 0) {
            // 直接用合并后POI
            const hospitals = mergedData.pois.map((poi, idx) => {
              const hospital = {
                id: poi.id || idx,
                name: poi.name,
                address: poi.address,
                latitude: poi.location_lat || (poi.location && parseFloat((`${poi.location}`).split(',')[1])),
                longitude: poi.location_lng || (poi.location && parseFloat((`${poi.location}`).split(',')[0])),
                type: poi.type,
                algo_hospital_category: poi.algo_hospital_category, // 只用算法字段
                algo_icon_type: poi.algo_icon_type, // 只用算法字段
                algo_display_order: poi.algo_display_order, // 只用算法字段
                // POI字段直接用typecode
                typecode: poi.typecode || '未知',
                childtype: poi.childtype || '未知',
                rating: 0,
                distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
                phone: poi.tel,
                website: poi.website,
                tags: poi.tags,
                intro: poi.intro,
                _sample: poi._sample,
              };
              console.log(`医院 ${hospital.name} 的typecode:`, hospital.typecode, 'childtype:', hospital.childtype);
              return hospital;
            });
            setHospitals(
              hospitals
                .filter(h => !!h.algo_hospital_category && typeof h.algo_hospital_category === 'string' && h.algo_hospital_category.trim() !== '' && h.algo_hospital_category !== '无' && h.algo_hospital_category !== 'null')
                .filter(h => !(h.tags && h.tags.includes('OptOut'))),
            );
            setIsSample(false);
            setLoading(false);
            return;
          }
        }
        // 兜底：高德API
        const amapKey = process.env.REACT_APP_AMAP_KEY;
        const poisRes = await fetch(`/api/amap/around?location=${searchLatLng.lng},${searchLatLng.lat}&radius=5000`);
        const poisData = await poisRes.json();
        if (poisData.status === '1' && poisData.pois && poisData.pois.length > 0) {
          const pois = poisData.pois.filter(poi => {
            const text = (poi.name || '') + (poi.type || '') + (poi.address || '');
            return !excludeKeywords.some(keyword => text.includes(keyword));
          });
          const hospitals = dedupHospitals(pois.map((poi, idx) => ({
            id: poi.id || idx,
            name: poi.name,
            address: poi.address,
            latitude: parseFloat(poi.location.split(',')[1]),
            longitude: parseFloat(poi.location.split(',')[0]),
            type: poi.type,
            algo_hospital_category: poi.hospital_category,
            algo_icon_type: poi.icon_type,
            rating: 0,
            distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
            phone: poi.tel,
            website: '',
            tags: [poi.type],
            intro: poi.name + (poi.type ? `（${poi.type}）` : ''),
            _sample: false,
          })));
          setHospitals(
            hospitals
              .filter(h => h.algo_hospital_category && h.algo_hospital_category !== '')
              .filter(h => !(h.tags && h.tags.includes('OptOut'))),
          );
          setIsSample(false);
        } else {
          setHospitals([]);
          setIsSample(false);
          message.warning('高德API无结果，请尝试其他搜索方式或调整搜索范围。');
        }
      } catch (e) {
        setHospitals([]);
        setIsSample(false);
        message.error('医院数据获取失败，请稍后再试。');
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
        async pos => {
          const newLoc = { lat: pos.coords.latitude, lng: pos.coords.longitude };
          console.log('定位成功，新位置:', newLoc);
          setLocation(newLoc);
          setSearchLatLng(newLoc); // 同步更新searchLatLng，保证地图和医院数据都以新定位为中心
          setLoadingLoc(false);
          message.success('定位成功');
          
          // 立即重新获取医院数据，不依赖useEffect的异步更新
          console.log('定位后立即重新获取医院数据...');
          setLoading(true);
          try {
            // 直接调用合并POI API
            const mergedRes = await fetch(`/api/merged-pois?location=${newLoc.lng},${newLoc.lat}&radius=5000`);
            console.log('合并POI API响应状态:', mergedRes.status);
            if (mergedRes.ok) {
              const mergedData = await mergedRes.json();
              console.log('合并POI API返回数据:', mergedData);
              if (mergedData.pois && mergedData.pois.length > 0) {
                // 直接用合并后POI
                const hospitals = mergedData.pois.map((poi, idx) => {
                  const hospital = {
                    id: poi.id || idx,
                    name: poi.name,
                    address: poi.address,
                    latitude: poi.location_lat || (poi.location && parseFloat((`${poi.location}`).split(',')[1])),
                    longitude: poi.location_lng || (poi.location && parseFloat((`${poi.location}`).split(',')[0])),
                    type: poi.type,
                    algo_hospital_category: poi.algo_hospital_category, // 只用算法字段
                    algo_icon_type: poi.algo_icon_type, // 只用算法字段
                    algo_display_order: poi.algo_display_order, // 只用算法字段
                    // POI字段直接用typecode
                    typecode: poi.typecode || '未知',
                    childtype: poi.childtype || '未知',
                    rating: 0,
                    distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
                    phone: poi.tel,
                    website: poi.website,
                    tags: poi.tags,
                    intro: poi.intro,
                    _sample: poi._sample,
                  };
                  console.log(`医院 ${hospital.name} 的typecode:`, hospital.typecode, 'childtype:', hospital.childtype);
                  return hospital;
                });
                setHospitals(
                  hospitals
                    .filter(h => !!h.algo_hospital_category && typeof h.algo_hospital_category === 'string' && h.algo_hospital_category.trim() !== '' && h.algo_hospital_category !== '无' && h.algo_hospital_category !== 'null')
                    .filter(h => !(h.tags && h.tags.includes('OptOut'))),
                );
                setIsSample(false);
                setLoading(false);
                return;
              }
            }
            // 兜底：高德API
            const amapKey = process.env.REACT_APP_AMAP_KEY;
            const poisRes = await fetch(`/api/amap/around?location=${newLoc.lng},${newLoc.lat}&radius=5000`);
            const poisData = await poisRes.json();
            if (poisData.status === '1' && poisData.pois && poisData.pois.length > 0) {
              const pois = poisData.pois.filter(poi => {
                const text = (poi.name || '') + (poi.type || '') + (poi.address || '');
                return !excludeKeywords.some(keyword => text.includes(keyword));
              });
              const hospitals = dedupHospitals(pois.map((poi, idx) => ({
                id: poi.id || idx,
                name: poi.name,
                address: poi.address,
                latitude: parseFloat(poi.location.split(',')[1]),
                longitude: parseFloat(poi.location.split(',')[0]),
                type: poi.type,
                algo_hospital_category: poi.hospital_category,
                algo_icon_type: poi.icon_type,
                rating: 0,
                distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
                phone: poi.tel,
                website: '',
                tags: [poi.type],
                intro: poi.name + (poi.type ? `（${poi.type}）` : ''),
                _sample: false,
              })));
              setHospitals(
                hospitals
                  .filter(h => h.algo_hospital_category && h.algo_hospital_category !== '')
                  .filter(h => !(h.tags && h.tags.includes('OptOut'))),
              );
              setIsSample(false);
            } else {
              setHospitals([]);
              setIsSample(false);
              message.warning('高德API无结果，请尝试其他搜索方式或调整搜索范围。');
            }
          } catch (e) {
            console.error('定位后获取医院数据失败:', e);
            setHospitals([]);
            setIsSample(false);
            message.error('医院数据获取失败，请稍后再试。');
          }
          setLoading(false);
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
        const geoRes = await fetch(`/api/amap/geo?address=${encodeURIComponent(address)}`);
        const geoData = await geoRes.json();
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
    console.log('搜索位置:', newSearchLatLng);
    
    // 立即获取医院数据，不依赖useEffect
    try {
      // 优先使用合并POI API
      const mergedRes = await fetch(`/api/merged-pois?location=${newSearchLatLng.lng},${newSearchLatLng.lat}&radius=5000`);
      console.log('搜索时合并POI API响应状态:', mergedRes.status);
      if (mergedRes.ok) {
        const mergedData = await mergedRes.json();
        console.log('搜索时合并POI API返回数据:', mergedData);
        if (mergedData.pois && mergedData.pois.length > 0) {
          // 直接用合并后POI
          const hospitals = mergedData.pois.map((poi, idx) => {
            const hospital = {
              id: poi.id || idx,
              name: poi.name,
              address: poi.address,
              latitude: poi.location_lat || (poi.location && parseFloat((`${poi.location}`).split(',')[1])),
              longitude: poi.location_lng || (poi.location && parseFloat((`${poi.location}`).split(',')[0])),
              type: poi.type,
              algo_hospital_category: poi.algo_hospital_category, // 只用算法字段
              algo_icon_type: poi.algo_icon_type, // 只用算法字段
              algo_display_order: poi.algo_display_order, // 只用算法字段
              // POI字段直接用typecode
              typecode: poi.typecode || '未知',
              childtype: poi.childtype || '未知',
              rating: 0,
              distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
              phone: poi.tel,
              website: poi.website,
              tags: poi.tags,
              intro: poi.intro,
              _sample: poi._sample,
            };
            console.log(`搜索时医院 ${hospital.name} 的typecode:`, hospital.typecode, 'childtype:', hospital.childtype);
            return hospital;
          });
          setHospitals(
            hospitals
              .filter(h => !!h.algo_hospital_category && typeof h.algo_hospital_category === 'string' && h.algo_hospital_category.trim() !== '' && h.algo_hospital_category !== '无' && h.algo_hospital_category !== 'null')
              .filter(h => !(h.tags && h.tags.includes('OptOut'))),
          );
          setIsSample(false);
          setLoading(false);
          return;
        }
      }
      // 兜底：高德API
      const amapKey = process.env.REACT_APP_AMAP_KEY;
      const poisRes = await fetch(`/api/amap/around?location=${newSearchLatLng.lng},${newSearchLatLng.lat}&radius=5000`);
      const poisData = await poisRes.json();
      if (poisData.status === '1' && poisData.pois && poisData.pois.length > 0) {
        const pois = poisData.pois.filter(poi => {
          const text = (poi.name || '') + (poi.type || '') + (poi.address || '');
          return !excludeKeywords.some(keyword => text.includes(keyword));
        });
        const hospitals = dedupHospitals(pois.map((poi, idx) => ({
          id: poi.id || idx,
          name: poi.name,
          address: poi.address,
          latitude: parseFloat(poi.location.split(',')[1]),
          longitude: parseFloat(poi.location.split(',')[0]),
          type: poi.type,
          algo_hospital_category: poi.hospital_category,
          algo_icon_type: poi.icon_type,
          rating: 0,
          distance: poi.distance ? parseFloat(poi.distance) / 1000 : 0,
          phone: poi.tel,
          website: '',
          tags: [poi.type],
          intro: poi.name + (poi.type ? `（${poi.type}）` : ''),
          _sample: false,
        })));
        setHospitals(
          hospitals
            .filter(h => h.algo_hospital_category && h.algo_hospital_category !== '')
            .filter(h => !(h.tags && h.tags.includes('OptOut'))),
        );
        setIsSample(false);
      } else {
        setHospitals([]);
        setIsSample(false);
        message.warning('高德API无结果，请尝试其他搜索方式或调整搜索范围。');
      }
    } catch (e) {
      console.error('搜索时获取医院数据失败:', e);
      setHospitals([]);
      setIsSample(false);
      message.error('医院数据获取失败，请稍后再试。');
    }
    setLoading(false);
  };

  // 优化地图容器为正方形自适应
  // 修改renderMap函数，只显示定位信息和医院列表，不渲染GoogleMapReact
  const renderMap = () => (
    <div className="home-map-container"
      style={{ width: '100%', aspectRatio: '1/1', maxWidth: 600, margin: '0 auto', background: '#f5f5f5', borderRadius: 16, boxShadow: '0 4px 16px rgba(0,0,0,0.08)', overflow: 'hidden', minHeight: 320 }}
    >
      <div id="map-container" ref={mapRef} style={{ width: '100%', height: '100%' }} />
    </div>
  );

  return (
    <div className="home-aero-bg">
      <div className="home-title">医院智能推荐平台</div>
      {showLocateTip && (
        <div style={{textAlign: 'center', margin: '12px 0'}}>
          <Button type="primary" onClick={handleLocate} style={{fontWeight: 'bold', fontSize: 16}}>
            📍点击授权定位，体验自动推荐
          </Button>
        </div>
      )}
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
          icon={
            <img
              src={locateGif}
              alt="定位"
              style={{width: 24, height: 24, verticalAlign: 'middle'}}
              onError={e => {
                e.target.onerror = null;
                e.target.style.display = 'none';
                // 兜底SVG小蓝人icon
                const svg = document.createElement('span');
                svg.innerHTML = '<svg width=\'24\' height=\'24\' viewBox=\'0 0 40 40\'><path d=\'M20 20 L20 5 A15 15 0 0 1 35 20 Z\' fill=\'rgba(30,144,255,0.25)\'/><circle cx=\'20\' cy=\'28\' r=\'6\' fill=\'#2196f3\' stroke=\'#1565c0\' stroke-width=\'2\'/><rect x=\'17\' y=\'28\' width=\'6\' height=\'8\' rx=\'3\' fill=\'#2196f3\'/><ellipse cx=\'20\' cy=\'24\' rx=\'4\' ry=\'3\' fill=\'#90caf9\'/></svg>';
                e.target.parentNode.appendChild(svg);
                console.warn('定位按钮GIF图片丢失，已自动切换为SVG小蓝人icon。请将locate_animated.gif放入public目录。');
              }}
            />
          }
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
            <div style={{textAlign:'center',color:'#888',marginTop:40,fontSize:18}}>
              当前区域无医院数据，请尝试扩大搜索范围或更换位置。
            </div>
          ) : (
            sortHospitals(filterHospitals(hospitals)).map(hospital => (
              <HospitalCard
                key={hospital.id}
                hospital={hospital}
                onClick={() => {
                  setSelectedHospital({ ...hospital }); // 保证所有字段传递
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
