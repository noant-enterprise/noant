import os

path = "backend/internal/repository/repository.go"
with open(path, "r", encoding="utf-8") as f:
    content = f.read()

# FIX 1: CategoryRepository.Create
old = """func (r *CategoryRepository) Create(ctx context.Context, cat *domain.Category) error {
	query := `INSERT INTO categories (id, name, description, color, created_at)
			  VALUES (UUID(), ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, cat.Name, cat.Description, cat.Color)
	return err
}"""

new = """func (r *CategoryRepository) Create(ctx context.Context, cat *domain.Category) error {
	query := `INSERT INTO categories (id, name, description, color, created_at)
			  VALUES (UUID(), ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, cat.Name, cat.Description, cat.Color)
	if err != nil {
		return err
	}

	var id string
	err = r.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ? ORDER BY created_at DESC LIMIT 1", cat.Name).Scan(&id)
	if err == nil {
		cat.ID = id
	}
	return nil
}"""

if old in content:
    content = content.replace(old, new)
    print("✅ CategoryRepository.Create fixed")
else:
    print("⚠️ CategoryRepository.Create not found")

# FIX 2: TeamRepository.Create
old = """func (r *TeamRepository) Create(ctx context.Context, ownerID string, member *domain.TeamMember) error {
	query := `INSERT INTO team_members (id, owner_id, user_id, role, is_active, joined_at)
			  VALUES (UUID(), ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, ownerID, member.UserID, member.Role, member.IsActive)
	return err
}"""

new = """func (r *TeamRepository) Create(ctx context.Context, ownerID string, member *domain.TeamMember) error {
	query := `INSERT INTO team_members (id, owner_id, user_id, role, is_active, joined_at)
			  VALUES (UUID(), ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, ownerID, member.UserID, member.Role, member.IsActive)
	if err != nil {
		return err
	}

	var id string
	err = r.db.QueryRowContext(ctx, "SELECT id FROM team_members WHERE owner_id = ? AND user_id = ? ORDER BY joined_at DESC LIMIT 1", ownerID, member.UserID).Scan(&id)
	if err == nil {
		member.ID = id
	}
	return nil
}"""

if old in content:
    content = content.replace(old, new)
    print("✅ TeamRepository.Create fixed")
else:
    print("⚠️ TeamRepository.Create not found")

# FIX 3: APIKeyRepository.Create
old = """func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `INSERT INTO api_keys (id, user_id, name, key_hash, is_active, created_at)
			  VALUES (UUID(), ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, key.UserID, key.Name, key.Key, key.IsActive)
	return err
}"""

new = """func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `INSERT INTO api_keys (id, user_id, name, key_hash, is_active, created_at)
			  VALUES (UUID(), ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, key.UserID, key.Name, key.Key, key.IsActive)
	if err != nil {
		return err
	}

	var id string
	err = r.db.QueryRowContext(ctx, "SELECT id FROM api_keys WHERE user_id = ? AND name = ? ORDER BY created_at DESC LIMIT 1", key.UserID, key.Name).Scan(&id)
	if err == nil {
		key.ID = id
	}
	return nil
}"""

if old in content:
    content = content.replace(old, new)
    print("✅ APIKeyRepository.Create fixed")
else:
    print("⚠️ APIKeyRepository.Create not found")

# FIX 4: ArchiveRepository.CreateFolder
old = """func (r *ArchiveRepository) CreateFolder(ctx context.Context, folder *domain.ArchiveFolder) error {
	query := `INSERT INTO archive_folders (id, user_id, name, type, color, created_at)
			  VALUES (UUID(), ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, folder.UserID, folder.Name, folder.Type, folder.Color)
	return err
}"""

new = """func (r *ArchiveRepository) CreateFolder(ctx context.Context, folder *domain.ArchiveFolder) error {
	query := `INSERT INTO archive_folders (id, user_id, name, type, color, created_at)
			  VALUES (UUID(), ?, ?, ?, ?, NOW())`

	_, err := r.db.ExecContext(ctx, query, folder.UserID, folder.Name, folder.Type, folder.Color)
	if err != nil {
		return err
	}

	var id string
	err = r.db.QueryRowContext(ctx, "SELECT id FROM archive_folders WHERE user_id = ? AND name = ? ORDER BY created_at DESC LIMIT 1", folder.UserID, folder.Name).Scan(&id)
	if err == nil {
		folder.ID = id
	}
	return nil
}"""

if old in content:
    content = content.replace(old, new)
    print("✅ ArchiveRepository.CreateFolder fixed")
else:
    print("⚠️ ArchiveRepository.CreateFolder not found")

with open(path, "w", encoding="utf-8") as f:
    f.write(content)

# BUILD
os.chdir("backend")
r = os.system("go build -o noant.exe .")
print("Build successful" if r == 0 else "Build failed")