package model

// Company struct
type Company struct {
	Model
	Name  string `json:"name"`
	TaxId string `json:"tax_id"`
}
