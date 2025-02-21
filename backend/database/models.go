package database

import "time"

// Library schema
type Library struct {
	ID   uint   `json:"id" gorm:"primaryKey; autoIncrement"`
	Name string `json:"name" binding:"required"`
}

// Users schema
type Users struct {
	ID            uint   `json:"id" gorm:"primaryKey; autoIncrement"`
	Name          string `json:"name" binding:"required"`
	Email         string `json:"email" binding:"required"`
	ContactNumber string `json:"contact_number" binding:"required"`
	Role          string `json:"role" binding:"required"`
	LibID         uint   `json:"lib_id" binding:"required"`
}

// Books schema
type Book struct {
	ISBN            int    `json:"isbn" binding:"required"`
	LibID           uint   `json:"lib_id" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Authors         string `json:"authors" binding:"required"`
	Publisher       string `json:"publisher" binding:"required"`
	Version         string `json:"version" binding:"required"`
	TotalCopies     int    `json:"total_copies" binding:"required"`
	AvailableCopies int    `json:"available_copies" binding:"required"`
}

// Request Events schema

type ReaderRequestEvents struct {
	ReqID        uint       `json:"req_id" gorm:"primaryKey;autoIncrement"`
	BookID       int        `json:"book_id" binding:"required"`
	UserID       int        `json:"reader_id" binding:"required"`
	RequestDate  time.Time  `json:"request_date" binding:"required"`
	ApprovalDate *time.Time `json:"approval_date" binding:"required"`
	ApproverID   *int       `json:"approver_id" binding:"required"`
	RequestType  string     `json:"request_type" binding:"required"`
}

// IssueRegistery schema
type IssueRegistery struct {
	IssueID            uint      `json:"issue_id" gorm:"primaryKey; autoIncrement"`
	ISBN               int       `json:"isbn" binding:"required"`
	ReaderID           int       `json:"reader_id" column:"reader_id" binding:"required"`
	IssueApproverID    int       `json:"issue_approver_id" binding:"required"`
	IssueStatus        string    `json:"issue_status" binding:"required"`
	IssueDate          time.Time `json:"issue_date" binding:"required"`
	ExpectedReturnDate time.Time `json:"expected_return_date" binding:"required"`
	ReturnDate         time.Time `json:"return_date" binding:"required"`
	ReturnApproverID   int       `json:"return_approver_id" binding:"required"`
}
