package benchmarks

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type benchRow struct {
	ID      string `gorm:"column:id;primaryKey;type:text"`
	Payload string `gorm:"column:payload;type:text"`
}

func (benchRow) TableName() string { return "uuid_benchmark_rows" }

func BenchmarkUUIDGeneration(b *testing.B) {
	for _, v := range []int{4, 7} {
		b.Run(fmt.Sprint(v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var err error
				if v == 4 {
					_, err = uuid.NewRandom()
				} else {
					_, err = uuid.NewV7()
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Run only against a disposable database: this creates and drops its table.
func BenchmarkPostgresUUID(b *testing.B) {
	dsn := os.Getenv("UUID_BENCH_DSN")
	if dsn == "" {
		b.Skip("set UUID_BENCH_DSN to a disposable PostgreSQL database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: false, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		b.Fatal(err)
	}
	defer sqlDB.Close()
	const count = 100000
	for _, v := range []int{4, 7} {
		b.Run(fmt.Sprint(v), func(b *testing.B) {
			b.StopTimer()
			rows := make([]benchRow, count)
			for i := range rows {
				var id uuid.UUID
				if v == 4 {
					id, err = uuid.NewRandom()
				} else {
					id, err = uuid.NewV7()
				}
				if err != nil {
					b.Fatal(err)
				}
				rows[i] = benchRow{ID: id.String(), Payload: "representative payload"}
			}
			indices := rand.New(rand.NewSource(42)).Perm(count)[:1000]
			var totalInsert, totalLookup time.Duration
			var totalIndex int64
			for n := 0; n < b.N; n++ {
				table := "uuid_bench_" + strings.ReplaceAll(uuid.NewString(), "-", "")
				if err := db.Exec("CREATE TABLE " + table + " (id TEXT PRIMARY KEY, payload TEXT NOT NULL)").Error; err != nil {
					b.Fatal(err)
				}
				b.Cleanup(func() { db.Exec("DROP TABLE IF EXISTS " + table) })
				start := time.Now()
				if err := gorm.G[benchRow](db).Table(table).CreateInBatches(context.Background(), &rows, 1000); err != nil {
					b.Fatal(err)
				}
				insert := time.Since(start)
				var indexBytes int64
				if err := db.Raw("SELECT pg_relation_size(?::regclass)", table+"_pkey").Scan(&indexBytes).Error; err != nil {
					b.Fatal(err)
				}
				start = time.Now()
				for _, i := range indices {
					if _, err := gorm.G[benchRow](db).Table(table).Where("id = ?", rows[i].ID).First(context.Background()); err != nil {
						b.Fatal(err)
					}
				}
				lookup := time.Since(start)
				totalInsert += insert
				totalLookup += lookup
				totalIndex += indexBytes
				if err := db.Exec("DROP TABLE " + table).Error; err != nil {
					b.Fatal(err)
				}

			}
			b.ReportMetric(0, "ns/op")
			b.ReportMetric(float64(totalInsert.Microseconds())/float64(count*b.N), "insert-us/row")
			b.ReportMetric(float64(totalLookup.Microseconds())/float64(len(indices)*b.N), "lookup-us/op")
			b.ReportMetric(float64(totalIndex)/float64(b.N), "index-bytes")
		})
	}
}
