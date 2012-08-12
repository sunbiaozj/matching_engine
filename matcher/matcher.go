package matcher

import (
	"fmt"
)

type M struct {
	buys, sells *heap
	stockId     uint32
}

func NewMatcher(stockId uint32) *M {
	buys := newHeap(BUY)
	sells := newHeap(SELL)
	return &M{buys: buys, sells: sells, stockId: stockId}
}

func (m *M) AddSell(s *Order) {
	if s.BuySell != SELL {
		panic("Added non-sell trade as a sell")
	}
	if s.StockId != m.stockId {
		panic(fmt.Sprintf("Added sell trade with stock-id %s expecting %s", s.StockId, m.stockId))
	}
	if !m.fillableSell(s) {
		m.sells.push(s)
	}
}

func (m *M) AddBuy(b *Order) {
	if b.BuySell != BUY {
		panic("Added non-buy trade as a buy")
	}
	if b.StockId != m.stockId {
		panic(fmt.Sprintf("Added buy trade with stock-id %s expecting %s", b.StockId, m.stockId))
	}
	if b.Price == MarketPrice {
		panic("It is illegal to submit a buy at market price")
	}
	if !m.fillableBuy(b) {
		m.buys.push(b)
	}
}

func (m *M) fillableBuy(b *Order) bool {
	for {
		s := m.sells.peek()
		if s == nil {
			return false
		}
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				m.sells.pop()
				b.Amount -= amount
				completeTrade(b, s, price, amount)
				continue
			}
			if s.Amount > b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				s.Amount -= amount
				completeTrade(b, s, price, amount)
				return true // The buy has been used up
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				completeTrade(b, s, price, amount)
				m.sells.pop()
				return true // The buy has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func (m *M) fillableSell(s *Order) bool {
	for {
		b := m.buys.peek()
		if b == nil {
			return false
		}
		if b.Price >= s.Price {
			if b.Amount > s.Amount {
				amount := s.Amount
				price := price(b.Price, s.Price)
				m.sells.pop()
				b.Amount -= amount
				completeTrade(b, s, price, amount)
				return true // The sell has been used up
			}
			if s.Amount >  b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				s.Amount -= amount
				completeTrade(b, s, price, amount)
				m.sells.pop()
				continue
			}
			if s.Amount == b.Amount {
				amount := b.Amount
				price := price(b.Price, s.Price)
				completeTrade(b, s, price, amount)
				m.sells.pop()
				return true // The sell has been used up
			}
		} else {
			return false
		}
	}
	panic("Unreachable")
}

func price(bPrice, sPrice int64) int64 {
	if sPrice == MarketPrice {
		return bPrice
	}
	d := bPrice - sPrice
	return sPrice + (d >> 1)
}

func completeTrade(b, s *Order, price int64, amount uint32) {
	b.ResponseFunc(NewResponse(-price, amount, b.TradeId, s.TraderId))
	s.ResponseFunc(NewResponse(price, amount, s.TradeId, b.TraderId))
}
