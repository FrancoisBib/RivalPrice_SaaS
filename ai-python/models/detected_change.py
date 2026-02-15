from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()


class DetectedChange(Base):
    __tablename__ = "detected_changes"

    id = Column(Integer, primary_key=True)
    page_id = Column(Integer, nullable=False, index=True)
    page_type = Column(String(20), nullable=True)  # pricing, features
    
    # Price changes
    old_price = Column(String(50), nullable=True)
    new_price = Column(String(50), nullable=True)
    change_percent = Column(Float, nullable=True)
    
    # Availability changes
    old_availability = Column(String(50), nullable=True)
    new_availability = Column(String(50), nullable=True)
    
    # Feature changes
    old_features = Column(Text, nullable=True)  # JSON array of features
    new_features = Column(Text, nullable=True)
    features_added = Column(Text, nullable=True)  # JSON
    features_removed = Column(Text, nullable=True)  # JSON
    
    # Messaging changes
    old_text = Column(Text, nullable=True)
    new_text = Column(Text, nullable=True)
    
    # Change type: price_increase, price_decrease, availability_change, 
    # feature_added, feature_removed, messaging_change, content_change
    change_type = Column(String(50), nullable=False)
    
    # Hash for quick comparison
    old_hash = Column(String(64), nullable=True)
    new_hash = Column(String(64), nullable=True)
    
    detected_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    raw_data = Column(Text, nullable=True)

    def __repr__(self):
        return f"<DetectedChange(id={self.id}, page_id={self.page_id}, change_type={self.change_type})>"
