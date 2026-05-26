# Workflow & Implementation Guide

## 1. Project Structure Overview

```
disposisi/
├── models/                          # Database models
│   ├── disposisi.go                # ✓ DONE - Enhanced model dengan parent-child
│   ├── user.go                     # User model
│   ├── user_jabatan.go             # User-Jabatan relationship
│   ├── jabatan.go                  # Jabatan (roles)
│   ├── surat_masuk.go
│   └── notifikasi.go               # ✓ DONE - For notifications
│
├── dto/                            # Request/Response Data Transfer Objects
│   └── disposisi.go                # ✓ DONE - Comprehensive DTOs
│
├── repositories/                   # Database Access Layer
│   └── disposisi_repository.go     # ✓ DONE - Repository pattern implementation
│
├── services/                       # Business Logic Layer
│   ├── disposisi_service.go        # ✓ DONE - Core forwarding logic
│   ├── permission_service.go       # ✓ DONE - Role-based permission validation
│   └── notification_service.go     # ✓ DONE - Notification handling
│
├── helpers/                        # Utility functions
│   ├── jwt.go                      # Existing JWT helper
│   └── disposisi_query_helper.go   # ✓ DONE - Query builders & helpers
│
├── controllers/                    # HTTP Handlers (TODO - next step)
│   └── disposisi.go                # To be implemented
│
├── middlewares/                    # HTTP Middlewares (existing)
│   ├── auth.go
│   └── role.go
│
├── routes/                         # Route definitions
│   └── route.go                    # Update dengan disposisi routes
│
├── config/                         # Configuration
│   └── db.go                       # Database setup
│
└── DISPOSISI_ARCHITECTURE.md       # ✓ DONE - Architecture documentation
└── BEST_PRACTICES.md               # ✓ DONE - Best practices guide
```

## 2. Layer Integration Flow

### 2.1 Request Flow (HTTP → Response)

```
HTTP Request
    ↓
┌─────────────────────────────────┐
│  Controller/Handler             │
│  - Validate input               │
│  - Extract JWT claims (userID)  │
│  - Call service method          │
└──────────────┬──────────────────┘
               ↓
┌─────────────────────────────────┐
│  Service Layer                  │
│  - Business logic               │
│  - Permission checks            │
│  - Transaction management       │
│  - Call repository & other svcs │
└──────────────┬──────────────────┘
               ↓
    ┌──────────┴──────────┬──────────────┐
    ↓                     ↓              ↓
┌─────────────────┐ ┌──────────────┐ ┌──────────────────┐
│  Repository     │ │ Permission   │ │ Notification     │
│  (Database)     │ │ Service      │ │ Service (go)     │
│  - CRUD ops     │ │ - Role check │ │ - Create record  │
│  - Queries      │ │ - Hierarchy  │ │ - Send email/push│
└─────────────────┘ └──────────────┘ └──────────────────┘
    ↓
┌─────────────────────────────────┐
│  Database (PostgreSQL)          │
│  - Persistent storage           │
│  - Transaction handling         │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│  Response (DTO)                 │
│  - JSON serialization           │
│  - HTTP status code             │
└─────────────────────────────────┘
    ↓
HTTP Response
```

## 3. Core Use Cases Implementation

### 3.1 Use Case: Forward Disposisi (Individual)

**Scenario:**
- Kepsek menerima disposisi dari TU tentang surat masuk
- Kepsek forward ke WAKA Kurikulum untuk ditangani

**Code Flow:**

```go
// 1. CONTROLLER - Handle HTTP request
func (c *DisposisiController) ForwardDisposisi(ctx *gin.Context) {
    // Extract user ID dari JWT token
    userID := ctx.GetUint("user_id")
    
    // Parse request
    var req dto.CreateForwardRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(400, dto.ErrorResponse{
            Code:    "INVALID_REQUEST",
            Message: "Invalid request format",
        })
        return
    }
    
    // Get disposisi ID dari URL
    disposisiID := ctx.Param("id")
    
    // Call service
    newDisposisi, err := c.service.ForwardDisposisi(
        uint(disposisiID),
        &req,
        userID,
    )
    
    if err != nil {
        ctx.JSON(400, dto.ErrorResponse{
            Code:    "FORWARD_FAILED",
            Message: err.Error(),
        })
        return
    }
    
    // Return success response
    ctx.JSON(200, newDisposisi)
}

// 2. SERVICE - Business logic
func (s *DisposisiServiceImpl) ForwardDisposisi(
    disposisiID uint,
    req *dto.CreateForwardRequest,
    currentUserID uint,
) (*models.Disposisi, error) {
    // Step 1: Get parent disposisi
    parentDisposisi, err := s.repo.GetByID(disposisiID)
    if err != nil {
        return nil, fmt.Errorf("parent disposisi not found: %w", err)
    }
    
    // Step 2: Validate user is recipient
    if parentDisposisi.ToUserID != currentUserID {
        return nil, errors.New("unauthorized: not recipient")
    }
    
    // Step 3: Validate permission (delegate to PermissionService)
    canForward, err := s.permissionSvc.CanForward(currentUserID, req.ToUserID)
    if err != nil || !canForward {
        return nil, errors.New("no permission to forward")
    }
    
    // Step 4: Create child disposisi
    childDisposisi := &models.Disposisi{
        SuratMasukID:      parentDisposisi.SuratMasukID,
        FromUserID:        currentUserID,
        ToUserID:          req.ToUserID,
        ParentDisposisiID: &disposisiID,
        Level:             parentDisposisi.Level + 1,
        Status:            models.StatusPending,
        Catatan:           req.Catatan,
    }
    
    // Step 5: Use transaction untuk consistency
    tx := s.db.BeginTx(s.db.Statement.Context, nil)
    
    // Create child
    if err := s.repo.Create(childDisposisi); err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to create child: %w", err)
    }
    
    // Update parent status
    if err := s.repo.UpdateStatus(disposisiID, models.StatusForwarded); err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to update parent: %w", err)
    }
    
    // Mark parent as read
    s.repo.MarkAsRead(disposisiID)
    
    tx.Commit()
    
    // Step 6: Send notification (async - non-blocking)
    go s.notificationSvc.NotifyDisposisiReceived(childDisposisi.ID, req.ToUserID)
    
    return childDisposisi, nil
}

// 3. PERMISSION SERVICE - Permission validation
func (p *PermissionServiceImpl) CanForward(
    fromUserID uint,
    toUserID uint,
) (bool, error) {
    // Get jabatan masing-masing
    fromJabatans, _ := p.GetUserJabatan(fromUserID)
    toJabatans, _ := p.GetUserJabatan(toUserID)
    
    if len(fromJabatans) == 0 || len(toJabatans) == 0 {
        return false, errors.New("user has no jabatan")
    }
    
    fromRole := fromJabatans[0].NamaJabatan
    toRole := toJabatans[0].NamaJabatan
    
    // Check role hierarchy
    return p.CanForwardToRole(fromRole, toRole), nil
}

func (p *PermissionServiceImpl) CanForwardToRole(fromRole, toRole string) bool {
    // TU dan Kepsek bisa forward ke siapa saja
    if fromRole == "TATA_USAHA" || fromRole == "KEPALA_SEKOLAH" {
        return true
    }
    
    // WAKA bisa forward ke Guru, Staff, atau WAKA lain
    if strings.Contains(fromRole, "WAKIL") {
        if toRole == "GURU" || toRole == "STAFF" || strings.Contains(toRole, "WAKIL") {
            return true
        }
    }
    
    // Guru/Staff hanya bisa forward ke level atas
    if fromRole == "GURU" || fromRole == "STAFF" {
        if toRole == "WAKIL_KEPALA" || toRole == "KEPALA_SEKOLAH" || strings.Contains(toRole, "KORDINATOR") {
            return true
        }
    }
    
    return false
}

// 4. REPOSITORY - Database operations
func (r *DisposisiRepositoryImpl) Create(disposisi *models.Disposisi) error {
    return r.db.Create(disposisi).Error
}

func (r *DisposisiRepositoryImpl) UpdateStatus(id uint, status string) error {
    return r.db.
        Model(&models.Disposisi{}).
        Where("id = ? AND deleted_at IS NULL", id).
        Update("status", status).Error
}

// 5. NOTIFICATION SERVICE - Async notification
func (n *NotificationServiceImpl) NotifyDisposisiReceived(
    disposisiID uint,
    recipientID uint,
) error {
    // Create notifikasi record
    notif := &models.Notifikasi{
        UserID:      recipientID,
        Type:        "disposisi_received",
        Title:       "Disposisi Baru",
        Message:     "Anda menerima disposisi baru",
        IsRead:      false,
        CreatedAt:   time.Now(),
    }
    
    n.db.Create(notif)
    
    // TODO: Send email, push notification, etc
    
    return nil
}
```

**Result:**
```
Database State:
- Parent: Disposisi#1 (Kepsek←TU) status=forwarded
  └─ Child: Disposisi#2 (WAKA←Kepsek) status=pending
- Notification: Created untuk WAKA
```

### 3.2 Use Case: Get Inbox dengan Pagination

```go
// 1. CONTROLLER
func (c *DisposisiController) GetInbox(ctx *gin.Context) {
    userID := ctx.GetUint("user_id")
    
    page := ctx.DefaultQuery("page", "1")
    pageSize := ctx.DefaultQuery("page_size", "20")
    
    inbox, err := c.service.GetInbox(userID, page, pageSize)
    if err != nil {
        ctx.JSON(500, dto.ErrorResponse{
            Code:    "INBOX_ERROR",
            Message: err.Error(),
        })
        return
    }
    
    ctx.JSON(200, inbox)
}

// 2. SERVICE - Prepare data
func (s *DisposisiServiceImpl) GetInbox(
    userID uint,
    page, pageSize int,
) (*dto.InboxListResponse, error) {
    // Get dari repository
    disposisi, total, err := s.repo.GetInbox(userID, page, pageSize)
    if err != nil {
        return nil, err
    }
    
    // Convert ke DTO
    items := make([]dto.InboxItemResponse, len(disposisi))
    for i, d := range disposisi {
        items[i] = dto.InboxItemResponse{
            ID:           d.ID,
            FromUser:     UserBasicInfo{d.FromUser.ID, d.FromUser.Name},
            Status:       d.Status,
            Dibaca:       d.Dibaca,
            SuratNomor:   d.SuratMasuk.NoSurat,
            SuratPerihal: d.SuratMasuk.PerihalSurat,
            CreatedAt:    d.CreatedAt,
        }
    }
    
    return &dto.InboxListResponse{
        Data: items,
        Pagination: dto.PaginationMeta{
            Page:       page,
            PageSize:   pageSize,
            TotalItems: int(total),
            TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
        },
    }, nil
}

// 3. REPOSITORY - Database query
func (r *DisposisiRepositoryImpl) GetInbox(
    userID uint,
    page, pageSize int,
) ([]models.Disposisi, int64, error) {
    var disposisi []models.Disposisi
    var total int64
    
    offset := (page - 1) * pageSize
    
    // Count total
    r.db.
        Where("to_user_id = ? AND deleted_at IS NULL", userID).
        Model(&models.Disposisi{}).
        Count(&total)
    
    // Get dengan pagination
    err := r.db.
        Where("to_user_id = ? AND deleted_at IS NULL", userID).
        Preload("FromUser").
        Preload("ToUser").
        Preload("SuratMasuk").
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&disposisi).Error
    
    return disposisi, total, err
}
```

**Response:**
```json
{
    "data": [
        {
            "id": 5,
            "from_user": {
                "id": 2,
                "name": "Kepala Sekolah",
                "email": "kepsek@school.id"
            },
            "status": "pending",
            "dibaca": false,
            "surat_nomor": "001/2024",
            "surat_perihal": "Rapat Kurikulum",
            "created_at": "2024-05-25T10:30:00Z"
        }
    ],
    "pagination": {
        "page": 1,
        "page_size": 20,
        "total_items": 150,
        "total_pages": 8
    }
}
```

### 3.3 Use Case: Get History (Full Trace)

```go
// 1. CONTROLLER
func (c *DisposisiController) GetHistory(ctx *gin.Context) {
    suratID := ctx.Param("surat_id")
    
    history, err := c.service.GetHistory(uint(suratID))
    if err != nil {
        ctx.JSON(400, dto.ErrorResponse{...})
        return
    }
    
    ctx.JSON(200, history)
}

// 2. SERVICE
func (s *DisposisiServiceImpl) GetHistory(suratID uint) (*dto.HistoryResponse, error) {
    // Get root disposisi
    root, err := s.repo.GetRootDisposisi(suratID)
    if err != nil {
        return nil, err
    }
    
    // Get full tree (dengan recursive preload)
    tree, err := s.repo.GetDisposisiTree(root.ID)
    if err != nil {
        return nil, err
    }
    
    // Get surat info
    var surat models.SuratMasuk
    s.db.First(&surat, suratID)
    
    // Convert tree ke DTO
    treeDTO := s.convertDisposisiToTreeNode(tree)
    
    return &dto.HistoryResponse{
        SuratMasukID:  suratID,
        SuratNomor:    surat.NoSurat,
        SuratPerihal:  surat.PerihalSurat,
        RootDisposisi: treeDTO,
        Status:        root.Status,
    }, nil
}

// 3. REPOSITORY - Recursive query
func (r *DisposisiRepositoryImpl) GetDisposisiTree(
    rootID uint,
) (*models.Disposisi, error) {
    var disposisi models.Disposisi
    
    err := r.db.
        Where("id = ?", rootID).
        Preload("FromUser").
        Preload("ToUser").
        Preload("ChildDisposisi", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted_at IS NULL")
        }).
        First(&disposisi).Error
    
    // Recursively load children
    r.loadDisposisiTreeRecursive(&disposisi, 0, 5)
    
    return &disposisi, err
}

func (r *DisposisiRepositoryImpl) loadDisposisiTreeRecursive(
    disposisi *models.Disposisi,
    level, maxLevel int,
) {
    if level >= maxLevel || disposisi == nil {
        return
    }
    
    for i := range disposisi.ChildDisposisi {
        child := &disposisi.ChildDisposisi[i]
        
        // Load relation untuk child
        r.db.
            Preload("FromUser").
            Preload("ToUser").
            Preload("ChildDisposisi", func(db *gorm.DB) *gorm.DB {
                return db.Where("deleted_at IS NULL")
            }).
            First(child)
        
        // Recursive
        r.loadDisposisiTreeRecursive(child, level+1, maxLevel)
    }
}
```

**Response (Tree Structure):**
```json
{
    "surat_masuk_id": 1,
    "surat_nomor": "001/2024",
    "root_disposisi": {
        "id": 1,
        "from_user": { "id": 1, "name": "Tata Usaha" },
        "to_user": { "id": 2, "name": "Kepala Sekolah" },
        "status": "forwarded",
        "created_at": "2024-05-25T08:00:00Z",
        "children": [
            {
                "id": 2,
                "from_user": { "id": 2, "name": "Kepala Sekolah" },
                "to_user": { "id": 3, "name": "WAKA Kurikulum" },
                "status": "forwarded",
                "created_at": "2024-05-25T09:00:00Z",
                "children": [
                    {
                        "id": 3,
                        "from_user": { "id": 3, "name": "WAKA Kurikulum" },
                        "to_user": { "id": 5, "name": "Guru Matematika" },
                        "status": "completed",
                        "complete_at": "2024-05-25T16:00:00Z",
                        "children": []
                    }
                ]
            }
        ]
    }
}
```

## 4. Implementation Checklist

### Phase 1: Core Models & Repositories
- [x] Update Disposisi model dengan parent-child support
- [x] Create DTO layer
- [x] Implement Repository pattern
- [ ] Create database migrations

### Phase 2: Services
- [x] Disposisi Service (forwarding logic)
- [x] Permission Service (role hierarchy)
- [x] Notification Service
- [ ] Test all services

### Phase 3: API Controllers (TODO)
- [ ] Create DisposisiController
- [ ] Implement endpoints:
  - POST   /disposisi/:id/forward
  - POST   /disposisi/:id/complete
  - POST   /disposisi/:id/reject
  - POST   /disposisi/read
  - GET    /disposisi/inbox
  - GET    /disposisi/sent
  - GET    /disposisi/:id
  - GET    /surat/:surat_id/history
  - GET    /disposisi/stats

### Phase 4: Integration & Testing
- [ ] Write unit tests
- [ ] Integration tests
- [ ] API documentation (Swagger/OpenAPI)

## 5. SQL Migration Template

```sql
-- Migrate existing disposisi table
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS parent_disposisi_id BIGINT;
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS level INT DEFAULT 0;
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS baca_at TIMESTAMP;
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS complete_at TIMESTAMP;
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Update column names for consistency
ALTER TABLE disposisi RENAME COLUMN id_disposisi TO id;
ALTER TABLE disposisi RENAME COLUMN id_surat_masuk TO surat_masuk_id;
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS from_user_id BIGINT;
ALTER TABLE disposisi ADD COLUMN IF NOT EXISTS to_user_id BIGINT;

-- Add indexes
CREATE INDEX idx_disposisi_to_user_id ON disposisi(to_user_id, deleted_at);
CREATE INDEX idx_disposisi_from_user_id ON disposisi(from_user_id, deleted_at);
CREATE INDEX idx_disposisi_surat_masuk_id ON disposisi(surat_masuk_id);
CREATE INDEX idx_disposisi_parent_id ON disposisi(parent_disposisi_id);
CREATE INDEX idx_disposisi_status ON disposisi(status, deleted_at);
CREATE INDEX idx_disposisi_created_at ON disposisi(created_at DESC);

-- Add foreign key constraints
ALTER TABLE disposisi 
ADD CONSTRAINT fk_disposisi_parent 
FOREIGN KEY (parent_disposisi_id) 
REFERENCES disposisi(id);
```

## 6. Next Steps

1. **Create Controllers** - Implement HTTP handlers dengan validation
2. **Add Routes** - Register semua disposisi endpoints
3. **Write Tests** - Unit tests untuk service & repository
4. **API Documentation** - Swagger/OpenAPI specs
5. **Frontend Integration** - Implement UI untuk forwarding, inbox, history
