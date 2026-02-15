from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()


class DetectedChange(Base):
    __tablename__ = "detected_changes"

    id = Column(Integer, primary_key=True)
    page_id = Column(Integer, nullable=False, index=True)
    old_price = Column(String(50), nullable=True)
    new_price = Column(String(50), nullable=True)
    change_percent = Column(Float, nullable=True)
    old_availability = Column(String(50), nullable=True)
    new_availability = Column(String(50), nullable=True)
    change_type = Column(String(20), nullable=False)  # price_change, availability_change, both
    detected_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    raw_data = Column(Text, nullable=True)

    def __repr__(self):
        return f"<DetectedChange(id={self.id}, page_id={self.page_id}, change_type={self.change_type})>"
