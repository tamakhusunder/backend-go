package model

type User struct {
	ID        string `bson:"_id,omitempty" json:"id,omitempty"`
	Email     string `bson:"email" json:"email"`
	Password  string `bson:"password" json:"password"`
	Role      string `bson:"role" json:"role"`
	Token     string `bson:"token" json:"token"`
	IPAddress string `bson:"ip_address"`
}
