package domain

type Task struct {
	//Id          uuid.UUID `json:"id,omitempty" bson:"id"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Status      Status `json:"status" bson:"status"`
	ProjectID   string `json:"project_id" bson:"project_id"`
}

type TasksRepository interface {
	Create(task Task) (Task, error)
}
