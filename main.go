// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shinchi-pmtech/ringi/infrastructure/organization"
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
	resolver := organization.NewMemoryApproverResolver()

	create := usecase.NewCreateApplication(repo, resolver)
	approve := usecase.NewApproveApplication(repo)
	_ = handler.NewApplicationHandler(approve)
	// ... HTTPサーバーの起動(記事では省略)

	// 金額によって承認ルートが変わる
	if err := create.Execute("APP-001", "tanaka", "マウス購入", 5000); err != nil {
		log.Fatal(err)
	}
	printRoute(repo, "APP-001", "マウス購入")

	if err := create.Execute("APP-002", "tanaka", "開発端末の購入", 150000); err != nil {
		log.Fatal(err)
	}
	printRoute(repo, "APP-002", "開発端末の購入")

	// 10万円未満の申請は、課長の承認だけで承認済みになる
	fmt.Println()
	app, err := repo.FindByID("APP-001")
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Submit(); err != nil {
		log.Fatal(err)
	}
	if err := repo.Save(app); err != nil {
		log.Fatal(err)
	}
	if err := approve.Execute("APP-001", "kacho"); err != nil {
		log.Fatal(err)
	}
	printRoute(repo, "APP-001", "マウス購入")

	// 組織図に情報がない申請者はルートを決められない
	err = create.Execute("APP-003", "unknown", "備品購入", 3000)
	fmt.Println("\n組織図にない申請者:", err)
}

// printRoute は申請をDBから読み直して、金額・状態・承認ルートを表示する
func printRoute(repo *persistence.ApplicationSQLiteRepository, id, label string) {
	app, err := repo.FindByID(application.ApplicationID(id))
	if err != nil {
		log.Fatal(err)
	}

	route := ""
	for i, s := range app.Steps() {
		if i > 0 {
			route += " → "
		}
		mark := "○"
		if s.Approved() {
			mark = "●"
		}
		route += mark + string(s.ApproverID())
	}

	fmt.Printf("%s %s %s %s\n",
		pad(label, 18), pad(comma(app.Amount().Yen())+"円", 12), pad(string(app.Status()), 10), route)
}

// pad は全角文字を2桁として数えて右側を空白で埋める
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

// comma は数値を3桁区切りの文字列にする
func comma(n int) string {
	s := fmt.Sprintf("%d", n)
	out := ""
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(r)
	}
	return out
}
