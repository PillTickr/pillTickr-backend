// handlers/medicine.go
package handlers

import (
	"net/http"
	"pillTickr-backend/db"
	"pillTickr-backend/utils"

	"github.com/gin-gonic/gin"
)

// GET /medicines
func GetMedicines(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	rows, err := db.DB.Query(`SELECT medicine_id, name, description, dosage, instructions, created_at 
		FROM medicines WHERE user_id = ?`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch medicines"})
		return
	}
	defer rows.Close()

	medicines := []gin.H{}
	for rows.Next() {
		var m struct {
			ID           int    `json:"medicine_id"`
			Name         string `json:"name"`
			Description  string `json:"description"`
			Dosage       string `json:"dosage"`
			Instructions string `json:"instructions"`
			CreatedAt    string `json:"created_at"`
		}
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Dosage, &m.Instructions, &m.CreatedAt); err == nil {
			medicines = append(medicines, gin.H{
				"id": m.ID, "name": m.Name, "description": m.Description,
				"dosage": m.Dosage, "instructions": m.Instructions, "created_at": m.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, medicines)
}

// POST /medicines
func CreateMedicine(c *gin.Context) {
	userID, ok := utils.GetUserID(c)
	if !ok {
		return
	}

	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		Dosage       string `json:"dosage"`
		Instructions string `json:"instructions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := db.DB.Exec(`INSERT INTO medicines (user_id, name, description, dosage, instructions)
		VALUES (?, ?, ?, ?, ?)`, userID, req.Name, req.Description, req.Dosage, req.Instructions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create medicine", "error": err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"medicine_id": id})
}
