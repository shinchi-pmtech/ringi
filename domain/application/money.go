// domain/application/money.go
package application

import (
	"errors"
	"fmt"
)

var ErrNegativeAmount = errors.New("金額に負の値は指定できません")

// Money は日本円の金額を表す値オブジェクト。
// 円未満の端数は扱わないので、整数で保持する
type Money struct {
	yen int
}

// NewMoney は金額を生成する。負の値は受け付けない
func NewMoney(yen int) (Money, error) {
	if yen < 0 {
		return Money{}, fmt.Errorf("%w: %d", ErrNegativeAmount, yen)
	}
	return Money{yen: yen}, nil
}

func (m Money) Yen() int { return m.yen }

// GreaterThanOrEqual は m が other 以上かを返す
func (m Money) GreaterThanOrEqual(other Money) bool {
	return m.yen >= other.yen
}

func (m Money) String() string {
	return fmt.Sprintf("%d円", m.yen)
}
