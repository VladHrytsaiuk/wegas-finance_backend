package models

type MedicalRecord struct {
	Base
	FamilyID string `json:"family_id" gorm:"index"`
	UserID   string `json:"user_id" gorm:"index"`

	Date       int64  `json:"date"`
	Type       string `json:"type"` // doctor, analysis, medicine
	Title      string `json:"title"`
	
	DoctorName string `json:"doctor_name"`
	ClinicName string `json:"clinic_name"`
	
	Diagnosis  string `json:"diagnosis"`
	Treatment  string `json:"treatment"`
	
	Cost     int64  `json:"cost"`
	Currency string `json:"currency"`

	Files []MedicalFile `json:"files" gorm:"foreignKey:RecordID;constraint:OnDelete:CASCADE;"`
}

type MedicalFile struct {
	Base
	RecordID string `json:"record_id" gorm:"index"`
	FilePath string `json:"file_path"`
	FileType string `json:"file_type"`
	Name     string `json:"name"`
}