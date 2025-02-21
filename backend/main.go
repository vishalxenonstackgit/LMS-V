package main

import (
	"backend/database"
	"fmt"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func createLibrary(library string) (map[string]interface{}, string) {
	var lib database.Library
	result := database.DB.First(&lib, "name = ?", library)
	if result.Error != nil {
		newLib := database.Library{Name: library}
		result := database.DB.Create(&newLib)
		if result.Error != nil {
			err := result.Error.Error()
			return nil, err
		}
		return map[string]interface{}{"success": "Library created successfully", "lib_id": newLib.ID}, "nil"
	}
	return nil, "Library already exists"
}

func main() {
	r := gin.Default()
	database.ConnectDatabase()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	r.POST("add-book", func(c *gin.Context) {

		type CombinedData struct {
			Book  database.Book `json:"book" binding:"required"`
			Email string        `json:"email" binding:"required"`
		}

		var combinedData CombinedData

		if err := c.ShouldBindJSON(&combinedData); err != nil {
			c.JSON(500, gin.H{"error1": err.Error()})
			return
		}

		var user database.Users

		error1 := database.DB.First(&user, "email = ?", combinedData.Email)
		if error1.Error != nil {
			c.JSON(400, gin.H{"error2": "user does not exists"})
			return
		}

		if user.Role != "admin" {
			c.JSON(400, gin.H{"error3": "Current user is not an admin"})
			return
		}

		var result database.Book

		err := database.DB.First(&result, "isbn = ?", combinedData.Book.ISBN)
		if err.Error != nil {
			result := database.DB.Create(&combinedData.Book)
			if result.Error != nil {
				c.JSON(500, gin.H{"error4": result.Error.Error()})
				return
			}
			c.JSON(200, gin.H{"msg": "Book published successfully"})
			return
		}

		data := database.Book{
			ISBN:            combinedData.Book.ISBN,
			LibID:           combinedData.Book.LibID,
			Title:           combinedData.Book.Title,
			Authors:         combinedData.Book.Authors,
			Publisher:       combinedData.Book.Publisher,
			Version:         combinedData.Book.Version,
			TotalCopies:     result.TotalCopies + 1,
			AvailableCopies: result.AvailableCopies + 1,
		}
		res := database.DB.Where("isbn = ?", combinedData.Book.ISBN).Save(&data)
		if res.Error != nil {
			c.JSON(500, gin.H{"error5": res.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"msg": "Book Published Successfully"})

	})

	r.GET("/get-libraries", func(c *gin.Context) {
		var libraries []database.Library
		result := database.DB.Find(&libraries)
		if result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"libraries": libraries})
	})

	r.POST("/get-library-books", func(c *gin.Context) {
		var books []database.Book
		var libID map[string]interface{}
		if err := c.ShouldBindJSON(&libID); err != nil {
			c.JSON(500, gin.H{"error1": err.Error()})
			return
		}
		err := database.DB.Find(&books, "lib_id = ?", libID["lib_id"])
		if err.Error != nil {
			c.JSON(400, gin.H{"error": "There are no books in this library"})
			return
		}
		c.JSON(200, gin.H{"books": books})
	})

	r.GET("/get-request-events", func(c *gin.Context) {
		var events []database.ReaderRequestEvents
		if err := database.DB.Find(&events).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"msg": events})
	})

	r.POST("/user-issue-requests", func(c *gin.Context) {
		var my_requests []database.ReaderRequestEvents
		var reader_id map[string]interface{}

		if err := c.ShouldBindJSON(&reader_id); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		result := database.DB.Find(&my_requests, "user_id = ?", reader_id["reader_id"])
		if result.Error != nil {
			c.JSON(500, gin.H{"error": result.Error.Error()})
			return
		}
		fmt.Println(my_requests[0].RequestDate.Format("2006-01-02"))
		c.JSON(200, gin.H{"msg": my_requests})
	})

	r.POST("login/admin", func(c *gin.Context) {
		var email map[string]interface{}

		if err := c.ShouldBindJSON(&email); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var data database.Users

		err := database.DB.Find(&data, "email = ?", email["email"])
		if err.Error != nil {
			c.JSON(400, gin.H{"error": "No user exists with this email"})
			return
		}
		if data.Role != "admin" {
			c.JSON(400, gin.H{"error": "You are not an admin"})
			return
		}
		c.JSON(200, gin.H{"msg": "Login success"})
	})

	r.POST("/create-issue-request", func(c *gin.Context) {
		var request database.ReaderRequestEvents
		type Data struct {
			ISBN     int `json:"isbn" binding:"required"`
			ReaderID int `json:"reader_id" binding:"required"`
		}

		var data Data

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var result database.ReaderRequestEvents
		if err := database.DB.First(&result, "user_id = ? AND book_id = ?", data.ReaderID, data.ISBN).Error; err != nil {
			var book database.Book
			if err := database.DB.First(&book, "isbn = ?", data.ISBN).Error; err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			if book.AvailableCopies == 0 {
				c.JSON(400, gin.H{"error": "This Book is not available yet"})
				return
			}

			request.BookID = data.ISBN
			request.UserID = data.ReaderID
			request.RequestType = "issue"
			request.RequestDate = time.Now()

			if err := database.DB.Create(&request).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"msg": "Request Created Successfully"})
			return
		}
		c.JSON(400, gin.H{"error": "You had already requested for this book"})

	})

	r.POST("/create-reader", func(c *gin.Context) {
		var reader database.Users
		var result database.Users

		if err := c.ShouldBindJSON(&reader); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err := database.DB.First(&result, "email = ?", reader.Email)
		if err.Error != nil {
			result := database.DB.Create(&reader)
			if result.Error != nil {
				c.JSON(500, gin.H{"error": result.Error.Error()})
				return
			}
			c.JSON(200, gin.H{"msg": "Reader account created successfully"})
			return
		}
		c.JSON(400, gin.H{"error": "A user with same email already exists"})

	})

	r.POST("/create-owner", func(c *gin.Context) {
		var data map[string]interface{}
		var owner database.Users
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		result, err := createLibrary(data["library"].(string))

		if err == "nil" {
			owner = database.Users{
				Name:          data["name"].(string),
				Email:         data["email"].(string),
				ContactNumber: data["contact_number"].(string),
				Role:          data["role"].(string),
				LibID:         result["lib_id"].(uint),
			}

			var r database.Users

			err := database.DB.First(&r, "email = ?", owner.Email)
			if err.Error != nil {
				if err := database.DB.Create(&owner); err != nil {
					c.JSON(200, gin.H{"msg": "Owner created successfully"})
					return
				}
			}
			result := database.DB.Delete(&database.Library{}, result["lib_id"])
			if result.Error != nil {
				fmt.Print("Library deletion failed")
			}
			c.JSON(400, gin.H{"error": "Owner already exists with this email id"})

		} else {
			c.JSON(400, gin.H{"error": err})
		}

	})

	r.POST("/create-admin", func(c *gin.Context) {
		var admin database.Users
		var result database.Users

		if err := c.ShouldBindJSON(&admin); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err := database.DB.First(&result, "email = ?", admin.Email)
		if err.Error != nil {
			result := database.DB.Create(&admin)
			if result.Error != nil {
				c.JSON(500, gin.H{"error": result.Error.Error()})
				return
			}
			c.JSON(200, gin.H{"msg": "admin created successfully"})
			return
		}
		c.JSON(400, gin.H{"error": "A user with same email already exists"})
	})

	r.POST("/login", func(c *gin.Context) {
		var email map[string]interface{}
		var result database.Users

		if err := c.ShouldBindJSON(&email); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err := database.DB.First(&result, "email = ?", email["email"])
		if err.Error != nil {
			c.JSON(400, gin.H{"error": "User with this email does not exist"})
			return
		}
		if result.Role == "owner" {
			c.JSON(200, gin.H{"msg": "owner"})
			return
		}
		if result.Role == "admin" {
			c.JSON(200, gin.H{"msg": "admin"})
			return
		}
		if result.Role == "reader" {
			c.JSON(200, gin.H{"msg": "reader"})
			return
		}

	})

	r.POST("/get-owner-library", func(c *gin.Context) {
		var email map[string]interface{}
		var result database.Users
		var library database.Library

		if err := c.ShouldBindJSON(&email); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		err := database.DB.First(&result, "email = ?", email["email"])
		if err.Error != nil {
			c.JSON(400, gin.H{"error": "Invalid email"})
			return
		}
		data := database.DB.First(&library, "id = ?", result.LibID)
		if data.Error != nil {
			c.JSON(500, gin.H{"error": data.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"msg": library})

	})

	r.POST("/get-library-admins", func(c *gin.Context) {
		var library database.Library
		var admins []database.Users

		if err := c.ShouldBindJSON(&library); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		err := database.DB.Find(&admins, "lib_id = ? AND role = ?", library.ID, "admin")
		if err.Error != nil {
			c.JSON(400, gin.H{"error": err.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"msg": admins})
	})

	r.POST("/delete-admin", func(c *gin.Context) {
		var admin database.Users
		if err := c.ShouldBindJSON(&admin); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err := database.DB.Delete(&database.Users{}, "id = ?", admin.ID)
		if err.Error != nil {
			c.JSON(500, gin.H{"error": err.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"msg": "Admin deleted successfully"})
	})

	r.POST("/get-user-data", func(c *gin.Context) {
		var email map[string]interface{}
		var data database.Users
		if err := c.ShouldBindJSON(&email); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err := database.DB.First(&data, "email = ?", email["email"])
		if err.Error != nil {
			c.JSON(400, gin.H{"error": "Email not found"})
			return
		}
		c.JSON(200, gin.H{"msg": data})

	})

	r.POST("/remove-book", func(c *gin.Context) {
		var isbn map[string]interface{}
		if err := c.ShouldBindJSON(&isbn); err != nil {
			c.JSON(500, gin.H{"error1": err.Error()})
			return
		}

		var book database.Book
		err := database.DB.First(&book, "isbn = ?", isbn["isbn"])
		if err.Error != nil {
			c.JSON(500, gin.H{"error": err.Error.Error()})
			return
		}

		if book.TotalCopies == 0 || book.AvailableCopies == 0 {
			c.JSON(400, gin.H{"error": "Can't remove this book as total copies are already 0"})
			return
		}

		if book.AvailableCopies == book.TotalCopies {
			book.AvailableCopies = book.AvailableCopies - 1
			book.TotalCopies = book.TotalCopies - 1
			err := database.DB.Where("isbn = ?", book.ISBN).Save(&book)
			if err.Error != nil {
				c.JSON(500, gin.H{"error": "Unable to remove book"})
				return
			}
			c.JSON(200, gin.H{"msg": "1 copy of this book has been removed successfully"})
			return
		}
		book.TotalCopies = book.TotalCopies - 1
		res := database.DB.Where("isbn = ?", book.ISBN).Save(&book)
		if res.Error != nil {
			c.JSON(500, gin.H{"error": "Unable to remove book"})
			return
		}
		c.JSON(200, gin.H{"msg": "1 copy of this book has been removed successfully"})

	})

	// database.DB.Delete(&database.Book{}, "lib_id = ?", 1)
	// database.DB.Delete(&database.ReaderRequestEvents{}, "1==1")
	r.Run()
}
