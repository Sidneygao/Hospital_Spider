import React from 'react';
import { Card } from 'antd';
import './HospitalCard.css';

const getHospitalLevel = (type = '') => {
  const match = type.match(/(三级甲等|三甲|二级甲等|三级|二级|一级)/);
  return match ? match[0] : '';
};

// 根据algo_icon_type获取对应的图标
const getHospitalIcon = (iconType) => {
  switch (iconType) {
  case 'icon_tier3_hospital_bold':
    return (
      <svg width="24" height="24" viewBox="0 0 24 24" style={{ marginRight: 8 }}>
        <circle cx="12" cy="12" r="10" fill="#fff" stroke="#d32f2f" strokeWidth="2"/>
        <rect x="9" y="5" width="6" height="14" rx="2" fill="#d32f2f"/>
        <rect x="5" y="9" width="14" height="6" rx="2" fill="#d32f2f"/>
      </svg>
    );
  case 'icon_general_hospital_bold':
    return (
      <svg width="22" height="22" viewBox="0 0 22 22" style={{ marginRight: 8 }}>
        <circle cx="11" cy="11" r="9" fill="#fff" stroke="#d32f2f" strokeWidth="1.8"/>
        <rect x="8.5" y="4.5" width="5" height="13" rx="1.8" fill="#d32f2f"/>
        <rect x="4.5" y="8.5" width="13" height="5" rx="1.8" fill="#d32f2f"/>
      </svg>
    );
  case 'icon_small_red_cross_bold':
    return (
      <svg width="20" height="20" viewBox="0 0 20 20" style={{ marginRight: 8 }}>
        <circle cx="10" cy="10" r="8" fill="#fff" stroke="#d32f2f" strokeWidth="1.5"/>
        <rect x="7.5" y="3.5" width="5" height="13" rx="1.5" fill="#d32f2f"/>
        <rect x="3.5" y="7.5" width="13" height="5" rx="1.5" fill="#d32f2f"/>
      </svg>
    );
  case 'icon_small_red_cross_normal':
    return (
      <svg width="16" height="16" viewBox="0 0 16 16" style={{ marginRight: 8 }}>
        <circle cx="8" cy="8" r="6" fill="#fff" stroke="#d32f2f" strokeWidth="0.8"/>
        <rect x="6" y="3" width="4" height="10" rx="0.8" fill="#d32f2f"/>
        <rect x="3" y="6" width="10" height="4" rx="0.8" fill="#d32f2f"/>
      </svg>
    );
  case 'icon_tooth':
    return (
      <svg width="18" height="18" viewBox="0 0 18 18" style={{ marginRight: 8 }}>
        <path d="M3,9 Q2,5 5.5,4 Q8,3 11.5,4 Q15,5 13.5,9 Q12,15 8,15 Q4,15 3,9 Z" stroke="#1e90ff" strokeWidth="1.5" fill="none"/>
        <path d="M6.5,13 Q8,11 11.5,13" stroke="#1e90ff" strokeWidth="0.8" fill="none"/>
      </svg>
    );
  case 'icon_er':
    return (
      <svg width="18" height="18" viewBox="0 0 18 18" style={{ marginRight: 8 }}>
        <rect x="2" y="2" width="14" height="14" rx="2" fill="#ff6b35" stroke="#d32f2f" strokeWidth="1"/>
        <text x="9" y="12" textAnchor="middle" fill="white" fontSize="10" fontWeight="bold">ER</text>
      </svg>
    );
  default:
    return (
      <svg width="16" height="16" viewBox="0 0 16 16" style={{ marginRight: 8 }}>
        <circle cx="8" cy="8" r="6" fill="#fff" stroke="#666" strokeWidth="1"/>
        <rect x="6" y="3" width="4" height="10" rx="1" fill="#666"/>
        <rect x="3" y="6" width="10" height="4" rx="1" fill="#666"/>
      </svg>
    );
  }
};

export default function HospitalCard({ hospital, onClick }) {
  const level = getHospitalLevel(hospital.type);
  const icon = getHospitalIcon(hospital.algo_icon_type);

  return (
    <Card
      className="hospital-card-aero"
      hoverable
      onClick={onClick}
      data-hospital-id={hospital.id}
    >
      {hospital._sample && (
        <div style={{color:'#d32f2f',fontWeight:'bold',marginBottom:8,fontSize:15}}>兜底SAMPLE数据</div>
      )}
      <div className="hospital-card-header">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          {icon}
          <span className="hospital-card-title">{hospital.name}</span>
        </div>
        {level && <span className="hospital-card-level" style={{ color: '#1976d2', marginLeft: 8, fontWeight: 500, fontSize: 16 }}>{level}</span>}
      </div>
      {/* 新增：显示计算出的医院类别和排序 */}
      <div style={{ color: '#888', fontSize: 13, marginBottom: 2 }}>
        类别：{hospital.algo_hospital_category || '无'}（排序：{hospital.algo_display_order ?? '无'}）
      </div>
      <div className="card-info">{hospital.address}</div>
      <div className="card-meta">
        <span>评分：{hospital.rating}</span>
        <span>距离：{hospital.distance}km</span>
      </div>
    </Card>
  );
}
