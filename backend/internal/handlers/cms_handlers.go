package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CMSHandler struct{ db *sql.DB }

func NewCMSHandler(db *sql.DB) *CMSHandler { return &CMSHandler{db: db} }

var safeType = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,40}$`)

type entryInput struct {
	Title     string          `json:"title"`
	Slug      string          `json:"slug"`
	Status    string          `json:"status"`
	SortOrder int             `json:"sort_order"`
	Data      json.RawMessage `json:"data"`
}

func (h *CMSHandler) Overview(c *gin.Context) {
	counts := map[string]int{}
	rows, _ := h.db.Query(`SELECT content_type,COUNT(*) FROM cms_entries WHERE deleted_at IS NULL GROUP BY content_type`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var k string
			var n int
			_ = rows.Scan(&k, &n)
			counts[k] = n
		}
	}
	var media, users, doctors int
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM media_assets WHERE deleted_at IS NULL`).Scan(&media)
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&users)
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM doctor_profiles`).Scan(&doctors)
	c.JSON(200, gin.H{"entries": counts, "media": media, "users": users, "doctors": doctors})
}
func (h *CMSHandler) ListEntries(c *gin.Context) {
	t := c.Param("type")
	if !safeType.MatchString(t) {
		c.JSON(400, gin.H{"error": "Invalid content type"})
		return
	}
	q := "%" + strings.TrimSpace(c.Query("q")) + "%"
	status := c.Query("status")
	rows, err := h.db.Query(`SELECT id,title,slug,status,data,sort_order,created_at,updated_at FROM cms_entries WHERE content_type=$1 AND deleted_at IS NULL AND (title ILIKE $2 OR slug ILIKE $2) AND ($3='' OR status=$3) ORDER BY sort_order,updated_at DESC`, t, q, status)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to load content"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, title, slug, status string
		var data []byte
		var order int
		var created, updated time.Time
		if rows.Scan(&id, &title, &slug, &status, &data, &order, &created, &updated) == nil {
			items = append(items, gin.H{"id": id, "title": title, "slug": slug, "status": status, "data": json.RawMessage(data), "sort_order": order, "created_at": created, "updated_at": updated})
		}
	}
	c.JSON(200, items)
}
func validEntry(i entryInput) error {
	if strings.TrimSpace(i.Title) == "" || strings.TrimSpace(i.Slug) == "" {
		return fmt.Errorf("title and slug are required")
	}
	if i.Status != "DRAFT" && i.Status != "PUBLISHED" && i.Status != "ARCHIVED" {
		return fmt.Errorf("invalid status")
	}
	if len(i.Data) == 0 {
		i.Data = json.RawMessage(`{}`)
	}
	if !json.Valid(i.Data) {
		return fmt.Errorf("data must be valid JSON")
	}
	return nil
}
func (h *CMSHandler) CreateEntry(c *gin.Context) {
	t := c.Param("type")
	var i entryInput
	if !safeType.MatchString(t) || c.ShouldBindJSON(&i) != nil || validEntry(i) != nil {
		c.JSON(400, gin.H{"error": "Valid title, slug, status and data are required"})
		return
	}
	id := fmt.Sprintf("cms-%d", time.Now().UnixNano())
	_, err := h.db.Exec(`INSERT INTO cms_entries(id,content_type,title,slug,status,data,sort_order,created_by,updated_by)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$8)`, id, t, i.Title, i.Slug, i.Status, i.Data, i.SortOrder, c.GetString("user_id"))
	if err != nil {
		c.JSON(409, gin.H{"error": "Slug already exists for this content type"})
		return
	}
	h.audit(c, "CREATE_CMS_ENTRY", id, t)
	c.JSON(201, gin.H{"id": id, "message": "Content created"})
}
func (h *CMSHandler) UpdateEntry(c *gin.Context) {
	var i entryInput
	if c.ShouldBindJSON(&i) != nil || validEntry(i) != nil {
		c.JSON(400, gin.H{"error": "Invalid content"})
		return
	}
	res, err := h.db.Exec(`UPDATE cms_entries SET title=$1,slug=$2,status=$3,data=$4,sort_order=$5,updated_by=$6,updated_at=NOW() WHERE id=$7 AND deleted_at IS NULL`, i.Title, i.Slug, i.Status, i.Data, i.SortOrder, c.GetString("user_id"), c.Param("id"))
	if err != nil {
		c.JSON(409, gin.H{"error": "Unable to update; check the slug"})
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(404, gin.H{"error": "Content not found"})
		return
	}
	h.audit(c, "UPDATE_CMS_ENTRY", c.Param("id"), c.Param("type"))
	c.JSON(200, gin.H{"message": "Content updated"})
}
func (h *CMSHandler) DeleteEntry(c *gin.Context) {
	res, _ := h.db.Exec(`UPDATE cms_entries SET deleted_at=NOW(),updated_by=$1 WHERE id=$2 AND deleted_at IS NULL`, c.GetString("user_id"), c.Param("id"))
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(404, gin.H{"error": "Content not found"})
		return
	}
	h.audit(c, "DELETE_CMS_ENTRY", c.Param("id"), c.Param("type"))
	c.JSON(200, gin.H{"message": "Moved to archive"})
}
func (h *CMSHandler) ListMedia(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id,file_name,url,mime_type,size_bytes,alt_text,created_at FROM media_assets WHERE deleted_at IS NULL ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to load media"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, name, url, mime, alt string
		var size int64
		var at time.Time
		if rows.Scan(&id, &name, &url, &mime, &size, &alt, &at) == nil {
			items = append(items, gin.H{"id": id, "file_name": name, "url": url, "mime_type": mime, "size_bytes": size, "alt_text": alt, "created_at": at})
		}
	}
	c.JSON(200, items)
}
func (h *CMSHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil || file.Size > 10<<20 {
		c.JSON(400, gin.H{"error": "Choose an image up to 10 MB"})
		return
	}
	opened, _ := file.Open()
	defer opened.Close()
	buf := make([]byte, 512)
	n, _ := opened.Read(buf)
	mime := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(mime, "image/") {
		c.JSON(400, gin.H{"error": "Only image files are allowed"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	id := fmt.Sprintf("media-%d", time.Now().UnixNano())
	storage := id + ext
	if err = os.MkdirAll("uploads", 0755); err != nil {
		c.JSON(500, gin.H{"error": "Storage unavailable"})
		return
	}
	if err = c.SaveUploadedFile(file, filepath.Join("uploads", storage)); err != nil {
		c.JSON(500, gin.H{"error": "Upload failed"})
		return
	}
	url := "/uploads/" + storage
	_, err = h.db.Exec(`INSERT INTO media_assets(id,file_name,storage_name,url,mime_type,size_bytes,alt_text,uploaded_by)VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, file.Filename, storage, url, mime, file.Size, c.PostForm("alt_text"), c.GetString("user_id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to register media"})
		return
	}
	h.audit(c, "UPLOAD_MEDIA", id, "media")
	c.JSON(201, gin.H{"id": id, "url": url, "message": "Image uploaded"})
}
func (h *CMSHandler) DeleteMedia(c *gin.Context) {
	res, _ := h.db.Exec(`UPDATE media_assets SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, c.Param("id"))
	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(404, gin.H{"error": "Media not found"})
		return
	}
	h.audit(c, "DELETE_MEDIA", c.Param("id"), "media")
	c.JSON(200, gin.H{"message": "Media archived"})
}
func (h *CMSHandler) ListUsers(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id,email,phone,full_name,role,is_active,created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to load users"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, email, phone, name, role string
		var active bool
		var at time.Time
		if rows.Scan(&id, &email, &phone, &name, &role, &active, &at) == nil {
			items = append(items, gin.H{"id": id, "email": email, "phone": phone, "full_name": name, "role": role, "is_active": active, "created_at": at})
		}
	}
	c.JSON(200, items)
}
func (h *CMSHandler) audit(c *gin.Context, action, id, t string) {
	_, _ = h.db.Exec(`INSERT INTO audit_events(id,user_id,action,entity_type,entity_id,ip_address,details_json)VALUES($1,$2,$3,'CMS',$4,$5,$6)`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), c.GetString("user_id"), action, id, c.ClientIP(), fmt.Sprintf(`{"content_type":%q}`, t))
}
