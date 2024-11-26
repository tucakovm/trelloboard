package domain

type User struct {
	Id       string `bson:"_id,omitempty" json:"id"`
	Username string `bson:"username" json:"username"`
	Role     string `bson:"role" json:"role"`
}
