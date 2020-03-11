package postgres_test

import (
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/reader"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/gofw/sqlquery/postgres"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v3"
	"os"
	"testing"
)

// TODO test autoincrement bigint, smallint
// TODO Null::Type as default value should be empty.
// TODO default value should be without type value::TYPE

func TestPostgres_Describe(t *testing.T) {

	// skip test on travis if database is not mysql
	if os.Getenv("TRAVIS") != "" && os.Getenv("DB") != "postgresql" {
		return
	}

	var cfg sqlquery.Config
	var err error
	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &json.JsonOptions{File: "../tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &json.JsonOptions{File: "../tests/db.psql.json"})
	}

	if assert.NoError(t, err) {
		b, err := sqlquery.NewBuilderFromConfig(&cfg)
		assert.NoError(t, err)

		driver := postgres.Postgres{}
		sel := driver.Describe("tests", "users", b, nil)
		assert.IsType(t, &sqlquery.Select{}, sel)

		rows, err := sel.All()
		assert.NoError(t, err)

		defer rows.Close()
		var cols []*sqlquery.Column

		for rows.Next() {
			var c sqlquery.Column
			var coltype string
			err := rows.Scan(&c.Name, &c.Position, &c.NullAble, &c.PrimaryKey, &coltype, &c.DefaultValue, &c.Length, &c.Autoincrement)
			assert.NoError(t, err)
			c.Type = driver.ConvertColumnType(coltype, &c)
			cols = append(cols, &c)
		}

		var columns []map[string]interface{}
		id := sqlquery.NewInt("integer")
		id.Min = 1
		id.Max = 2147483647
		columns = append(columns, map[string]interface{}{"Name": "id", "Position": 1, "NullAble": false, "Type": id, "PrimaryKey": true, "DefaultValue": "nextval('users_seq'::regclass)", "Length": 0, "Autoincrement": true})

		name := sqlquery.NewText("character varying")
		name.Size = 100
		columns = append(columns, map[string]interface{}{"Name": "name", "Position": 2, "NullAble": true, "Type": name, "PrimaryKey": false, "DefaultValue": "'Wall-E'::character varying", "Length": 100, "Autoincrement": false})

		surname := sqlquery.NewText("character varying")
		surname.Size = 100
		columns = append(columns, map[string]interface{}{"Name": "surname", "Position": 3, "NullAble": true, "Type": surname, "PrimaryKey": false, "DefaultValue": "NULL::character varying", "Length": 100, "Autoincrement": false})

		age := sqlquery.NewInt("integer")
		age.Min = -2147483648
		age.Max = 2147483647
		columns = append(columns, map[string]interface{}{"Name": "age", "Position": 4, "NullAble": true, "Type": age, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		birthday := sqlquery.NewDate("date")
		columns = append(columns, map[string]interface{}{"Name": "birthday", "Position": 5, "NullAble": true, "Type": birthday, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		deletedAt := sqlquery.NewDateTime("timestamp without time zone")
		columns = append(columns, map[string]interface{}{"Name": "deleted_at", "Position": 6, "NullAble": true, "Type": deletedAt, "PrimaryKey": false, "DefaultValue": "NULL::timestamp without time zone", "Length": 0, "Autoincrement": false})

		// Integers
		typeInt := sqlquery.NewInt("integer")
		typeInt.Min = -2147483648
		typeInt.Max = 2147483647
		columns = append(columns, map[string]interface{}{"Name": "type_int", "Position": 7, "NullAble": true, "Type": typeInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeBigInt := sqlquery.NewInt("bigint")
		typeBigInt.Min = -9223372036854775808
		typeBigInt.Max = 9223372036854775807
		columns = append(columns, map[string]interface{}{"Name": "type_bigint", "Position": 8, "NullAble": true, "Type": typeBigInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeSmallInt := sqlquery.NewInt("smallint")
		typeSmallInt.Min = -32768
		typeSmallInt.Max = 32767
		columns = append(columns, map[string]interface{}{"Name": "type_smallint", "Position": 9, "NullAble": true, "Type": typeSmallInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Floats
		typeFloat := sqlquery.NewFloat("real")
		columns = append(columns, map[string]interface{}{"Name": "type_real", "Position": 10, "NullAble": true, "Type": typeFloat, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeDouble := sqlquery.NewFloat("double precision")
		typeDouble.Name = "Float"
		columns = append(columns, map[string]interface{}{"Name": "type_double", "Position": 11, "NullAble": true, "Type": typeDouble, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeNumeric := sqlquery.NewFloat("numeric")
		typeNumeric.Name = "Float"
		columns = append(columns, map[string]interface{}{"Name": "type_numeric", "Position": 12, "NullAble": true, "Type": typeNumeric, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Texts
		typeVarChar := sqlquery.NewText("character varying")
		typeVarChar.Size = 10
		columns = append(columns, map[string]interface{}{"Name": "type_varchar", "Position": 13, "NullAble": true, "Type": typeVarChar, "PrimaryKey": false, "DefaultValue": "NULL::character varying", "Length": 10, "Autoincrement": false})

		typeChar := sqlquery.NewText("character")
		typeChar.Size = 10
		columns = append(columns, map[string]interface{}{"Name": "type_char", "Position": 14, "NullAble": true, "Type": typeChar, "PrimaryKey": false, "DefaultValue": "NULL::bpchar", "Length": 10, "Autoincrement": false})

		typeText := sqlquery.NewTextArea("text")
		columns = append(columns, map[string]interface{}{"Name": "type_text", "Position": 15, "NullAble": true, "Type": typeText, "PrimaryKey": false, "DefaultValue": "", "Length": typeText.Size, "Autoincrement": false})

		// Date and Time
		typeDate := sqlquery.NewDate("date")
		columns = append(columns, map[string]interface{}{"Name": "type_date", "Position": 16, "NullAble": true, "Type": typeDate, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeTimeStampTz := sqlquery.NewDateTime("timestamp with time zone")
		columns = append(columns, map[string]interface{}{"Name": "type_timestamp_tz", "Position": 17, "NullAble": true, "Type": typeTimeStampTz, "PrimaryKey": false, "DefaultValue": "NULL::timestamp without time zone", "Length": 0, "Autoincrement": false})
		typeTimeStamp := sqlquery.NewDateTime("timestamp without time zone")
		columns = append(columns, map[string]interface{}{"Name": "type_timestamp", "Position": 18, "NullAble": true, "Type": typeTimeStamp, "PrimaryKey": false, "DefaultValue": "NULL::timestamp without time zone", "Length": 0, "Autoincrement": false})

		typeTime := sqlquery.NewTime("time without time zone")
		columns = append(columns, map[string]interface{}{"Name": "type_time", "Position": 19, "NullAble": true, "Type": typeTime, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeTimeTz := sqlquery.NewTime("time with time zone")
		columns = append(columns, map[string]interface{}{"Name": "type_time_tz", "Position": 20, "NullAble": true, "Type": typeTimeTz, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Type is not defined Yet (Geometry)
		columns = append(columns, map[string]interface{}{"Name": "type_uuid", "Position": 21, "NullAble": true, "Type": nil, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		for index, col := range cols {
			assert.Equal(t, columns[index]["Name"], col.Name)
			assert.Equal(t, columns[index]["Position"], col.Position)
			assert.Equal(t, columns[index]["NullAble"], col.NullAble)
			assert.Equal(t, columns[index]["PrimaryKey"], col.PrimaryKey)
			assert.Equal(t, columns[index]["Type"], col.Type, col.Name, columns[index]["Type"])
			assert.Equal(t, null.NewString(columns[index]["DefaultValue"].(string), columns[index]["DefaultValue"].(string) != ""), col.DefaultValue)
			assert.Equal(t, null.NewInt(int64(columns[index]["Length"].(int)), columns[index]["Length"].(int) > 0), col.Length)
			assert.Equal(t, columns[index]["Autoincrement"], col.Autoincrement)
		}
	}
}

func TestPostgres_ForeignKeys(t *testing.T) {
	// skip test on travis if database is not mysql
	if os.Getenv("TRAVIS") != "" && os.Getenv("DB") != "postgresql" {
		return
	}

	var cfg sqlquery.Config
	var err error
	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &json.JsonOptions{File: "../tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &json.JsonOptions{File: "../tests/db.psql.json"})
	}

	if assert.NoError(t, err) {
		b, err := sqlquery.NewBuilderFromConfig(&cfg)
		assert.NoError(t, err)
		for index, table := range []string{"users", "user_posts", "robots", "posts", "histories", "addresses"} {
			driver := postgres.Postgres{}
			sel := driver.ForeignKeys("tests", table, b)
			assert.IsType(t, &sqlquery.Select{}, sel)

			rows, err := sel.All()
			assert.NoError(t, err)

			defer rows.Close()
			var fkeys []*sqlquery.ForeignKey

			for rows.Next() {
				f := sqlquery.ForeignKey{Primary: &sqlquery.Relation{}, Secondary: &sqlquery.Relation{}}
				err := rows.Scan(&f.Name, &f.Primary.Table, &f.Primary.Column, &f.Secondary.Table, &f.Secondary.Column)
				assert.NoError(t, err)
				fkeys = append(fkeys, &f)
			}

			switch index {
			case 0: //users
				assert.Equal(t, 0, len(fkeys))
			case 1: //user_posts (manyToMany)
				assert.Equal(t, 2, len(fkeys))
				assert.Equal(t, []string{"user_posts_ibfk_1", "user_posts", "user_id", "users", "id"}, []string{fkeys[0].Name, fkeys[0].Primary.Table, fkeys[0].Primary.Column, fkeys[0].Secondary.Table, fkeys[0].Secondary.Column})
				assert.Equal(t, []string{"user_posts_ibfk_2", "user_posts", "post_id", "posts", "id"}, []string{fkeys[1].Name, fkeys[1].Primary.Table, fkeys[1].Primary.Column, fkeys[1].Secondary.Table, fkeys[1].Secondary.Column})
			case 2: //robots
				assert.Equal(t, 0, len(fkeys))
			case 3: //posts
				assert.Equal(t, 0, len(fkeys))
			case 4: //histories (hasMany)
				assert.Equal(t, 1, len(fkeys))
				assert.Equal(t, []string{"histories_ibfk_1", "histories", "user_id", "users", "id"}, []string{fkeys[0].Name, fkeys[0].Primary.Table, fkeys[0].Primary.Column, fkeys[0].Secondary.Table, fkeys[0].Secondary.Column})
			case 5: //addresses (hasOne)
				assert.Equal(t, 1, len(fkeys))
				assert.Equal(t, []string{"addresses_ibfk_1", "addresses", "id", "users", "id"}, []string{fkeys[0].Name, fkeys[0].Primary.Table, fkeys[0].Primary.Column, fkeys[0].Secondary.Table, fkeys[0].Secondary.Column})
			}
		}
	}
}
