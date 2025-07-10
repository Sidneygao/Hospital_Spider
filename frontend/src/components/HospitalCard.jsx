import React from 'react';
import { Card } from 'antd';
import './HospitalCard.css';

export default function HospitalCard({ hospital }) {
  return (
    <Card
      className="hospital-card-aero"
      style={{
        background: 'rgba(180,220,255,0.7)',
        borderRadius: 18,
        boxShadow: '0 4px 24px 0 rgba(0, 40, 120, 0.12)',
        marginBottom: 24,
        border: 'none',
      }}
      hoverable
    >
      <div className="card-title">{hospital.name}</div>
      <div className="card-info">{hospital.address}</div>
      <div className="card-meta">
        <span>类型：{hospital.type}</span>
        <span>评分：{hospital.rating}</span>
        <span>距离：{hospital.distance}km</span>
      </div>
    </Card>
  );
} 