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
	for _, table := range []string{"approval_steps", "applications"} {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			log.Fatal(err)
		}
	}

	// 「課長 → 部長」の2段承認で申請を作る
	app, err := application.NewApplication("APP-001", "tanaka", "開発端末の購入",
		application.ApprovalRoute{"kacho", "bucho"})
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Submit(); err != nil {
		log.Fatal(err)
	}
	if err := repo.Save(app); err != nil {
		log.Fatal(err)
	}
	printProgress(repo, "提出")

	// 順番を飛ばした承認は拒否される
	if err := approve.Execute("APP-001", "bucho"); err != nil {
		fmt.Println("部長が先に承認:", err)
	}

	// 1段目を承認しても、申請そのものはまだ「申請中」
	if err := approve.Execute("APP-001", "kacho"); err != nil {
		log.Fatal(err)
	}
	printProgress(repo, "課長が承認")

	// 差戻すと、再提出時に進捗はリセットされる
	if err := reject.Execute("APP-001", "bucho"); err != nil {
		log.Fatal(err)
	}
	printProgress(repo, "部長が差戻し")

	if err := resubmit.Execute("APP-001"); err != nil {
		log.Fatal(err)
	}
	printProgress(repo, "再提出")

	// 今度は最後まで承認する
	if err := approve.Execute("APP-001", "kacho"); err != nil {
		log.Fatal(err)
	}
	if err := approve.Execute("APP-001", "bucho"); err != nil {
		log.Fatal(err)
	}
	printProgress(repo, "全段承認")
}

// printProgress は申請をDBから読み直して、状態と承認の進捗を1行で表示する。
//
//	● 承認済み / ○ 未承認 / [ ] 次に承認する段
func printProgress(repo *persistence.ApplicationSQLiteRepository, label string) {
	app, err := repo.FindByID("APP-001")
	if err != nil {
		log.Fatal(err)
	}

	steps := app.Steps()
	route := ""
	for i, s := range steps {
		if i > 0 {
			route += " → "
		}

		mark := "○"
		if s.Approved() {
			mark = "●"
		}
		cell := fmt.Sprintf("%s%s", mark, s.ApproverID())

		// 申請中のときだけ、次に承認する段を [ ] で囲む
		isNext := !s.Approved() && (i == 0 || steps[i-1].Approved())
		if isNext && app.Status() == application.StatusSubmitted {
			cell = "[" + cell + "]"
		}
		route += cell
	}

	fmt.Printf("%s %s %s\n", pad(label, 14), pad(string(app.Status()), 10), route)
}

// pad は全角文字を2桁として数えて右側を空白で埋める。
// Printf の %-12s は全角を1桁と数えるため、日本語ラベルでは桁が揃わない
func pad(s string, width int) string {
	w := 0
	for _, r := range s {
		if r > 0x7f {
			w += 2
		} else {
			w++
		}
	}
	for ; w < width; w++ {
		s += " "
	}
	return s
}
