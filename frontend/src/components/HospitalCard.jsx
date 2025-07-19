import React from 'react';
import { Card } from 'antd';
import './HospitalCard.css';

const getHospitalLevel = (type = '') => {
  const match = type.match(/(三级甲等|三甲|二级甲等|三级|二级|一级)/);
  return match ? match[0] : '';
};

export default function HospitalCard({ hospital, onClick }) {
  const level = getHospitalLevel(hospital.type);
  return (
    <Card className="hospital-card-aero" hoverable onClick={onClick}>
      {hospital._sample && (
        <div style={{color:'#d32f2f',fontWeight:'bold',marginBottom:8,fontSize:15}}>兜底SAMPLE数据</div>
      )}
      <div className="hospital-card-header">
        <span className="hospital-card-title">{hospital.name}</span>
        {level && <span className="hospital-card-level" style={{ color: '#1976d2', marginLeft: 8, fontWeight: 500, fontSize: 16 }}>{level}</span>}
      </div>
      {/* 新增：显示计算出的医院类别和排序 */}
      <div style={{ color: '#888', fontSize: 13, marginBottom: 2 }}>
        类别：{hospital.algo_hospital_category || '无'}（排序：{hospital.algo_display_order ?? '无'}）
      </div>
      {/* 如需渲染ICON，可用algo_icon_type字段 */}
      {/* <img src={getIconUrl(hospital.algo_icon_type)} alt="icon" /> */}
      <div className="card-info">{hospital.address}</div>
      <div className="card-meta">
        <span>评分：{hospital.rating}</span>
        <span>距离：{hospital.distance}km</span>
      </div>
    </Card>
  );
}
