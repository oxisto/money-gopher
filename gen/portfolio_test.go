package portfoliov1

import (
	"testing"
	"time"

	"github.com/oxisto/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestPortfolioEvent_MakeUniqueName(t *testing.T) {
	type fields struct {
		Name          string
		Type          PortfolioEventType
		Time          *timestamppb.Timestamp
		PortfolioName string
		SecurityName  string
		Amount        float64
		Price         *Currency
		Fees          *Currency
		Taxes         *Currency
	}
	tests := []struct {
		name   string
		fields fields
		want   assert.Want[*PortfolioEvent]
	}{
		{
			name: "happy path",
			fields: fields{
				SecurityName:  "stock",
				PortfolioName: "bank/myportfolio",
				Amount:        10,
				Type:          PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY,
				Time:          timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local)),
			},
			want: func(t *testing.T, tx *PortfolioEvent) bool {
				return assert.Equals(t, "71847445c432dde2", tx.Name)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &PortfolioEvent{
				Name:          tt.fields.Name,
				Type:          tt.fields.Type,
				Time:          tt.fields.Time,
				PortfolioName: tt.fields.PortfolioName,
				SecurityName:  tt.fields.SecurityName,
				Amount:        tt.fields.Amount,
				Price:         tt.fields.Price,
				Fees:          tt.fields.Fees,
				Taxes:         tt.fields.Taxes,
			}
			tx.MakeUniqueName()
			tt.want(t, tx)
		})
	}
}
