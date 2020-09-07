package godaddyclient

// TxtRecord is POGO for json marshelling
type TxtRecord struct {
	Data string `json:"data"`
	Name string `json:"name"`
	Type string `json:"type"`
}
