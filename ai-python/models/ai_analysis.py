from datetime import datetime
from sqlalchemy import Column, Integer, String, DateTime, Text, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

Base = declarative_base()


class AIAnalysis(Base):
    __tablename__ = "ai_analysis"

    id = Column(Integer, primary_key=True)
    change_id = Column(Integer, nullable=False, index=True)
    
    # LLM generated content
    summary = Column(Text, nullable=False)
    recommendation = Column(String(500), nullable=True)
    
    # Change details for context
    change_type = Column(String(50), nullable=True)
    page_type = Column(String(20), nullable=True)
    old_price = Column(String(50), nullable=True)
    new_price = Column(String(50), nullable=True)
    change_percent = Column(String(20), nullable=True)
    
    # AI model used
    model = Column(String(50), nullable=True)
    
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    def __repr__(self):
        return f"<AIAnalysis(id={self.id}, change_id={self.change_id}, summary={self.summary[:50]}...)>"
