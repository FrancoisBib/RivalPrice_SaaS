import os
import sys
import re
import json
import hashlib
import logging
from datetime import datetime
from typing import Optional, List, Dict, Any

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
    def __init__(self, id: int, page_id: int, price: str, availability: str, scraped_at: datetime, raw_data: dict = None, page_type: str = None):
        self.id = id
        self.page_id = page_id
        self.price = price
        self.availability = availability
        self.scraped_at = scraped_at
        self.raw_data = raw_data or {}
        self.page_type = page_type or "pricing"
        self.content_hash = self._compute_hash()
    
    def compute_content_hash(self, content: dict) -> str:
        """Compute hash of specific content"""
        return hashlib.sha256(json.dumps(content, sort_keys=True).encode()).hexdigest()
    
    def _compute_hash(self) -> str:
        """Compute hash of entire snapshot content"""
        content = json.dumps(self.raw_data, sort_keys=True)
        return hashlib.sha256(content.encode()).hexdigest()


class IntelligentChangeDetectionService:
    def compute_content_hash(self, content: Dict[str, Any]) -> str:
        """Compute hash of specific content"""
        return hashlib.sha256(json.dumps(content, sort_keys=True).encode()).hexdigest()
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
                cleaned = cleaned.replace('.', '').replace(',', '.')
            else:
                cleaned = cleaned.replace(',', '')
        elif ',' in cleaned:
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
        conn = self.engine.connect()
        
        query = text("""
            SELECT s.id, s.monitored_page_id, s.price, s.availability, s.scraped_at, s.raw_data, m.page_type
            FROM snapshots s
            JOIN monitored_pages m ON s.monitored_page_id = m.id
            WHERE s.monitored_page_id = :page_id
            ORDER BY s.scraped_at DESC
            LIMIT :limit
        """)
        
        result = conn.execute(query, {"page_id": page_id, "limit": limit})
        snapshots = []
        
        for row in result:
            raw_data = {}
            if row[5]:
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
                raw_data=raw_data,
                page_type=row[6] or "pricing"
            ))
        
        conn.close()
        return snapshots
    
    def compute_global_hash(self, raw_data: Dict[str, Any]) -> str:
        """Compute global hash of all content"""
        content = {
            "price": raw_data.get("price", ""),
            "availability": raw_data.get("availability", ""),
            "features": raw_data.get("features", []),
            "pricing_blocks": raw_data.get("pricing_blocks", []),
            "text_content": raw_data.get("text_content", ""),
        }
        return hashlib.sha256(json.dumps(content, sort_keys=True).encode()).hexdigest()
    
    def detect_pricing_changes(self, old_data: Dict, new_data: Dict) -> Dict[str, Any]:
        """Detect changes in pricing blocks"""
        changes = {
            "has_change": False,
            "price_changed": False,
            "old_price": None,
            "new_price": None,
            "change_percent": None,
            "pricing_blocks_added": [],
            "pricing_blocks_removed": [],
            "pricing_blocks_changed": [],
        }
        
        old_price = self.parse_price(old_data.get("price", ""))
        new_price = self.parse_price(new_data.get("price", ""))
        
        if old_price is not None and new_price is not None and old_price != new_price:
            changes["has_change"] = True
            changes["price_changed"] = True
            changes["old_price"] = str(old_price)
            changes["new_price"] = str(new_price)
            changes["change_percent"] = self.calculate_price_change_percent(old_price, new_price)
        
        # Compare pricing blocks
        old_blocks = old_data.get("pricing_blocks", [])
        new_blocks = new_data.get("pricing_blocks", [])
        
        old_block_hashes = {self.compute_content_hash(b): b for b in old_blocks}
        new_block_hashes = {self.compute_content_hash(b): b for b in new_blocks}
        
        added = set(new_block_hashes.keys()) - set(old_block_hashes.keys())
        removed = set(old_block_hashes.keys()) - set(new_block_hashes.keys())
        
        changes["pricing_blocks_added"] = [new_block_hashes[h] for h in added]
        changes["pricing_blocks_removed"] = [old_block_hashes[h] for h in removed]
        
        if added or removed:
            changes["has_change"] = True
        
        return changes
    
    def detect_feature_changes(self, old_data: Dict, new_data: Dict) -> Dict[str, Any]:
        """Detect changes in features"""
        changes = {
            "has_change": False,
            "features_added": [],
            "features_removed": [],
        }
        
        old_features = set(old_data.get("features", []))
        new_features = set(new_data.get("features", []))
        
        added = new_features - old_features
        removed = old_features - new_features
        
        changes["features_added"] = list(added)
        changes["features_removed"] = list(removed)
        
        if added or removed:
            changes["has_change"] = True
        
        return changes
    
    def detect_messaging_changes(self, old_data: Dict, new_data: Dict) -> Dict[str, Any]:
        """Detect changes in visible text/messaging"""
        changes = {
            "has_change": False,
            "old_text": "",
            "new_text": "",
        }
        
        old_text = old_data.get("text_content", "") or old_data.get("title", "")
        new_text = new_data.get("text_content", "") or new_data.get("title", "")
        
        if old_text != new_text:
            changes["has_change"] = True
            changes["old_text"] = old_text
            changes["new_text"] = new_text
        
        return changes
    
    def detect_changes(self, page_id: int) -> Optional[DetectedChange]:
        """Detect changes between latest two snapshots for a page using intelligent comparison"""
        snapshots = self.get_latest_snapshots(page_id, limit=2)
        
        if len(snapshots) < 2:
            logger.info(f"Page {page_id}: Not enough snapshots to compare (found {len(snapshots)})")
            return None
        
        latest = snapshots[0]
        previous = snapshots[1]
        
        # Step 1: Compare global hashes
        old_hash = self.compute_global_hash(previous.raw_data)
        new_hash = self.compute_global_hash(latest.raw_data)
        
        if old_hash == new_hash:
            logger.info(f"Page {page_id}: No changes detected (hash identical)")
            return None
        
        logger.info(f"Page {page_id}: Hash changed from {old_hash[:16]}... to {new_hash[:16]}...")
        
        # Step 2: Detect specific changes based on page type
        change_types = []
        
        pricing_changes = self.detect_pricing_changes(previous.raw_data, latest.raw_data)
        if pricing_changes["price_changed"]:
            if pricing_changes["change_percent"] > 0:
                change_types.append("price_increase")
            else:
                change_types.append("price_decrease")
        
        if latest.page_type == "features":
            feature_changes = self.detect_feature_changes(previous.raw_data, latest.raw_data)
            if feature_changes["features_added"]:
                change_types.append("feature_added")
            if feature_changes["features_removed"]:
                change_types.append("feature_removed")
        
        messaging_changes = self.detect_messaging_changes(previous.raw_data, latest.raw_data)
        if messaging_changes["has_change"]:
            change_types.append("messaging_change")
        
        # If pricing blocks changed but no price, it's still a content change
        if pricing_changes["has_change"] and not pricing_changes["price_changed"]:
            change_types.append("content_change")
        
        if not change_types:
            change_types.append("content_change")
        
        # Create detected change record
        detected = DetectedChange(
            page_id=page_id,
            page_type=latest.page_type,
            old_price=pricing_changes.get("old_price"),
            new_price=pricing_changes.get("new_price"),
            change_percent=pricing_changes.get("change_percent"),
            old_availability=previous.availability,
            new_availability=latest.availability,
            old_features=json.dumps(previous.raw_data.get("features", [])),
            new_features=json.dumps(latest.raw_data.get("features", [])),
            features_added=json.dumps(feature_changes.get("features_added", [])) if latest.page_type == "features" else None,
            features_removed=json.dumps(feature_changes.get("features_removed", [])) if latest.page_type == "features" else None,
            old_text=messaging_changes.get("old_text"),
            new_text=messaging_changes.get("new_text"),
            change_type="_".join(change_types),
            old_hash=old_hash,
            new_hash=new_hash,
            detected_at=datetime.utcnow(),
            raw_data=json.dumps({
                "latest_snapshot_id": latest.id,
                "previous_snapshot_id": previous.id,
                "pricing_changes": pricing_changes,
                "feature_changes": feature_changes if latest.page_type == "features" else {},
                "messaging_changes": messaging_changes,
            })
        )
        
        # Save to database
        session = self.Session()
        session.add(detected)
        session.commit()
        
        logger.info(f"✅ Detected change for page {page_id}: {detected.change_type}")
        
        session.close()
        return detected
    
    def run_detection_for_all_pages(self):
        """Run change detection for all pages that have at least 2 snapshots"""
        conn = self.engine.connect()
        
        query = text("""
            SELECT DISTINCT monitored_page_id 
            FROM snapshots 
            GROUP BY monitored_page_id 
            HAVING COUNT(*) >= 2
        """)
        
        result = conn.execute(query)
        page_ids = [row[0] for row in result]
        conn.close()
        
        logger.info(f"🔍 Running intelligent change detection for {len(page_ids)} pages")
        
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
    service = IntelligentChangeDetectionService()
    
    # Run detection
    changes = service.run_detection_for_all_pages()
    print(f"Detected {changes} changes")
