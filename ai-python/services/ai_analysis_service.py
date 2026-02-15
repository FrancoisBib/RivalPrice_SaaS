import os
import sys
import json
import logging
from datetime import datetime
from typing import Optional, Dict, Any

from sqlalchemy import create_engine, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

# Add parent to path for imports
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

try:
    from openai import OpenAI
    OPENAI_AVAILABLE = True
except ImportError:
    OPENAI_AVAILABLE = False
    logging.warning("OpenAI not installed. AI analysis will use rule-based fallback.")

from models.ai_analysis import AIAnalysis

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

Base = declarative_base()


class AIAnalysisService:
    def __init__(self, db_url: str = None, openai_api_key: str = None):
        if db_url is None:
            db_url = os.getenv(
                "DATABASE_URL",
                "postgresql://rivalprice:rivalprice_secret@localhost:5432/rivalprice"
            )
        
        self.engine = create_engine(db_url)
        Base.metadata.create_all(self.engine)
        self.Session = sessionmaker(bind=self.engine)
        
        # Initialize OpenAI client
        self.openai_api_key = openai_api_key or os.getenv("OPENAI_API_KEY")
        if OPENAI_AVAILABLE and self.openai_api_key:
            self.client = OpenAI(api_key=self.openai_api_key)
        else:
            self.client = None
    
    def get_unanalyzed_changes(self) -> list:
        """Get detected changes that haven't been analyzed yet"""
        conn = self.engine.connect()
        
        query = text("""
            SELECT dc.id, dc.page_id, dc.change_type, dc.page_type, 
                   dc.old_price, dc.new_price, dc.change_percent,
                   dc.old_features, dc.new_features,
                   dc.old_text, dc.new_text,
                   dc.raw_data
            FROM detected_changes dc
            LEFT JOIN ai_analysis aa ON dc.id = aa.change_id
            WHERE aa.id IS NULL
            ORDER BY dc.detected_at DESC
            LIMIT 10
        """)
        
        result = conn.execute(query)
        changes = []
        
        for row in result:
            changes.append({
                "change_id": row[0],
                "page_id": row[1],
                "change_type": row[2],
                "page_type": row[3],
                "old_price": row[4],
                "new_price": row[5],
                "change_percent": row[6],
                "old_features": row[7],
                "new_features": row[8],
                "old_text": row[9],
                "new_text": row[10],
                "raw_data": row[11]
            })
        
        conn.close()
        return changes
    
    def generate_summary_with_llm(self, change_data: Dict[str, Any]) -> tuple:
        """Use OpenAI LLM to generate summary and recommendation"""
        if not self.client:
            return self.generate_rule_based_summary(change_data)
        
        try:
            # Build prompt
            change_type = change_data.get("change_type", "unknown")
            page_type = change_data.get("page_type", "pricing")
            
            prompt = f"""Analyze this competitor pricing change and provide a summary and recommendation.

Change Type: {change_type}
Page Type: {page_type}
Old Price: {change_data.get('old_price', 'N/A')}
New Price: {change_data.get('new_price', 'N/A')}
Change: {change_data.get('change_percent', 'N/A')}%

Old Features: {change_data.get('old_features', 'N/A')}
New Features: {change_data.get('new_features', 'N/A')}

Old Text: {change_data.get('old_text', 'N/A')}
New Text: {change_data.get('new_text', 'N/A')}

Provide a JSON response with:
1. summary: A brief 1-2 sentence summary of what changed
2. recommendation: A simple actionable recommendation (max 100 chars)

Response format:
{{"summary": "...", "recommendation": "..."}}"""

            response = self.client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {"role": "system", "content": "You are a pricing analyst helping a business understand competitor pricing changes."},
                    {"role": "user", "content": prompt}
                ],
                max_tokens=200,
                temperature=0.7
            )
            
            content = response.choices[0].message.content
            
            # Parse JSON from response
            try:
                # Try to extract JSON from response
                if "{" in content and "}" in content:
                    json_start = content.find("{")
                    json_end = content.rfind("}") + 1
                    json_str = content[json_start:json_end]
                    data = json.loads(json_str)
                    return data.get("summary", content[:100]), data.get("recommendation", "")[:500]
            except:
                pass
            
            # Fallback: return first 100 chars
            return content[:100], ""
            
        except Exception as e:
            logger.error(f"LLM error: {e}")
            return self.generate_rule_based_summary(change_data)
    
    def generate_rule_based_summary(self, change_data: Dict[str, Any]) -> tuple:
        """Generate summary using rules when LLM is not available"""
        change_type = change_data.get("change_type", "")
        old_price = change_data.get("old_price", "N/A")
        new_price = change_data.get("new_price", "N/A")
        change_percent = change_data.get("change_percent", 0)
        
        # Generate summary based on change type
        if "price_increase" in change_type:
            summary = f"Competitor increased price from {old_price} to {new_price} (+{change_percent:.1f}%)"
            recommendation = "Review your pricing strategy"
        elif "price_decrease" in change_type:
            summary = f"Competitor decreased price from {old_price} to {new_price} ({change_percent:.1f}%)"
            recommendation = "Consider matching or reducing price"
        elif "feature_added" in change_type:
            summary = "Competitor added new features"
            recommendation = "Evaluate feature gap"
        elif "feature_removed" in change_type:
            summary = "Competitor removed features"
            recommendation = "Highlight your superior features"
        elif "messaging_change" in change_type:
            summary = "Competitor changed messaging"
            recommendation = "Update your value proposition"
        else:
            summary = f"Competitor made changes: {change_type}"
            recommendation = "Analyze impact"
        
        return summary, recommendation
    
    def analyze_change(self, change_data: Dict[str, Any]) -> Optional[AIAnalysis]:
        """Analyze a single change and store the result"""
        # Get summary and recommendation
        if self.client:
            summary, recommendation = self.generate_summary_with_llm(change_data)
        else:
            summary, recommendation = self.generate_rule_based_summary(change_data)
        
        # Create analysis record
        analysis = AIAnalysis(
            change_id=change_data["change_id"],
            summary=summary,
            recommendation=recommendation[:500] if recommendation else None,
            change_type=change_data.get("change_type"),
            page_type=change_data.get("page_type"),
            old_price=change_data.get("old_price"),
            new_price=change_data.get("new_price"),
            change_percent=str(change_data.get("change_percent", "")),
            model="gpt-4o-mini" if self.client else "rule-based",
            created_at=datetime.utcnow()
        )
        
        # Save to database
        session = self.Session()
        session.add(analysis)
        session.commit()
        
        logger.info(f"✅ Analyzed change {change_data['change_id']}: {summary[:50]}...")
        
        session.close()
        return analysis
    
    def run_analysis(self):
        """Analyze all unanalyzed changes"""
        changes = self.get_unanalyzed_changes()
        
        if not changes:
            logger.info("No new changes to analyze")
            return 0
        
        logger.info(f"🔍 Analyzing {len(changes)} changes with AI...")
        
        analyzed = 0
        for change in changes:
            try:
                self.analyze_change(change)
                analyzed += 1
            except Exception as e:
                logger.error(f"Error analyzing change {change.get('change_id')}: {e}")
        
        logger.info(f"🎉 Analysis complete: {analyzed} changes analyzed")
        return analyzed


if __name__ == "__main__":
    service = AIAnalysisService()
    
    # Run analysis
    analyzed = service.run_analysis()
    print(f"Analyzed {analyzed} changes")
