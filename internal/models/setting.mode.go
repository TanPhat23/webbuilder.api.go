package models

type Setting struct {
	Id 		string `json:"id"`
	Name 	string `json:"name"`
	Settingstype string `json:"settingstype"`
	Settings map[string]any `json:"settings"`
	ElementId string `json:"elementId"`
}
