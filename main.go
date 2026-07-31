// main.go
package main

import (
	"fmt"

	"github.com/shinchi-pmtech/ringi/domain/application"
	"github.com/shinchi-pmtech/ringi/infrastructure/persistence"
	"github.com/shinchi-pmtech/ringi/presentation/handler"
	"github.com/shinchi-pmtech/ringi/usecase"
)

func main() {
	repo := persistence.NewApplicationMemoryRepository()
	approve := usecase.NewApproveApplication(repo)
	reject := usecase.NewRejectApplication(repo)
	resubmit := usecase.NewResubmitApplication(repo)
	_ = handler.NewApplicationHandler(approve)
	// ... HTTPサーバーの起動(記事では省略)

	// 動作確認:申請 → 提出 → 差戻し → 再提出 → 承認
	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	_ = repo.Save(app)
	_ = app.Submit()
	fmt.Println("提出:", app.Status())

	if err := approve.Execute("APP-001", "tanaka"); err != nil {
		fmt.Println("自己承認:", err)
	}
	if err := reject.Execute("APP-001", "suzuki"); err == nil {
		fmt.Println("差戻し:", app.Status())
	}
	if err := app.Approve("suzuki"); err != nil {
		fmt.Println("差戻し中の承認:", err)
	}
	if err := resubmit.Execute("APP-001"); err == nil {
		fmt.Println("再提出:", app.Status())
	}
	if err := approve.Execute("APP-001", "suzuki"); err == nil {
		fmt.Println("承認:", app.Status())
	}
	if err := app.Resubmit(); err != nil {
		fmt.Println("承認後の再提出:", err)
	}

	// 自己検証のデモ:不正な値は復元時に弾かれる
	if _, err := application.NewStatus("banana"); err != nil {
		fmt.Println("復元:", err)
	}
}
