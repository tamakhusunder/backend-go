package usermodel

type User struct {
	ID       string `bson:"_id,omitempty" json:"id,omitempty"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"password"` // hashed later
	Role     string `bson:"role" json:"role"`
}
