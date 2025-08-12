package subscription

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
