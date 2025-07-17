import React from 'react';
import './HospitalDetailModal.css';

export default function HospitalDetailModal({ visible, hospital, onClose }) {
  if (!visible || !hospital) return null;
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose}>×</button>
        <div className="modal-title">{hospital.name}</div>
        <div className="modal-meta">
          <span>类型：{hospital.type}</span>
          <span>评分：{hospital.rating}</span>
        </div>
        <div className="modal-info">{hospital.address}</div>
        <div className="modal-tags">
          {hospital.tags && hospital.tags.map(tag => (
            <span className="modal-tag" key={tag}>{tag}</span>
          ))}
        </div>
        <div className="modal-intro">{hospital.intro}</div>
        <div className="modal-contact">
          <span>电话：<a href={`tel:${hospital.phone}`}>{hospital.phone}</a></span>
          <span>官网：<a href={hospital.website} target="_blank" rel="noopener noreferrer">{hospital.website}</a></span>
        </div>
        <div className="modal-map">
          <img
            src={`/api/google/staticmap?center=${hospital.latitude},${hospital.longitude}&zoom=15&size=400x200&markers=color:blue|label:P|${hospital.latitude},${hospital.longitude}`}
            alt="医院地图"
            style={{width:'100%',borderRadius:8,marginTop:12}}
          />
        </div>
      </div>
    </div>
  );
} 