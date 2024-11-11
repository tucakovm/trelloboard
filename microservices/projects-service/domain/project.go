package domain

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"time"
)

type Project struct {
	Id             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	CompletionDate time.Time          `bson:"completionDate" json:"completionDate"`
	MinMembers     int32              `bson:"minMembers" json:"minMembers"`
	MaxMembers     int32              `bson:"maxMembers" json:"maxMembers"`
	Manager        User               `bson:"manager" json:"manager"`
	//Members        []User
}

type Projects []*Project

func (p *Project) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Project) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}
