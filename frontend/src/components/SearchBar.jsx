import React, { useState } from 'react';
import { Input, Button } from 'antd';

export default function SearchBar({ onSearch, loading }) {
  const [value, setValue] = useState('');

  const handleSearch = () => {
    if (value.trim()) {
      onSearch(value.trim());
    }
  };

  return (
    <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
      <Input
        placeholder="请输入地标或地址，如曼谷四面佛"
        value={value}
        onChange={e => setValue(e.target.value)}
        onPressEnter={handleSearch}
        disabled={loading}
        style={{ maxWidth: 400 }}
      />
      <Button type="primary" onClick={handleSearch} loading={loading}>
        搜索
      </Button>
    </div>
  );
} 