/*
 * Copyright © 2015-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @Copyright 	2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 */

package jwk_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ory/hydra/jwk"
	"github.com/ory/sqlcon/dockertest"
	"github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/require"
)

var createJWKMigrations = []*migrate.Migration{
	{
		Id: "1-data",
		Up: []string{
			`INSERT INTO hydra_jwk (sid, kid, version, keydata) VALUES ('1-sid', '1-kid', 0, 'some-key')`,
		},
		Down: []string{
			`DELETE FROM hydra_jwk WHERE sid='1-sid'`,
		},
	},
	{
		Id: "2-data",
		Up: []string{
			`INSERT INTO hydra_jwk (sid, kid, version, keydata, created_at) VALUES ('2-sid', '2-kid', 0, 'some-key', NOW())`,
		},
		Down: []string{
			`DELETE FROM hydra_jwk WHERE sid='2-sid'`,
		},
	},
	{
		Id: "3-data",
		Up: []string{
			`INSERT INTO hydra_jwk (sid, kid, version, keydata, created_at) VALUES ('3-sid', '3-kid', 0, 'some-key', NOW())`,
		},
		Down: []string{
			`DELETE FROM hydra_jwk WHERE sid='3-sid'`,
		},
	},
}

var migrations = &migrate.MemoryMigrationSource{
	Migrations: []*migrate.Migration{
		{Id: "0-data", Up: []string{"DROP TABLE IF EXISTS hydra_jwk"}},
		jwk.Migrations.Migrations[0],
		createJWKMigrations[0],
		jwk.Migrations.Migrations[1],
		createJWKMigrations[1],
		jwk.Migrations.Migrations[2],
		createJWKMigrations[2],
	},
}

func TestMigrations(t *testing.T) {
	var dbs = map[string]*sqlx.DB{}
	if testing.Short() {
		return
	}

	dockertest.Parallel([]func(){
		func() {
			db, err := dockertest.ConnectToTestPostgreSQL()
			if err != nil {
				log.Fatalf("Could not connect to database: %v", err)
			}
			dbs["postgres"] = db
		},
		func() {
			db, err := dockertest.ConnectToTestMySQL()
			if err != nil {
				log.Fatalf("Could not connect to database: %v", err)
			}
			dbs["mysql"] = db
		},
	})

	for k, db := range dbs {
		t.Run(fmt.Sprintf("database=%s", k), func(t *testing.T) {
			migrate.SetTable("hydra_jwk_migration_integration")
			for step := range migrations.Migrations {
				t.Run(fmt.Sprintf("step=%d", step), func(t *testing.T) {
					n, err := migrate.ExecMax(db.DB, db.DriverName(), migrations, migrate.Up, 1)
					require.NoError(t, err)
					require.Equal(t, n, 1)
				})
			}

			for step := range migrations.Migrations {
				t.Run(fmt.Sprintf("step=%d", step), func(t *testing.T) {
					n, err := migrate.ExecMax(db.DB, db.DriverName(), migrations, migrate.Down, 1)
					require.NoError(t, err)
					require.Equal(t, n, 1)
				})
			}
		})
	}
}
