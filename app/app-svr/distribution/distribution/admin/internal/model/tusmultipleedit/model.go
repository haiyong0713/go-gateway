package tusmultipleedit

type Overview struct {
	Name         string            `json:"name"`
	FieldName    string            `json:"field_name"`
	Performances []*TusPerformance `json:"performances"`
}

type TusPerformance struct {
	TusValue string `json:"tus_value"`
	Text     string `json:"text"`
}
