package handlers

import (
	"net/http"
	"strings"

	"lostfound/database"
	"lostfound/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type claimInput struct {
	Message string `json:"message" binding:"required,min=10"`
	Contact string `json:"contact" binding:"required"`
}

// CreateClaim: someone says "this is mine" (or "I found this")
func CreateClaim(c *gin.Context) {
	var item models.Item
	if err := database.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	userID := c.GetUint("userID")

	if item.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot claim your own post"})
		return
	}
	if item.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this item is no longer open"})
		return
	}

	var input claimInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "describe the item (at least 10 characters) and add your contact"})
		return
	}

	// One claim per person per item
	var existing models.Claim
	if err := database.DB.Where("item_id = ? AND claimer_id = ?", item.ID, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "you have already sent a claim for this item"})
		return
	}

	claim := models.Claim{
		ItemID:    item.ID,
		ClaimerID: userID,
		Message:   strings.TrimSpace(input.Message),
		Contact:   strings.TrimSpace(input.Contact),
		Status:    "pending",
	}
	if err := database.DB.Create(&claim).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save claim"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "claim sent"})
}

// ItemClaims: the poster sees the claims on their item
func ItemClaims(c *gin.Context) {
	var item models.Item
	if err := database.DB.First(&item, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if item.UserID != c.GetUint("userID") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the poster can see the claims"})
		return
	}

	var claims []models.Claim
	database.DB.Preload("Claimer").
		Where("item_id = ?", item.ID).
		Order("created_at desc").
		Find(&claims)

	for i := range claims {
		claims[i].ClaimerName = claims[i].Claimer.Name
		if claims[i].Status != "approved" {
			claims[i].Contact = "" // hidden until the poster approves
		}
	}

	c.JSON(http.StatusOK, claims)
}

// MyClaims: the claims I have sent
func MyClaims(c *gin.Context) {
	var claims []models.Claim
	database.DB.Preload("Item.User").
		Where("claimer_id = ?", c.GetUint("userID")).
		Order("created_at desc").
		Find(&claims)

	out := []gin.H{}
	for _, cl := range claims {
		if cl.Item.ID == 0 {
			continue // the post was deleted
		}
		row := gin.H{
			"id":         cl.ID,
			"item_id":    cl.ItemID,
			"item_title": cl.Item.Title,
			"item_type":  cl.Item.Type,
			"status":     cl.Status,
			"message":    cl.Message,
			"created_at": cl.CreatedAt,
		}
		// Contact details are shared only after approval
		if cl.Status == "approved" {
			row["poster_name"] = cl.Item.User.Name
			row["poster_contact"] = cl.Item.Contact
			row["kept_at"] = cl.Item.KeptAt
		}
		out = append(out, row)
	}

	c.JSON(http.StatusOK, out)
}

// DecideClaim: the poster approves or rejects a claim
func DecideClaim(c *gin.Context) {
	var claim models.Claim
	if err := database.DB.Preload("Item").First(&claim, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "claim not found"})
		return
	}
	if claim.Item.UserID != c.GetUint("userID") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the poster can decide this claim"})
		return
	}
	if claim.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this claim was already decided"})
		return
	}

	var input struct {
		Decision string `json:"decision" binding:"required,oneof=approve reject"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "decision must be approve or reject"})
		return
	}

	if input.Decision == "approve" && claim.Item.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this item is no longer open"})
		return
	}

	// All the changes below happen together, or none of them
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if input.Decision == "reject" {
			return tx.Model(&claim).Update("status", "rejected").Error
		}

		if err := tx.Model(&claim).Update("status", "approved").Error; err != nil {
			return err
		}
		// the item is now claimed
		if err := tx.Model(&models.Item{}).Where("id = ?", claim.ItemID).Update("status", "claimed").Error; err != nil {
			return err
		}
		// the other waiting claims are rejected
		return tx.Model(&models.Claim{}).
			Where("item_id = ? AND id <> ? AND status = ?", claim.ItemID, claim.ID, "pending").
			Update("status", "rejected").Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save the decision"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "claim " + input.Decision + "d"})
}
