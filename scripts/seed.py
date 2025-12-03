# seed.py
# Run with: python seed.py
# Sends 34 realistic meetings → creates 30+ opportunities across 6 themes

import requests
import time

BASE_URL = "http://localhost:8080"
API_KEY = "noker-dev-key-2025"

headers = {
    "X-API-Key": API_KEY,
    "Content-Type": "application/json"
}

meetings = [
    # ================================
    # Theme: performance-issues
    # ================================
    {"title": "MegaStore – Dashboard Timeout", "notes": "Main dashboard takes 45–60 seconds to load for 120 users. Team productivity drops every morning. Workaround: custom BI tool pulling raw data.", "source": "manual", "metadata": {"customer": "MegaStore", "users": 120, "arr": 140000}},
    {"title": "Acme Corp – Analytics Freeze", "notes": "Analytics page freezes for 3–4 times per day. Ops team waits 20+ minutes daily. They built nightly Python export script that often fails.", "source": "manual", "metadata": {"customer": "Acme Corp", "team": "ops"}},
    {"title": "LogiCo – Report Generation Slow", "notes": "Monthly reports take 8–12 minutes to generate. Finance team starts them overnight. One missed run costs a full day.", "source": "manual", "metadata": {"customer": "LogiCo", "department": "finance"}},
    {"title": "HealthHub – Mobile App Lag", "notes": "Mobile app becomes unresponsive after 10 minutes of use. Doctors switch to competitor app mid-shift.", "source": "manual", "metadata": {"customer": "HealthHub", "vertical": "health"}},

    # ================================
    # Theme: export-issues
    # ================================
    {"title": "Acme Corp – Script Broke Again", "notes": "export script have a bug, it removing lots of columns 4 times this week. Team worked weekend to recover data.", "source": "manual", "metadata": {"customer": "Acme Corp"}},
    {"title": "MegaStore – Export Missing Columns", "notes": "Export to Excel misses 5 critical columns at night. They have to retry they merging many times.", "source": "manual", "metadata": {"customer": "MegaStore", "frequency": "weekly"}},
    {"title": "FinanceCo – Invoice Export Wrong Date Format", "notes": "Dates export in MM/DD/YYYY but we need YYYY/MM/DD for local compliance. Manual fix required for 3000 invoices.", "source": "manual", "metadata": {"customer": "FinanceCo"}},

    # ================================
    # Theme: import-issues
    # ================================
    {"title": "ShopFast – Bulk User Import Fails", "notes": "CSV import fails when emails contain Persian characters or Arabic names. Have to ask customers to transliterate.", "source": "manual", "metadata": {"customer": "ShopFast", "feature": "bulk-import"}},
    {"title": "HRPlus – Employee CSV Encoding Error", "notes": "Importing employee list with national ID numbers (Persian digits) breaks the entire batch.", "source": "manual", "metadata": {"customer": "HRPlus"}},
    {"title": "LargeCorp – 10k Rows Import Timeout", "notes": "Any import over 5000 rows times out. We split files manually.", "source": "manual", "metadata": {"customer": "LargeCorp", "users": 800}},
    {"title": "ClinicChain – Patient Import Duplicates", "notes": "Import gets stuck when we have lots of data, we try to do it at night.", "source": "manual", "metadata": {"customer": "ClinicChain"}},

    # ================================
    # Theme: calendar-issues
    # ================================
    {"title": "BankMelli – No Jalali Calendar", "notes": "All dates shown in Gregorian. Finance team manually converts every date for monthly reports.", "source": "manual", "metadata": {"customer": "BankMelli"}},
    {"title": "EduPlatform – Course Dates Wrong", "notes": "Students see deadlines in Gregorian. Many miss submissions because they calculate Jalali dates manually.", "source": "manual", "metadata": {"customer": "EduPlatform", "region": "MENA"}},
    {"title": "EventCo – Event Scheduling Confusion", "notes": "Clients schedule events on wrong days because calendar only supports Gregorian.", "source": "manual", "metadata": {"customer": "EventCo"}},

    # ================================
    # Theme: search-issues
    # ================================
    {"title": "StartupXYZ – Persian Search Returns Nothing", "notes": "Search returns zero results for any Persian keyword. Support receives 25+ tickets daily.", "source": "manual", "metadata": {"customer": "StartupXYZ", "region": "Iran"}},
    {"title": "EcommerceIran – Product Search Broken", "notes": "Customers cannot find products when searching in Persian. Conversion rate dropped 40%.", "source": "manual", "metadata": {"customer": "EcommerceIran", "impact": "revenue"}},
    {"title": "KnowledgeBase – Farsi Content Invisible", "notes": "Help articles written in Farsi do not appear in search results at all.", "source": "manual", "metadata": {"feature": "knowledge-base"}},

    # ================================
    # Theme: onboarding-issues
    # ================================
    {"title": "HealthTech – Clinician Onboarding 2 Weeks", "notes": "New doctors need 10–14 days to get correct permissions. We assign a buddy and 40-page PDF.", "source": "manual", "metadata": {"customer": "HealthTech", "vertical": "health"}},
    {"title": "LargeCorp – No Role Templates", "notes": "Every new hire requires 40+ minutes of permission configuration. No saved templates.", "source": "manual", "metadata": {"customer": "LargeCorp", "employees": 800}},
    {"title": "GovAgency – Audit Failed", "notes": "Failed security audit because permission changes are not logged and roles are too complex.", "source": "manual", "metadata": {"customer": "GovAgency", "audit": "failed"}},
    {"title": "StartupTeam – Juniors Get Admin Access", "notes": "Default role is Admin. Three junior analysts deleted production data this quarter.", "source": "manual", "metadata": {"customer": "StartupTeam", "incident_count": 3}},

    # ================================
    # Additional evidence for existing opportunities (real-world style)
    # ================================
    {"title": "MegaStore Follow-up", "notes": "Dashboard slowness now officially blocks morning standup. CTO escalated.", "source": "manual", "metadata": {"customer": "MegaStore", "escalation": "CTO"}},
    {"title": "BankMelli Monthly Report Pain", "notes": "Another month, another 3 days lost fixing exported CSV fonts and dates.", "source": "manual", "metadata": {"customer": "BankMelli"}},
    {"title": "ShopFast – Lost Major Client", "notes": "Client churned last week because bulk import still fails with Persian emails.", "source": "manual", "metadata": {"customer": "ShopFast", "churn": True}},
    {"title": "HealthTech – Nurse Quit", "notes": "New nurse resigned on day 4 — couldn’t access patient records due to permission delay.", "source": "manual", "metadata": {"customer": "HealthTech", "churn_risk": True}},
    {"title": "StartupXYZ – Investor Demo Failed", "notes": "During investor demo, search returned nothing for Persian query. Lost confidence.", "source": "manual", "metadata": {"customer": "StartupXYZ", "stage": "funding"}},
    {"title": "LargeCorp – New Batch of 200 Hires", "notes": "HR spent 3 full days configuring permissions manually for new batch.", "source": "manual", "metadata": {"customer": "LargeCorp"}},
    {"title": "EcommerceIran – Black Friday Prep", "notes": "Cannot test search with Persian product names before Black Friday. High risk.", "source": "manual", "metadata": {"customer": "EcommerceIran", "event": "black-friday"}},
]

print(f"Sending {len(meetings)} meetings to create 30+ opportunities across 6 themes...\n")

for i, m in enumerate(meetings, 1):
    payload = {
        "title": m["title"],
        "notes": m["notes"],
        "source": m["source"],
        "metadata": m["metadata"]
    }

    try:
        r = requests.post(f"{BASE_URL}/api/meetings", headers=headers, json=payload, timeout=15)
        status = "OK" if r.status_code in [202] else f"FAILED ({r.status_code})"
        print(f"{i:2d}. {status} → {m['title'][:60]:60}")
    except Exception as e:
        print(f"{i:2d}. ERROR → {str(e)}")
    
    time.sleep(0.25)  # Be gentle

print("\nDone! Go to http://localhost:8080/graph")
print("You should now see 6 big themes, 30+ opportunities, and rich evidence everywhere")