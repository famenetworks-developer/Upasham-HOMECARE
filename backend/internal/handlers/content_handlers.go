package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ContentHandler struct{ db *sql.DB }

func NewContentHandler(db *sql.DB) *ContentHandler { return &ContentHandler{db: db} }

type contentDocument struct {
	Content           json.RawMessage `json:"content"`
	DraftRevision     int             `json:"draft_revision"`
	PublishedRevision int             `json:"published_revision"`
	UpdatedAt         *time.Time      `json:"updated_at"`
	PublishedAt       *time.Time      `json:"published_at"`
}

func (h *ContentHandler) GetPublished(c *gin.Context) {
	var content []byte
	var revision int
	var publishedAt sql.NullTime
	err := h.db.QueryRow(`SELECT published_content, published_revision, published_at FROM site_content WHERE id='primary'`).Scan(&content, &revision, &publishedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load site content"})
		return
	}
	var published *time.Time
	if publishedAt.Valid {
		published = &publishedAt.Time
	}
	c.JSON(http.StatusOK, gin.H{"content": json.RawMessage(content), "revision": revision, "published_at": published})
}

func (h *ContentHandler) GetAdmin(c *gin.Context) {
	var d contentDocument
	var updatedAt, publishedAt sql.NullTime
	err := h.db.QueryRow(`SELECT draft_content, draft_revision, published_revision, updated_at, published_at FROM site_content WHERE id='primary'`).Scan(&d.Content, &d.DraftRevision, &d.PublishedRevision, &updatedAt, &publishedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to load content workspace"})
		return
	}
	if updatedAt.Valid {
		d.UpdatedAt = &updatedAt.Time
	}
	if publishedAt.Valid {
		d.PublishedAt = &publishedAt.Time
	}
	c.JSON(http.StatusOK, d)
}

func readContent(c *gin.Context) (json.RawMessage, error) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	var payload struct {
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || len(payload.Content) == 0 || !json.Valid(payload.Content) {
		return nil, fmt.Errorf("content must be valid JSON")
	}
	var shape map[string]interface{}
	if err := json.Unmarshal(payload.Content, &shape); err != nil || len(shape) == 0 {
		return nil, fmt.Errorf("content must be a non-empty object")
	}
	return payload.Content, nil
}

func (h *ContentHandler) SaveDraft(c *gin.Context) {
	content, err := readContent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid := c.GetString("user_id")
	var revision int
	err = h.db.QueryRow(`UPDATE site_content SET draft_content=$1, draft_revision=draft_revision+1, updated_by=$2, updated_at=NOW() WHERE id='primary' RETURNING draft_revision`, content, uid).Scan(&revision)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save draft"})
		return
	}
	h.auditContent(c, "SAVE_CONTENT_DRAFT", revision)
	c.JSON(http.StatusOK, gin.H{"message": "Draft saved", "draft_revision": revision})
}

func (h *ContentHandler) Publish(c *gin.Context) {
	uid := c.GetString("user_id")
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to publish"})
		return
	}
	defer tx.Rollback()
	var content []byte
	var revision int
	err = tx.QueryRow(`UPDATE site_content SET published_content=draft_content, published_revision=draft_revision, published_at=NOW(), updated_by=$1 WHERE id='primary' RETURNING published_content, published_revision`, uid).Scan(&content, &revision)
	if err == nil {
		_, err = tx.Exec(`INSERT INTO site_content_revisions (id,revision,content,status,created_by) VALUES ($1,$2,$3,'PUBLISHED',$4)`, fmt.Sprintf("content-%d", time.Now().UnixNano()), revision, content, uid)
	}
	if err != nil || tx.Commit() != nil {
		c.JSON(500, gin.H{"error": "Unable to publish content"})
		return
	}
	h.auditContent(c, "PUBLISH_SITE_CONTENT", revision)
	c.JSON(http.StatusOK, gin.H{"message": "Website published", "published_revision": revision})
}

func (h *ContentHandler) ListRevisions(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id,revision,status,created_by,created_at FROM site_content_revisions ORDER BY created_at DESC LIMIT 30`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to load history"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, status string
		var revision int
		var by sql.NullString
		var at time.Time
		if rows.Scan(&id, &revision, &status, &by, &at) == nil {
			items = append(items, gin.H{"id": id, "revision": revision, "status": status, "created_by": by.String, "created_at": at})
		}
	}
	c.JSON(200, items)
}

func (h *ContentHandler) RestoreRevision(c *gin.Context) {
	var content []byte
	var oldRevision int
	if h.db.QueryRow(`SELECT content,revision FROM site_content_revisions WHERE id=$1`, c.Param("id")).Scan(&content, &oldRevision) != nil {
		c.JSON(404, gin.H{"error": "Revision not found"})
		return
	}
	var revision int
	if h.db.QueryRow(`UPDATE site_content SET draft_content=$1,draft_revision=draft_revision+1,updated_by=$2,updated_at=NOW() WHERE id='primary' RETURNING draft_revision`, content, c.GetString("user_id")).Scan(&revision) != nil {
		c.JSON(500, gin.H{"error": "Unable to restore revision"})
		return
	}
	h.auditContent(c, "RESTORE_CONTENT_REVISION", revision)
	c.JSON(200, gin.H{"message": fmt.Sprintf("Revision %d restored as a new draft", oldRevision), "draft_revision": revision})
}

func (h *ContentHandler) auditContent(c *gin.Context, action string, revision int) {
	_, _ = h.db.Exec(`INSERT INTO audit_events (id,user_id,action,entity_type,entity_id,ip_address,details_json) VALUES ($1,$2,$3,'SiteContent','primary',$4,$5)`, fmt.Sprintf("aud-%d", time.Now().UnixNano()), c.GetString("user_id"), action, c.ClientIP(), fmt.Sprintf(`{"revision":%d}`, revision))
}
