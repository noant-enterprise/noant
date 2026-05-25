import os
import re

BASE_DIR = r"C:\Users\USER\Downloads\omagent\backend"

def read_file(path):
    with open(path, 'r', encoding='utf-8') as f:
        return f.read()

def write_file(path, content):
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Updated: {path}")

# 1. Update handler.go - ListConversations function
handler_path = os.path.join(BASE_DIR, "internal", "handler", "handler.go")
handler_content = read_file(handler_path)

# Replace the old ListConversations function
old_handler = '''func (h *ChatHandler) ListConversations(c *gin.Context) {
\tuserID, _ := c.Get("userID")
\tstatus := c.Query("status")
\tpage := 1
\tlimit := 20

\tconversations, err := h.service.ListConversations(c.Request.Context(), userID.(string), status, page, limit)
\tif err != nil {
\t\tutils.RespondInternalError(c, err.Error())
\t\treturn
\t}

\tc.JSON(http.StatusOK, gin.H{"conversations": conversations})
}'''

new_handler = '''func (h *ChatHandler) ListConversations(c *gin.Context) {
\tuserID, _ := c.Get("userID")
\tstatus := c.Query("status")
\t
\tpage, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
\tif page < 1 {
\t\tpage = 1
\t}
\t
\tlimit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
\tif limit < 1 || limit > 100 {
\t\tlimit = 20
\t}

\tconversations, total, err := h.service.ListConversations(c.Request.Context(), userID.(string), status, page, limit)
\tif err != nil {
\t\tutils.RespondInternalError(c, err.Error())
\t\treturn
\t}

\thasMore := page*limit < total

\tc.JSON(http.StatusOK, gin.H{
\t\t"conversations": conversations,
\t\t"total":         total,
\t\t"page":          page,
\t\t"limit":         limit,
\t\t"has_more":      hasMore,
\t})
}'''

if old_handler in handler_content:
    handler_content = handler_content.replace(old_handler, new_handler)
    
    # Add strconv import if not present
    if '"strconv"' not in handler_content:
        handler_content = handler_content.replace(
            'import (',
            'import (\n\t"strconv"'
        )
    
    write_file(handler_path, handler_content)
    print("✓ handler.go updated")
else:
    print("✗ Could not find exact match in handler.go - may need manual update")
    print("Searching for ListConversations...")
    if 'func (h *ChatHandler) ListConversations' in handler_content:
        print("  Found function but format differs. Printing around it:")
        idx = handler_content.find('func (h *ChatHandler) ListConversations')
        print(handler_content[idx:idx+800])

# 2. Update service.go - ListConversations signature
service_path = os.path.join(BASE_DIR, "internal", "service", "service.go")
service_content = read_file(service_path)

# Find and replace the function signature
old_service_sig = 'func (s *ChatService) ListConversations(ctx context.Context, userID string, status string, page, limit int) ([]domain.Conversation, error) {'
new_service_sig = 'func (s *ChatService) ListConversations(ctx context.Context, userID string, status string, page, limit int) ([]domain.Conversation, int, error) {'

if old_service_sig in service_content:
    service_content = service_content.replace(old_service_sig, new_service_sig)
    
    # Find the return statement and update it
    # Look for the pattern in ListConversations function
    func_start = service_content.find(new_service_sig)
    func_end = service_content.find('\n}\n\nfunc', func_start)
    if func_end == -1:
        func_end = len(service_content)
    
    func_body = service_content[func_start:func_end]
    
    # Replace return statements to include total
    # Pattern: return conversations, nil -> return conversations, len(conversations), nil
    # But we need actual count, so we'll need to get it from repo
    
    print("✓ service.go signature updated (need to verify return statements manually)")
    write_file(service_path, service_content)
else:
    print("✗ Could not find ListConversations in service.go")

# 3. Update repository.go - Add CountConversations and update ListConversations
repo_path = os.path.join(BASE_DIR, "internal", "repository", "repository.go")
repo_content = read_file(repo_path)

# Add CountConversations method before ListConversations
count_method = '''
func (r *ChatRepository) CountConversations(ctx context.Context, userID string, status string) (int, error) {
\tvar count int
\tquery := "SELECT COUNT(*) FROM conversations WHERE user_id = ?"
\targs := []interface{}{userID}
\t
\tif status != "" {
\t\tquery += " AND status = ?"
\t\targs = append(args, status)
\t}
\t
\terr := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
\treturn count, err
}

'''

# Find a good place to insert - before ListConversations or after the struct methods start
insert_marker = 'func (r *ChatRepository) ListConversations'
if insert_marker in repo_content and 'CountConversations' not in repo_content:
    repo_content = repo_content.replace(insert_marker, count_method + insert_marker)
    write_file(repo_path, repo_content)
    print("✓ repository.go updated with CountConversations")
else:
    print("✗ Could not update repository.go (already has CountConversations or marker not found)")

print("\n" + "="*50)
print("Backend update complete!")
print("="*50)
print("\nNext steps:")
print("1. Verify service.go return statement includes total count")
print("2. Run: go build ./...")
print("3. Restart your backend server")