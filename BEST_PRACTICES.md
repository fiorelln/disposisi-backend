# Best Practices - Sistem Disposisi Surat

## 1. Architecture Best Practices

### 1.1 Clean Architecture Principles
```
✓ Separation of Concerns
  - Controllers: HTTP layer only
  - Services: Business logic
  - Repositories: Database operations
  - DTOs: Input/Output contracts

✓ Dependency Injection
  - Inject dependencies lewat constructor
  - Interface-based design
  - Easy to test dan mock

✓ Error Handling
  - Specific error types
  - Wrap errors dengan context
  - Log meaningful errors
```

### 1.2 Database Transaction Best Practices
```go
// ✓ GOOD: Use transaction untuk multi-step operations
tx := s.db.BeginTx(s.db.Statement.Context, nil)
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := s.repo.Create(disposisi); err != nil {
    tx.Rollback()
    return err
}

if err := s.repo.UpdateStatus(parentID, status); err != nil {
    tx.Rollback()
    return err
}

tx.Commit()

// ✗ BAD: Multiple sequential operations tanpa transaction
s.repo.Create(disposisi)       // Bisa jadi orphan jika step selanjutnya gagal
s.repo.UpdateStatus(parentID)  // Error di sini akan inconsistent state
```

### 1.3 Preload Relations Efficiently
```go
// ✓ GOOD: Preload yang diperlukan saja
var disposisi models.Disposisi
s.db.
    Preload("FromUser").
    Preload("ToUser").
    Where("id = ?", id).
    First(&disposisi)

// ✗ BAD: Preload semua relations
var disposisi models.Disposisi
s.db.
    Preload("FromUser").
    Preload("ToUser").
    Preload("SuratMasuk").
    Preload("ChildDisposisi").
    Preload("ParentDisposisi").
    Where("id = ?", id).
    First(&disposisi)
```

## 2. Service Layer Best Practices

### 2.1 Business Logic Organization
```go
// ✓ GOOD: Clear separation of concerns
type DisposisiService interface {
    // Core operations
    ForwardDisposisi(disposisiID, toUserID, currentUserID uint) error
    CompleteDisposisi(disposisiID, currentUserID uint) error
    
    // Query operations
    GetInbox(userID uint, page, pageSize int) (*InboxListResponse, error)
    GetHistory(suratMasukID uint) (*HistoryResponse, error)
    
    // Validation
    ValidateForward(disposisiID, toUserID, currentUserID uint) (bool, error)
}

// ✗ BAD: Mixed concerns in one method
func (s *Service) ProcessDisposisi(id uint) error {
    // Validate
    // Save to DB
    // Create notification
    // Send email
    // Log
    // ... terlalu banyak di satu method
}
```

### 2.2 Permission Validation Pattern
```go
// ✓ GOOD: Explicit validation dengan delegation
func (s *DisposisiService) ForwardDisposisi(...) error {
    // 1. Get entity
    disposisi, err := s.repo.GetByID(disposisiID)
    
    // 2. Validate ownership
    if disposisi.ToUserID != currentUserID {
        return errors.New("unauthorized")
    }
    
    // 3. Validate status
    if !disposisi.IsPending() {
        return errors.New("cannot forward non-pending disposisi")
    }
    
    // 4. Delegate permission check
    canForward, err := s.permissionSvc.CanForward(currentUserID, toUserID)
    if !canForward {
        return errors.New("no permission to forward")
    }
    
    // 5. Execute business logic
    return s.repo.Create(childDisposisi)
}
```

### 2.3 Notification Pattern
```go
// ✓ GOOD: Async notification (non-blocking)
// Di service:
if err := s.repo.Create(disposisi); err != nil {
    return err
}

// Notify async (fire & forget)
go s.notificationSvc.NotifyDisposisiReceived(disposisi.ID, toUserID)

// ✗ BAD: Synchronous notification (blocking)
// Jika notification service lambat, akan block main flow
if err := s.notificationSvc.NotifyDisposisiReceived(...); err != nil {
    return err // Disposisi created tapi notification failed
}
```

## 3. Repository Pattern Best Practices

### 3.1 Query Optimization
```go
// ✓ GOOD: Pagination untuk large results
func (r *Repository) GetInbox(userID uint, page, pageSize int) ([]Disposisi, int64, error) {
    offset := (page - 1) * pageSize
    
    // Count untuk total
    var total int64
    r.db.Where(...).Model(&Disposisi{}).Count(&total)
    
    // Get dengan offset
    var disposisi []Disposisi
    r.db.Where(...).Offset(offset).Limit(pageSize).Find(&disposisi)
    
    return disposisi, total, nil
}

// ✗ BAD: Load semua records
func (r *Repository) GetInbox(userID uint) ([]Disposisi, error) {
    var disposisi []Disposisi
    r.db.Where("to_user_id = ?", userID).Find(&disposisi)
    // Membebani memory jika ada 10000+ records
    return disposisi, nil
}
```

### 3.2 Soft Delete Pattern
```go
// ✓ GOOD: Always check deleted_at
func (r *Repository) GetByID(id uint) (*Disposisi, error) {
    var disposisi Disposisi
    r.db.Where("id = ? AND deleted_at IS NULL", id).First(&disposisi)
    return &disposisi, nil
}

// ✗ BAD: Forget to filter deleted_at
func (r *Repository) GetByID(id uint) (*Disposisi, error) {
    var disposisi Disposisi
    r.db.First(&disposisi, id) // Bisa return deleted records
    return &disposisi, nil
}
```

### 3.3 Index Strategy
```sql
-- ✓ GOOD: Index pada frequently queried columns
CREATE INDEX idx_disposisi_to_user_id ON disposisi(to_user_id, deleted_at);
CREATE INDEX idx_disposisi_from_user_id ON disposisi(from_user_id, deleted_at);
CREATE INDEX idx_disposisi_surat_masuk_id ON disposisi(surat_masuk_id);
CREATE INDEX idx_disposisi_parent_id ON disposisi(parent_disposisi_id);
CREATE INDEX idx_disposisi_status ON disposisi(status, deleted_at);
```

## 4. API/Controller Best Practices

### 4.1 Request Validation
```go
// ✓ GOOD: Validate di controller/request level
type CreateForwardRequest struct {
    ToUserID   uint   `json:"to_user_id" binding:"required"`
    Catatan    string `json:"catatan" binding:"max=1000"`
    Sifat      string `json:"sifat" binding:"oneof=penting biasa rahasia"`
}

// ✗ BAD: Minimal validation
type CreateForwardRequest struct {
    ToUserID   uint
    Catatan    string
}
```

### 4.2 Error Response Standardization
```go
// ✓ GOOD: Consistent error response
type ErrorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

// Example responses:
// 400 Bad Request
{
    "code": "INVALID_REQUEST",
    "message": "Validation failed",
    "details": {
        "to_user_id": "required field"
    }
}

// 403 Forbidden
{
    "code": "INSUFFICIENT_PERMISSION",
    "message": "You don't have permission to forward to this user"
}

// 404 Not Found
{
    "code": "DISPOSISI_NOT_FOUND",
    "message": "Disposisi dengan ID tersebut tidak ditemukan"
}
```

### 4.3 Pagination Response
```go
// ✓ GOOD: Include metadata untuk pagination
{
    "data": [...],
    "pagination": {
        "page": 1,
        "page_size": 20,
        "total_items": 150,
        "total_pages": 8
    }
}

// Memudahkan frontend untuk:
// - Display current page
// - Calculate available pages
// - Show "X of Y" indicators
```

## 5. Permission/Security Best Practices

### 5.1 Permission Checking Pattern
```go
// ✓ GOOD: Always check:
// 1. Resource ownership
// 2. Action permission
// 3. Status validation

func (s *Service) CompleteDisposisi(id uint, currentUserID uint) error {
    // 1. Get resource
    disposisi, _ := s.repo.GetByID(id)
    
    // 2. Check ownership
    if disposisi.ToUserID != currentUserID {
        return errors.New("not the recipient")
    }
    
    // 3. Check status
    if !disposisi.IsPending() {
        return errors.New("invalid status for completion")
    }
    
    // 4. Execute
    return s.repo.UpdateStatus(id, models.StatusCompleted)
}
```

### 5.2 Role-Based Access Control (RBAC)
```go
// ✓ GOOD: Hierarchical role checking
type RoleLevel int
const (
    RoleHeadmaster   RoleLevel = 0  // Highest
    RoleViceHead     RoleLevel = 1
    RoleTU           RoleLevel = 2  // TU/Guru
    RoleTeacher      RoleLevel = 2
)

// Can forward:
// - To higher level (any role can forward up)
// - TU & Kepsek can forward anywhere
// - WAKA can forward to bawahan
// - Guru can forward up or to Koordinator

func (p *PermissionService) CanForward(fromRole, toRole string) bool {
    fromLevel := roleLevelMap[fromRole]
    toLevel := roleLevelMap[toRole]
    
    // Can always forward up
    if toLevel < fromLevel {
        return true
    }
    
    // TU dan Kepsek can forward anywhere
    if fromRole == "TATA_USAHA" || fromRole == "KEPALA_SEKOLAH" {
        return true
    }
    
    // ... other rules
}
```

## 6. Workflow Best Practices

### 6.1 Disposisi Workflow State Machine
```
pending → (user baca) → pending (with dibaca=true)
        → (user forward) → forwarded (create child)
                        → parent status: forwarded
                        → child status: pending

        → (user complete) → completed
                          → auto check parent: all children completed?
                          → if yes: parent status → completed

        → (user reject) → rejected
```

### 6.2 Auto-Parent Status Update Logic
```go
// ✓ GOOD: Automatic cascade completion
func (s *Service) CompleteDisposisi(...) error {
    // ... complete child
    
    // Check if all siblings completed
    siblings, _ := s.repo.GetChildDisposisi(parentID)
    
    allCompleted := true
    for _, sibling := range siblings {
        if !sibling.IsCompleted() && sibling.Status != "rejected" {
            allCompleted = false
            break
        }
    }
    
    if allCompleted {
        s.repo.UpdateStatus(parentID, models.StatusCompleted)
    }
}
```

## 7. Query Performance Best Practices

### 7.1 Inbox Query Optimization
```sql
-- ✓ GOOD: Use proper indexes
SELECT d.* 
FROM disposisi d
WHERE d.to_user_id = ? 
    AND d.deleted_at IS NULL
ORDER BY d.created_at DESC
LIMIT 20 OFFSET 0;

-- Index: (to_user_id, deleted_at, created_at)

-- ✗ BAD: Missing indexes
SELECT d.* 
FROM disposisi d
WHERE d.deleted_at IS NULL
    AND d.to_user_id IN (SELECT user_id FROM user_jabatan WHERE jabatan_id = ?)
ORDER BY d.created_at DESC;
```

### 7.2 Tree Query Optimization
```go
// ✓ GOOD: Use recursive CTE untuk tree queries
WITH RECURSIVE disposisi_tree AS (
    SELECT id, parent_id, level FROM disposisi WHERE id = ?
    UNION ALL
    SELECT d.id, d.parent_id, d.level FROM disposisi d
    INNER JOIN disposisi_tree dt ON d.parent_id = dt.id
)
SELECT * FROM disposisi_tree;

// ✗ BAD: Load all records in memory
var allDisposisi []Disposisi
s.db.Find(&allDisposisi) // Load semua
// Then iterate manually untuk build tree
```

## 8. Testing Best Practices

### 8.1 Unit Testing Pattern
```go
func TestForwardDisposisi(t *testing.T) {
    // Arrange
    mockRepo := &MockRepository{}
    mockPermission := &MockPermissionService{}
    service := NewDisposisiService(mockRepo, mockPermission)
    
    mockRepo.On("GetByID", uint(1)).Return(&models.Disposisi{
        ID:     1,
        ToUserID: 2,
        Status: "pending",
    }, nil)
    
    mockPermission.On("CanForward", uint(2), uint(3)).Return(true, nil)
    mockRepo.On("Create", mock.Anything).Return(nil)
    
    // Act
    err := service.ForwardDisposisi(1, &dto.CreateForwardRequest{
        ToUserID: 3,
        Catatan:  "test",
    }, 2)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertCalled(t, "Create", mock.Anything)
}
```

## 9. Logging & Monitoring Best Practices

### 9.1 Audit Trail
```go
// ✓ GOOD: Log semua important actions
func (s *Service) ForwardDisposisi(...) error {
    log.Printf("User %d forwarding disposisi %d to user %d", 
        currentUserID, disposisiID, toUserID)
    
    if err := s.repo.Create(childDisposisi); err != nil {
        log.Printf("ERROR: Failed to create child disposisi: %v", err)
        return err
    }
    
    log.Printf("SUCCESS: Disposisi %d forwarded, created child %d",
        disposisiID, childDisposisi.ID)
}
```

## 10. Common Pitfalls & Solutions

| Pitfall | Problem | Solution |
|---------|---------|----------|
| No transaction | Orphaned records | Use tx.BeginTx() for multi-step |
| N+1 queries | Slow performance | Use Preload() efficiently |
| Forgot soft delete | Show deleted items | Always filter deleted_at IS NULL |
| No pagination | Memory overflow | Always use LIMIT/OFFSET |
| Missing permission check | Security issue | Always validate in service layer |
| Blocking notification | Slow API | Use goroutine or queue |
| No error logging | Hard to debug | Log with context dan stack trace |
| Magic numbers | Unmaintainable | Use constants |
| No input validation | Injection attacks | Use struct tags binding |
| Circular preload | Memory leak | Limit preload levels |

