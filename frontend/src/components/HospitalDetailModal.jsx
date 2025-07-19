import React from 'react';
import './HospitalDetailModal.css';

export default function HospitalDetailModal({ visible, hospital, onClose }) {
  if (!visible || !hospital) return null;

  // 调试：输出hospital对象结构
  console.log('hospital对象调试', hospital);

  // 在文件顶部或组件内添加渲染函数
  function renderChildType(childtype) {
    if (childtype === undefined || childtype === null) {
      return 'undefined';
    }
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
    if (typecode === undefined || typecode === null) {
      return 'undefined';
    }
    if (typeof typecode === 'string') {
      return typecode === '' ? '""' : typecode;
    }
    if (typeof typecode === 'number') {
      return typecode.toString();
    }
    return String(typecode);
  }

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
        {/* 暂时注释掉地图，专注于POI/ChildType问题 */}
        {/* <div className="modal-map">
          <img
            src={`/api/google/staticmap?center=${hospital.latitude},${hospital.longitude}&zoom=15&size=400x200&markers=color:blue|label:P|${hospital.latitude},${hospital.longitude}`}
            alt="医院地图"
            style={{width:'100%',borderRadius:8,marginTop:12}}
          />
        </div> */}
        {/* 医院详情底部字段显示区，替换原有POI/ChildType显示 */}
        <div style={{ color: '#aaa', fontSize: 12, marginTop: 10, textAlign: 'left' }}>
          <div>
            <span>POI: {renderTypeCode(hospital.typecode)}</span>
            <span>ChildType: {renderChildType(hospital.childtype)}</span>
            <span>计算类别: {hospital.algo_hospital_category}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
