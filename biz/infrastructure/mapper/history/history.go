package history

import "time"
import "go.mongodb.org/mongo-driver/bson/primitive"

type History struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Class     string             `bson:"class" json:"class"`
	Dialogs   []*Dialog          `bson:"dialogs" json:"dialogs"`
	Report    *Report            `bson:"report" json:"report"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	EndTime   time.Time          `bson:"end_time" json:"end_time"`
}

type Dialog struct {
	Role    string `bson:"role" json:"role"`
	Content string `bson:"content" json:"content"`
}

type Report struct {
	Keywords   []string `bson:"keywords" json:"keywords"`
	Type       []string `bson:"type" json:"type"`
	Content    string   `bson:"content" json:"content"`
	Grade      string   `bson:"grade" json:"grade"`
	Suggestion []string `bson:"suggestion" json:"suggestion"`
}
