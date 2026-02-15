import os
import sys
import re
import json
import logging
from datetime import datetime
from typing import Optional, List, Tuple

from sqlalchemy import create_engine, Column, Integer, String, Float, DateTime, Text, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, Session

# Add parent to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from models.detected_change import DetectedChange

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

Base = declarative_base()


class Snapshot:
    """Represents a snapshot from the scraped data"""
    def __init__(self, id: int, page_id: int, price: str, availability: str, scraped_at: datetime, raw_data: dict = None):
        self.id = id
        self.page_id = page_id
        self.price = price
        self.availability = availability
        self.scraped_at = scraped_at
        self.raw_data = raw_data or {}


class ChangeDetectionService:
    def __init__(self, db_url: str = None):
        if db_url is None:
            db_url = os.getenv(
                "DATABASE_URL",
                "postgresql://rivalprice:rivalprice_secret@localhost:5432/rivalprice"
            )
        
        self.engine = create_engine(db_url)
        Base.metadata.create_all(self.engine)
        self.Session = sessionmaker(bind=self.engine)
    
    def parse_price(self, price_str: str) -> Optional[float]:
        """Extract numeric value from price string"""
        if not price_str:
            return None
        
        # Remove currency symbols and whitespace
        cleaned = re.sub(r'[^\d.,]', '', price_str)
        # Handle European format (1.234,56) vs US (1,234.56)
        if ',' in cleaned and '.' in cleaned:
            if cleaned.rfind(',') > cleaned.rfind('.'):
                # European: 1.234,56 -> 1234.56
                cleaned = cleaned.replace('.', '').replace(',', '.')
            else:
                # US: 1,234.56 -> 1234.56
                cleaned = cleaned.replace(',', '')
        elif ',' in cleaned:
            # Could be European decimal or US thousands
            parts = cleaned.split(',')
            if len(parts) == 2 and len(parts[1]) == 2:
                cleaned = cleaned.replace(',', '.')
            else:
                cleaned = cleaned.replace(',', '')
        
        try:
            return float(cleaned)
        except ValueError:
            return None
    
    def calculate_price_change_percent(self, old_price: float, new_price: float) -> float:
        """Calculate percentage change in price"""
        if old_price == 0:
            return 100.0 if new_price > 0 else 0.0
        return ((new_price - old_price) / old_price) * 100
    
    def get_latest_snapshots(self, page_id: int, limit: int = 2) -> List[Snapshot]:
        """Get the latest snapshots for a page from PostgreSQL"""
        # Connect directly to the snapshots table
        conn = self.engine.connect()
        
        query = text(f"""
            SELECT id, monitored_page_id, price, availability, scraped_at, raw_data
            FROM snapshots
            WHERE monitored_page_id = :page_id
            ORDER BY scraped_at DESC
            LIMIT :limit
        """)
        
        result = conn.execute(query, {"page_id": page_id, "limit": limit})
        snapshots = []
        
        for row in result:
            raw_data = {}
            if row[5]:
                # raw_data is JSONB
                if hasattr(row[5], 'keys'):
                    raw_data = dict(row[5])
                else:
                    try:
                        raw_data = json.loads(str(row[5]))
                    except:
                        pass
            
            snapshots.append(Snapshot(
                id=row[0],
                page_id=row[1],
                price=row[2] or "",
                availability=row[3] or "",
                scraped_at=row[4],
                raw_data=raw_data
            ))
        
        conn.close()
        return snapshots
    
    def detect_changes(self, page_id: int) -> Optional[DetectedChange]:
        """Detect changes between latest two snapshots for a page"""
        snapshots = self.get_latest_snapshots(page_id, limit=2)
        
        if len(snapshots) < 2:
            logger.info(f"Page {page_id}: Not enough snapshots to compare (found {len(snapshots)})")
            return None
        
        latest = snapshots[0]
        previous = snapshots[1]
        
        old_price = self.parse_price(previous.price)
        new_price = self.parse_price(latest.price)
        
        # Detect price change
        price_changed = False
        change_percent = None
        if old_price is not None and new_price is not None and old_price != new_price:
            price_changed = True
            change_percent = self.calculate_price_change_percent(old_price, new_price)
        
        # Detect availability change
        availability_changed = previous.availability != latest.availability
        
        if not price_changed and not availability_changed:
            logger.info(f"Page {page_id}: No changes detected")
            return None
        
        # Determine change type
        change_types = []
        if price_changed:
            change_types.append("price_change")
        if availability_changed:
            change_types.append("availability_change")
        change_type = "_".join(change_types)
        
        # Create detected change record
        detected = DetectedChange(
            page_id=page_id,
            old_price=previous.price,
            new_price=latest.price,
            change_percent=change_percent,
            old_availability=previous.availability,
            new_availability=latest.availability,
            change_type=change_type,
            detected_at=datetime.utcnow(),
            raw_data=json.dumps({
                "latest_snapshot_id": latest.id,
                "previous_snapshot_id": previous.id,
                "latest_raw": latest.raw_data,
                "previous_raw": previous.raw_data
            })
        )
        
        # Save to database
        session = self.Session()
        session.add(detected)
        session.commit()
        
        logger.info(f"✅ Detected change for page {page_id}: {change_type}, old_price={previous.price}, new_price={latest.price}")
        
        session.close()
        return detected
    
    def run_detection_for_all_pages(self):
        """Run change detection for all pages that have at least 2 snapshots"""
        conn = self.engine.connect()
        
        # Find pages with at least 2 snapshots
        query = text("""
            SELECT DISTINCT monitored_page_id 
            FROM snapshots 
            GROUP BY monitored_page_id 
            HAVING COUNT(*) >= 2
        """)
        
        result = conn.execute(query)
        page_ids = [row[0] for row in result]
        conn.close()
        
        logger.info(f"🔍 Running change detection for {len(page_ids)} pages")
        
        changes_detected = 0
        for page_id in page_ids:
            try:
                change = self.detect_changes(page_id)
                if change:
                    changes_detected += 1
            except Exception as e:
                logger.error(f"Error detecting changes for page {page_id}: {e}")
        
        logger.info(f"🎉 Change detection complete: {changes_detected} changes detected")
        return changes_detected


if __name__ == "__main__":
    service = ChangeDetectionService()
    
    # Run detection
    changes = service.run_detection_for_all_pages()
    print(f"Detected {changes} changes")
