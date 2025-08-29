package types

type User struct {
	UserId      string  `json:"userId"`
	Email       string  `json:"email"`
	DisplayName *string `json:"displayName"`
	TierId      string  `json:"tierId"`
	Subscribed  bool    `json:"subscribed"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}
