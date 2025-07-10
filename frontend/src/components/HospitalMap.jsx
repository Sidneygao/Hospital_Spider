import React from 'react';
import { GoogleMap, Marker, useJsApiLoader } from '@react-google-maps/api';

const containerStyle = {
  width: '100%',
  height: '400px',
};

const defaultCenter = {
  lat: 13.7563, // 曼谷默认坐标
  lng: 100.5018,
};

export default function HospitalMap({ hospitals = [], center = defaultCenter, onMarkerClick }) {
  const { isLoaded } = useJsApiLoader({
    googleMapsApiKey: process.env.REACT_APP_GOOGLE_MAPS_API_KEY || '',
  });

  return isLoaded ? (
    <GoogleMap
      mapContainerStyle={containerStyle}
      center={center}
      zoom={12}
    >
      {hospitals.map(h => (
        <Marker
          key={h.id}
          position={{ lat: h.latitude, lng: h.longitude }}
          title={h.name}
          onClick={() => onMarkerClick && onMarkerClick(h)}
        />
      ))}
    </GoogleMap>
  ) : (
    <div style={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      地图加载中...
    </div>
  );
} 