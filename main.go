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
	_ = handler.NewApplicationHandler(approve)
	// ... HTTPサーバーの起動(記事では省略)

	// 動作確認:申請 → 提出 → 自己承認(エラー)→ 別人が承認(成功)
	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	_ = repo.Save(app)
	_ = app.Submit()

	if err := approve.Execute("APP-001", "tanaka"); err != nil {
		fmt.Println("自己承認:", err)
	}
	if err := approve.Execute("APP-001", "suzuki"); err != nil {
		fmt.Println("承認失敗:", err)
	} else {
		fmt.Println("承認成功。状態:", app.Status())
	}
}
