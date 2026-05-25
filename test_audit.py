import requests
import json
import time

# Login
r = requests.post('http://localhost:5000/api/v1/auth/login', json={
    'email': 'audit_test@noant.com',
    'password': 'SecurePass123!'
})
token = r.json()['token']
print(f'Login: {r.status_code}')

# Create something to trigger audit
r = requests.post('http://localhost:5000/api/v1/training/categories', 
    headers={'Authorization': f'Bearer {token}', 'Content-Type': 'application/json'},
    json={'name': 'AuditTest2', 'description': 'Test', 'color': '#ff0000'})
print(f'Create Category: {r.status_code}')

# Wait for async audit
time.sleep(2)

# Fetch audit logs
r = requests.get('http://localhost:5000/api/v1/settings/audit-logs',
    headers={'Authorization': f'Bearer {token}'})
print(f'Audit Logs: {r.status_code}')

try:
    data = r.json()
    count = data.get('count', 0)
    print(f'Count: {count}')
    for log in data.get('audit_logs', [])[:5]:
        action = log.get('action', 'unknown')
        uid = log.get('user_id', 'unknown')
        created = log.get('created_at', 'unknown')
        print(f'  - {action} by {uid[:8]}... at {created}')
except Exception as e:
    print(f'Error: {e}')
    print(r.text[:500])