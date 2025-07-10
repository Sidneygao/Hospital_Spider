import React from 'react';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import Home from './pages/Home';

export default function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <Home />
    </ConfigProvider>
  );
} 