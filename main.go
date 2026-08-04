// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shinchi-pmtech/ringi/infrastructure/persistence"
	"github.com/shinchi-pmtech/ringi/presentation/handler"
	"github.com/shinchi-pmtech/ringi/usecase"

	"github.com/shinchi-pmtech/ringi/domain/application"
	_ "modernc.org/sqlite" // SQLiteドライバ(純Go実装)
)

func main() {
	db, err := sql.Open("sqlite", "ringi.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 差し替えたのはこの1箇所。usecase 以下は一切変更なし
	repo, err := persistence.NewApplicationSQLiteRepository(db)
	if err != nil {
		log.Fatal(err)
	}

	approve := usecase.NewApproveApplication(repo)
	reject := usecase.NewRejectApplication(repo)
	resubmit := usecase.NewResubmitApplication(repo)
	_ = handler.NewApplicationHandler(approve)
	// ... HTTPサーバーの起動(記事では省略)

	// デモを繰り返し実行できるよう、前回のデータを消しておく
	if _, err := db.Exec(`DELETE FROM applications`); err != nil {
		log.Fatal(err)
	}

	// 動作確認:申請 → 提出 → 差戻し → 再提出 → 承認
	app := application.NewApplication("APP-001", "tanaka", "開発端末の購入")
	_ = app.Submit()
	if err := repo.Save(app); err != nil {
		log.Fatal(err)
	}
	fmt.Println("提出:", app.Status())

	if err := approve.Execute("APP-001", "tanaka"); err != nil {
		fmt.Println("自己承認:", err)
	}
	if err := reject.Execute("APP-001", "suzuki"); err == nil {
		printStatus(repo, "差戻し")
	}
	if err := resubmit.Execute("APP-001"); err == nil {
		printStatus(repo, "再提出")
	}
	if err := approve.Execute("APP-001", "suzuki"); err == nil {
		printStatus(repo, "承認")
	}

	// SQLiteファイルからの復元デモ。プロセスを再起動しても状態は残る
	restored, err := repo.FindByID("APP-001")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("復元:", restored.Status())

	// リポジトリを迂回してDBに不正な値を仕込むと、復元時に NewStatus が弾く
	if _, err := db.Exec(`UPDATE applications SET status = 'banana' WHERE id = 'APP-001'`); err != nil {
		log.Fatal(err)
	}
	if _, err := repo.FindByID("APP-001"); err != nil {
		fmt.Println("不正データ:", err)
	}
}

func printStatus(repo *persistence.ApplicationSQLiteRepository, label string) {
	app, err := repo.FindByID("APP-001")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: %s\n", label, app.Status())
}
