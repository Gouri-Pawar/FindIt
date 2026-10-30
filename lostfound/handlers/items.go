package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"lostfound/database"
	"lostfound/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Only these photo types are allowed
var imageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type itemInput struct {
	Title       string `form:"title" json:"title" binding:"required"`
	Description string `form:"description" json:"description"`
	Category    string `form:"category" json:"category" binding:"required"`
	Location    string `form:"location" json:"location" binding:"required"`
	KeptAt      string `form:"kept_at" json:"kept_at"`
	Contact     string `form:"contact" json:"contact"`
	Type        string `form:"type" json:"type" binding:"required,oneof=lost found"`
}

type updateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Location    *string `json:"location"`
	KeptAt      *string `json:"kept_at"`
	Contact     *string `json:"contact"`
	ImageURL    *string `json:"image_url"`
	Status      *string `json:"status" binding:"omitempty,oneof=open claimed returned"`
}

// readPhoto reads the uploaded photo and checks that it is a real image
func readPhoto(c *gin.Context) ([]byte, string, error) {
	file, err := c.FormFile("photo")
	if err != nil {
		return nil, "", nil // no photo sent, that is fine
	}

	if file.Size > 5<<20 {
		return nil, "", errors.New("photo must be smaller than 5 MB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, "", errors.New("could not read the photo")
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, 5<<20))
	if err != nil {
		return nil, "", errors.New("could not read the photo")
	}

	// Check the real file type, not just the file name
	ctype := http.DetectContentType(data)
	if _, ok := imageTypes[ctype]; !ok {
		return nil, "", errors.New("only JPG, PNG, GIF or WEBP photos are allowed")
	}

	return data, ctype, nil
}

// decorate adds extra info to each item for the person who is viewing
func decorate(c *gin.Context, items []models.Item) {
	viewer := c.GetUint("userID") // 0 when not logged in

	ids := make([]uint, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}

	myClaims := map[uint]string{}
	pending := map[uint]int{}

	if viewer != 0 && len(ids) > 0 {
		var mine []models.Claim
		database.DB.Where("item_id IN ? AND claimer_id = ?", ids, viewer).Find(&mine)
		for _, cl := range mine {
			myClaims[cl.ItemID] = cl.Status
		}

		var rows []struct {
			ItemID uint
			N      int
		}
		database.DB.Model(&models.Claim{}).
			Select("item_id, count(*) as n").
			Where("item_id IN ? AND status = ?", ids, "pending").
			Group("item_id").
			Scan(&rows)
		for _, r := range rows {
			pending[r.ItemID] = r.N
		}
	}

	for i := range items {
		it := &items[i]
		it.PosterName = it.User.Name
		it.HasContact = it.Contact != ""
		it.MyClaim = myClaims[it.ID]

		isOwner := viewer != 0 && it.UserID == viewer
		if isOwner {
			it.PendingClaims = pending[it.ID]
		}
		// Contact is private: only the poster and an approved claimer can see it
		if !isOwner && it.MyClaim != "approved" {
			it.Contact = ""
		}
	}
}

// CreateItem: post a lost or found item (with optional photo)
func CreateItem(c *gin.Context) {
	var input itemInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please fill all required fields"})
		return
	}

	data, ctype, err := readPhoto(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keptAt := input.KeptAt
	if input.Type == "lost" {
		keptAt = "" // "kept at" only makes sense for found items
	}

	item := models.Item{
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Location:    input.Location,
		KeptAt:      keptAt,
		Contact:     input.Contact,
		Type:        input.Type,
		ImageData:   data, // the photo is saved inside PostgreSQL
		ImageType:   ctype,
		Status:      "open",
		UserID:      c.GetUint("userID"),
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save item"})
		return
	}

	// The photo is served from our own API route
	if len(data) > 0 {
		url := fmt.Sprintf("/items/%d/image", item.ID)
		database.DB.Model(&item).Update("image_url", url)
	}

	item.ImageData = nil // no need to keep the bytes in memory
	database.DB.Omit("image_data").Preload("User").First(&item, item.ID)
	item.PosterName = item.User.Name

	c.JSON(http.StatusCreated, item)
}

// ItemImage sends the photo of one item
func ItemImage(c *gin.Context) {
	var item models.Item
	err := database.DB.Select("id", "image_data", "image_type").First(&item, c.Param("id")).Error
	if err != nil || len(item.ImageData) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, item.ImageType, item.ImageData)
}

// applyFilters adds the search conditions from the URL
func applyFilters(c *gin.Context, db *gorm.DB) *gorm.DB {
	if t := c.Query("type"); t != "" {
		db = db.Where("type = ?", t)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		db = db.Where("category ILIKE ?", category)
	}
	if location := strings.TrimSpace(c.Query("location")); location != "" {
		db = db.Where("location ILIKE ?", "%"+location+"%")
	}
	// Every word must match somewhere. Capital or small letters do not matter.
	for _, word := range strings.Fields(c.Query("q")) {
		like := "%" + word + "%"
		db = db.Where(
			"(title ILIKE ? OR description ILIKE ? OR category ILIKE ? OR location ILIKE ? OR kept_at ILIKE ?)",
			like, like, like, like, like,
		)
	}
	return db
}

// ListItems: show items with search, filters and pages
func ListItems(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var total int64
	applyFilters(c, database.DB.Model(&models.Item{})).Count(&total)

	var items []models.Item
	err = applyFilters(c, database.DB.Model(&models.Item{})).
		Omit("image_data"). // do not load the heavy photo bytes for the list
		Preload("User").
		Order("created_at desc").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&items).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load items"})
		return
	}

	decorate(c, items)

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// GetItem: show one item
func GetItem(c *gin.Context) {
	var item models.Item

	if err := database.DB.Omit("image_data").Preload("User").First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	items := []models.Item{item}
	decorate(c, items)
	c.JSON(http.StatusOK, items[0])
}

// UpdateItem: change an item (owner only)
func UpdateItem(c *gin.Context) {
	var item models.Item
	if err := database.DB.Omit("image_data").First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	if item.UserID != c.GetUint("userID") {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only edit your own items"})
		return
	}

	var input updateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Category != nil {
		updates["category"] = *input.Category
	}
	if input.Location != nil {
		updates["location"] = *input.Location
	}
	if input.KeptAt != nil {
		updates["kept_at"] = *input.KeptAt
	}
	if input.Contact != nil {
		updates["contact"] = *input.Contact
	}
	if input.ImageURL != nil {
		updates["image_url"] = *input.ImageURL
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
		return
	}

	if err := database.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteItem: remove an item (owner only)
func DeleteItem(c *gin.Context) {
	var item models.Item
	if err := database.DB.Select("id", "user_id").First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	if item.UserID != c.GetUint("userID") {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only delete your own items"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
}