package domain

type Workspace struct {
	Name  string `json:"name"`
	Units []Unit `json:"units"`
}
