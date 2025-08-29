package types

import "time"

type QuotaHistory struct {
	UserId               string    `firestore:"userId" json:"userId"`
	TierId               string    `firestore:"tierId" json:"tierId"`
	LastSubscriptionDate time.Time `firestore:"lastSubscriptionDate" json:"lastSubscriptionDate"`
	Features             Feature   `firestore:"features" json:"features"`
	CreatedAt            time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time `firestore:"updatedAt" json:"updatedAt"`
}
