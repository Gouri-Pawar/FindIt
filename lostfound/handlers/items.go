package handlers

import (
	"net/http"

	"lostfound/database"
	"lostfound/models"

	"github.com/gin-gonic/gin"
)

type itemInput struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required"`
	Location    string `json:"location" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=lost found"`
	ImageURL    string `json:"image_url"`
}

// Pointers let us know which fields the user really sent
type updateInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Location    *string `json:"location"`
	ImageURL    *string `json:"image_url"`
	Status      *string `json:"status" binding:"omitempty,oneof=open claimed returned"`
}

// CreateItem: post a lost or found item
func CreateItem(c *gin.Context) {
	var input itemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.Item{
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Location:    input.Location,
		Type:        input.Type,
		ImageURL:    input.ImageURL,
		Status:      "open",
		UserID:      c.GetUint("userID"),
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// ListItems: show all items, newest first
func ListItems(c *gin.Context) {
	var items []models.Item

	if err := database.DB.Order("created_at desc").Limit(100).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetItem: show one item
func GetItem(c *gin.Context) {
	var item models.Item

	if err := database.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateItem: change an item (owner only)
func UpdateItem(c *gin.Context) {
	var item models.Item
	if err := database.DB.First(&item, c.Param("id")).Error; err != nil {
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
	if err := database.DB.First(&item, c.Param("id")).Error; err != nil {
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