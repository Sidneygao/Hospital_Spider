from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import List, Optional
from sqlalchemy import create_engine, Column, Integer, String, Float, Text, DateTime, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
import datetime

DATABASE_URL = "sqlite:///./hospital_spider.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# 数据模型
class Hospital(Base):
    __tablename__ = "hospitals"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    address = Column(String, nullable=False)
    latitude = Column(Float, nullable=False)
    longitude = Column(Float, nullable=False)
    phone = Column(String)
    hospital_type = Column(String)
    main_departments = Column(String)
    business_hours = Column(String)
    qualifications = Column(String)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.datetime.utcnow)
    ratings = relationship("Rating", back_populates="hospital")
    reviews = relationship("Review", back_populates="hospital")

class Rating(Base):
    __tablename__ = "ratings"
    id = Column(Integer, primary_key=True, index=True)
    hospital_id = Column(Integer, ForeignKey("hospitals.id"))
    source = Column(String, nullable=False)
    rating_value = Column(Float, nullable=False)
    confidence = Column(Float, default=0.5)
    rating_date = Column(DateTime)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    hospital = relationship("Hospital", back_populates="ratings")

class Review(Base):
    __tablename__ = "reviews"
    id = Column(Integer, primary_key=True, index=True)
    hospital_id = Column(Integer, ForeignKey("hospitals.id"))
    source = Column(String, nullable=False)
    user_name = Column(String)
    rating = Column(Float)
    review_text = Column(Text)
    review_date = Column(DateTime)
    sentiment_score = Column(Float)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)
    hospital = relationship("Hospital", back_populates="reviews")

# Pydantic模型
class HospitalOut(BaseModel):
    id: int
    name: str
    address: str
    latitude: float
    longitude: float
    phone: Optional[str]
    hospital_type: Optional[str]
    main_departments: Optional[str]
    business_hours: Optional[str]
    qualifications: Optional[str]
    class Config:
        orm_mode = True

# FastAPI 实例
app = FastAPI()

# 允许跨域
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 初始化数据库
Base.metadata.create_all(bind=engine)

# API 示例：医院列表查询
@app.get("/api/hospitals", response_model=List[HospitalOut])
def get_hospitals(limit: int = 10, skip: int = 0):
    db = SessionLocal()
    hospitals = db.query(Hospital).offset(skip).limit(limit).all()
    db.close()
    return hospitals

# API 示例：根据地理位置检索医院（简化版）
@app.get("/api/hospitals/search", response_model=List[HospitalOut])
def search_hospitals(lat: float = Query(...), lng: float = Query(...), radius: float = 10.0, limit: int = 10):
    db = SessionLocal()
    # 简化：返回全部医院，实际应按距离过滤
    hospitals = db.query(Hospital).limit(limit).all()
    db.close()
    return hospitals

# API 示例：获取单个医院详情
@app.get("/api/hospitals/{hospital_id}", response_model=HospitalOut)
def get_hospital(hospital_id: int):
    db = SessionLocal()
    hospital = db.query(Hospital).filter(Hospital.id == hospital_id).first()
    db.close()
    if not hospital:
        raise HTTPException(status_code=404, detail="Hospital not found")
    return hospital 