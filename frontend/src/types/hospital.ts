export interface Hospital {
  id: number;
  name: string;
  address: string;
  phone?: string;
  rating: number;
  type?: string;
  level?: string;
  specialties?: string[];
  description?: string;
  latitude?: number;
  longitude?: number;
  distance?: number;
  website?: string;
  operatingHours?: string;
  emergencyContact?: string;
  facilities?: string[];
  insurance?: string[];
  languages?: string[];
  accessibility?: string[];
  parking?: boolean;
  wheelchair?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface SearchParams {
  lat: number;
  lng: number;
  radius: number;
  limit: number;
  type?: string;
  level?: string;
  specialty?: string;
}

export interface SearchResult {
  hospitals: Hospital[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}

export interface HospitalDetail extends Hospital {
  reviews: Review[];
  photos: string[];
  services: Service[];
  doctors: Doctor[];
  departments: Department[];
}

export interface Review {
  id: number;
  userId: number;
  userName: string;
  rating: number;
  comment: string;
  createdAt: string;
  helpful: number;
}

export interface Service {
  id: number;
  name: string;
  description: string;
  price?: number;
  available: boolean;
}

export interface Doctor {
  id: number;
  name: string;
  specialty: string;
  experience: number;
  rating: number;
  photo?: string;
  available: boolean;
}

export interface Department {
  id: number;
  name: string;
  description: string;
  headDoctor?: string;
  phone?: string;
  location?: string;
}
