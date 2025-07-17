import React from 'react';
import { Card } from 'antd';
import './HospitalCard.css';

export default function HospitalCard({ hospital, onClick }) {
  return (
    <Card className="hospital-card-aero" hoverable onClick={onClick}>
      {hospital._sample && (
        <div style={{color:'#d32f2f',fontWeight:'bold',marginBottom:8,fontSize:15}}>兜底SAMPLE数据</div>
      )}
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
