import { Hospital, HospitalDetail } from '../types/hospital';

const API_BASE_URL =
  process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

// 通用请求函数
const request = async <T>(url: string, options?: RequestInit): Promise<T> => {
  try {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
};

// 获取所有医院
export const getHospitals = async (): Promise<Hospital[]> => {
  return request<Hospital[]>('/hospitals');
};

// 搜索医院
export const searchHospitals = async (
  lat: number,
  lng: number,
  radius = 10,
  limit = 20,
): Promise<Hospital[]> => {
  const params = new URLSearchParams({
    lat: lat.toString(),
    lng: lng.toString(),
    radius: radius.toString(),
    limit: limit.toString(),
  });

  return request<Hospital[]>(`/hospitals/search?${params}`);
};

// 获取医院详情
export const getHospitalDetail = async (
  id: number,
): Promise<HospitalDetail> => {
  return request<HospitalDetail>(`/hospitals/${id}`);
};

// 提交医院反馈
export const submitHospitalFeedback = async (
  hospitalId: number,
  feedback: {
    rating: number;
    comment: string;
    userId?: number;
  },
): Promise<{ success: boolean; message: string }> => {
  return request<{ success: boolean; message: string }>(
    `/hospitals/${hospitalId}/feedback`,
    {
      method: 'POST',
      body: JSON.stringify(feedback),
    },
  );
};

// 地理编码服务（模拟）
export const geocodeAddress = async (
  address: string,
): Promise<{
  lat: number;
  lng: number;
  formattedAddress: string;
}> => {
  // 这里应该调用真实的地理编码API
  // 暂时返回模拟数据
  const mockGeocoding = {
    北京市朝阳区建国门外大街1号: {
      lat: 39.9042,
      lng: 116.4074,
      formattedAddress: '北京市朝阳区建国门外大街1号',
    },
    上海市浦东新区陆家嘴环路1000号: {
      lat: 31.2304,
      lng: 121.4737,
      formattedAddress: '上海市浦东新区陆家嘴环路1000号',
    },
    广州市天河区珠江新城花城大道85号: {
      lat: 23.1291,
      lng: 113.2644,
      formattedAddress: '广州市天河区珠江新城花城大道85号',
    },
    深圳市南山区深南大道10000号: {
      lat: 22.5431,
      lng: 114.0579,
      formattedAddress: '深圳市南山区深南大道10000号',
    },
    杭州市西湖区文三路259号: {
      lat: 30.2741,
      lng: 120.1551,
      formattedAddress: '杭州市西湖区文三路259号',
    },
    成都市锦江区红星路三段1号: {
      lat: 30.5728,
      lng: 104.0668,
      formattedAddress: '成都市锦江区红星路三段1号',
    },
    武汉市江汉区解放大道634号: {
      lat: 30.5928,
      lng: 114.3055,
      formattedAddress: '武汉市江汉区解放大道634号',
    },
    西安市雁塔区高新路25号: {
      lat: 34.3416,
      lng: 108.9398,
      formattedAddress: '西安市雁塔区高新路25号',
    },
  };

  // 模拟网络延迟
  await new Promise((resolve) => setTimeout(resolve, 500));

  const result = mockGeocoding[address as keyof typeof mockGeocoding];
  if (result) {
    return result;
  }

  // 如果没有找到匹配的地址，返回默认位置
  return {
    lat: 13.7563,
    lng: 100.5018,
    formattedAddress: address,
  };
};

// 反向地理编码服务（模拟）
export const reverseGeocode = async (
  lat: number,
  lng: number,
): Promise<{
  address: string;
  city: string;
  country: string;
}> => {
  // 这里应该调用真实的反向地理编码API
  // 暂时返回模拟数据
  await new Promise((resolve) => setTimeout(resolve, 300));

  return {
    address: `位置 (${lat.toFixed(4)}, ${lng.toFixed(4)})`,
    city: '未知城市',
    country: '未知国家',
  };
};

// 获取医院类型列表
export const getHospitalTypes = async (): Promise<string[]> => {
  return ['综合医院', '专科医院', '中医医院', '妇幼保健院', '康复医院', '诊所'];
};

// 获取医院等级列表
export const getHospitalLevels = async (): Promise<string[]> => {
  return [
    '三级甲等',
    '三级乙等',
    '二级甲等',
    '二级乙等',
    '一级甲等',
    '一级乙等',
  ];
};

// 获取专科列表
export const getSpecialties = async (): Promise<string[]> => {
  return [
    '内科',
    '外科',
    '妇产科',
    '儿科',
    '眼科',
    '耳鼻喉科',
    '口腔科',
    '皮肤科',
    '神经科',
    '精神科',
    '传染科',
    '肿瘤科',
    '急诊科',
    '康复科',
    '中医科',
    '骨科',
    '泌尿科',
    '心血管科',
    '呼吸科',
    '消化科',
    '内分泌科',
    '血液科',
  ];
};

// 缓存管理
class CacheManager {
  private cache = new Map<
    string,
    { data: any; timestamp: number; ttl: number }
  >();

  set(key: string, data: any, ttl: number = 5 * 60 * 1000): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    });
  }

  get(key: string): any | null {
    const item = this.cache.get(key);
    if (!item) return null;

    if (Date.now() - item.timestamp > item.ttl) {
      this.cache.delete(key);
      return null;
    }

    return item.data;
  }

  clear(): void {
    this.cache.clear();
  }
}

export const cacheManager = new CacheManager();

// 带缓存的API请求
export const cachedRequest = async <T>(
  key: string,
  requestFn: () => Promise<T>,
  ttl: number = 5 * 60 * 1000,
): Promise<T> => {
  const cached = cacheManager.get(key);
  if (cached) {
    return cached;
  }

  const data = await requestFn();
  cacheManager.set(key, data, ttl);
  return data;
};
