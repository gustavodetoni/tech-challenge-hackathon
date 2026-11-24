package db
import (
 "embed"; "fmt"; "sort"
 "gorm.io/gorm"
)
//go:embed migrations/*.sql
var migrationFiles embed.FS
func Migrate(db *gorm.DB) error { if err:=db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())").Error;err!=nil{return err}; entries,err:=migrationFiles.ReadDir("migrations");if err!=nil{return err}; sort.Slice(entries,func(i,j int)bool{return entries[i].Name()<entries[j].Name()}); for _,e:=range entries { var n int; if err:=db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?",e.Name()).Scan(&n).Error;err!=nil{return err}; if n>0{continue}; body,err:=migrationFiles.ReadFile("migrations/"+e.Name());if err!=nil{return err}; if err:=db.Transaction(func(tx *gorm.DB)error{if err:=tx.Exec(string(body)).Error;err!=nil{return err};return tx.Exec("INSERT INTO schema_migrations(version) VALUES (?)",e.Name()).Error});err!=nil{return fmt.Errorf("migration %s: %w",e.Name(),err)} }; return nil }
