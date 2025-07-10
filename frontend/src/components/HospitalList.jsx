import React from 'react';
import { Card, List, Rate } from 'antd';

export default function HospitalList({ hospitals = [], onSelect }) {
  return (
    <List
      grid={{ gutter: 16, column: 1 }}
      dataSource={hospitals}
      renderItem={item => (
        <List.Item>
          <Card
            title={item.name}
            extra={<span>{item.hospital_type}</span>}
            onClick={() => onSelect && onSelect(item)}
            hoverable
            style={{ cursor: 'pointer' }}
          >
            <div>地址：{item.address}</div>
            <div>电话：{item.phone || '无'}</div>
            <div>评分：<Rate disabled value={item.rating || 0} allowHalf /> ({item.rating?.toFixed(1) || '暂无'})</div>
            <div>距离：{item.distance ? item.distance.toFixed(2) + ' km' : '未知'}</div>
          </Card>
        </List.Item>
      )}
    />
  );
} 