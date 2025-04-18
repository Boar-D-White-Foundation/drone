package main

import (
	"context"
	"fmt"

	"github.com/boar-d-white-foundation/drone/db"
)

const (
	keyAppliedMigrations = "drone:applied_migrations"
)

func migrate(ctx context.Context, database db.DB) error {
	return db.MigrateJson(ctx, database, keyAppliedMigrations, []db.Migration{
		{ID: "0001", Name: "add_default_greeted_users", Fn: addDefaultGreetedUsers},
		{ID: "0002", Name: "drop_poisoned_db_queue", Fn: dropPoisonedDBQueue},
		{ID: "0003", Name: "add_initial_okr_values", Fn: addInitialOkrValues},
	})
}

func addDefaultGreetedUsers(tx db.Tx) error {
	key := "boardwhite:on_join_greeted_users"
	greetedUsers, err := db.GetJsonDefault(tx, key, make(map[int64]struct{}))
	if err != nil {
		return fmt.Errorf("get %q: %w", key, err)
	}

	for _, uid := range []int64{142944542} {
		greetedUsers[uid] = struct{}{}
	}
	if err := db.SetJson(tx, key, greetedUsers); err != nil {
		return fmt.Errorf("set %q: %w", key, err)
	}

	return nil
}

func dropPoisonedDBQueue(tx db.Tx) error {
	key := "dbq:queue:boardwhite:post_code_snippet"
	if err := db.SetJson[*int](tx, key, nil); err != nil {
		return fmt.Errorf("set %q: %w", key, err)
	}

	return nil
}

func addInitialOkrValues(tx db.Tx) error {
	key := "boardwhite:okr:values"

	type okrs struct {
		TotalCount map[string]int `json:"total_count"`
	}

	counts := map[string]int{
		"#unfortunately2025": 39,
		"#bigtech_offer2025": 1,
		"#faang_offer2025":   1,
		"#senior_promo2025":  0,
		"#staff_promo2025":   0,
		"#usa2025":           0,
	}
	if err := db.SetJson(tx, key, okrs{TotalCount: counts}); err != nil {
		return fmt.Errorf("set %q: %w", key, err)
	}

	return nil
}
