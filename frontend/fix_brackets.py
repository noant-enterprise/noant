import os
import re

# Fix chats/page.tsx
path = r"src\app\(dashboard)\chats\page.tsx"
with open(path, 'r', encoding='utf-8') as f:
    content = f.read()

# Replace all << before generics with single <
content = content.replace('useAPI<<ConversationsResponse>()', 'useAPI<<ConversationsResponse>()')
content = content.replace('useState<<Conversation[]>([])', 'useState<<Conversation[]>([])')

with open(path, 'w', encoding='utf-8') as f:
    f.write(content)

# Verify
with open(path, 'r', encoding='utf-8') as f:
    check = f.read()

if '<<ConversationsResponse' in check or '<<Conversation[]' in check:
    print("FAIL: still has <<")
else:
    print("OK: all << fixed")

# Also fix any other files
for root, dirs, files in os.walk('src'):
    for file in files:
        if file.endswith(('.ts', '.tsx')):
            fp = os.path.join(root, file)
            with open(fp, 'r', encoding='utf-8') as f:
                c = f.read()
            fixed = c.replace('<<ConversationsResponse>', '<ConversationsResponse>')
            fixed = fixed.replace('<<Conversation[]>', '<Conversation[]>')
            fixed = fixed.replace('<<HTMLDivElement>', '<HTMLDivElement>')
            if fixed != c:
                with open(fp, 'w', encoding='utf-8') as f:
                    f.write(fixed)
                print(f"Fixed: {fp}")
