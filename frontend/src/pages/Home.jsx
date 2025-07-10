import React, { useState } from 'react';
import { message, Row, Col } from 'antd';
import SearchBar from '../components/SearchBar';
import HospitalMap from '../components/HospitalMap';
import HospitalList from '../components/HospitalList';
import { searchHospitals } from '../api/hospitals';

const defaultCenter = { lat: 13.7563, lng: 100.5018 };

export default function Home() {
  const [loading, setLoading] = useState(false);
  const [hospitals, setHospitals] = useState([]);
  const [center, setCenter] = useState(defaultCenter);

  // 地标搜索
  const handleSearch = async (landmark) => {
    setLoading(true);
    try {
      // 这里可接入地标转坐标API，暂用默认坐标
      const { data } = await searchHospitals({ lat: defaultCenter.lat, lng: defaultCenter.lng });
      setHospitals(data);
      setCenter(defaultCenter);
    } catch (e) {
      message.error('搜索失败');
    } finally {
      setLoading(false);
    }
  };

  // 地图 Marker 点击
  const handleMarkerClick = (hospital) => {
    message.info(`选中医院：${hospital.name}`);
  };

  // 列表点击
  const handleSelect = (hospital) => {
    setCenter({ lat: hospital.latitude, lng: hospital.longitude });
  };

  return (
    <div style={{ padding: 24 }}>
      <h2>医院智能推荐 - 热点图搜索</h2>
      <SearchBar onSearch={handleSearch} loading={loading} />
      <Row gutter={24}>
        <Col xs={24} md={16}>
          <HospitalMap hospitals={hospitals} center={center} onMarkerClick={handleMarkerClick} />
        </Col>
        <Col xs={24} md={8}>
          <HospitalList hospitals={hospitals} onSelect={handleSelect} />
        </Col>
      </Row>
    </div>
  );
} 