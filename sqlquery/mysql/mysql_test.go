package mysql_test

import (
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/reader"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/gofw/sqlquery/mysql"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"gopkg.in/guregu/null.v3"
)

func TestMysql_Describe(t *testing.T) {

	// skip test on travis if database is not mysql
	if os.Getenv("TRAVIS") != "" && os.Getenv("DB") != "mysql" {
		return
	}

	var cfg sqlquery.Config
	var err error
	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "../tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "../tests/db.json"})
	}

	if assert.NoError(t, err) {
		b, err := sqlquery.NewBuilderFromConfig(&cfg)
		assert.NoError(t, err)

		driver := mysql.Mysql{}
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
		id := sqlquery.NewInt("int(11) unsigned")
		id.Min = 0
		id.Max = 4294967295
		columns = append(columns, map[string]interface{}{"Name": "id", "Position": 1, "NullAble": false, "Type": id, "PrimaryKey": true, "DefaultValue": "", "Length": 0, "Autoincrement": true})
		name := sqlquery.NewText("varchar(100)")
		name.Size = 100
		columns = append(columns, map[string]interface{}{"Name": "name", "Position": 2, "NullAble": true, "Type": name, "PrimaryKey": false, "DefaultValue": "Wall-E", "Length": 100, "Autoincrement": false})

		surname := sqlquery.NewText("varchar(100)")
		surname.Size = 100
		columns = append(columns, map[string]interface{}{"Name": "surname", "Position": 3, "NullAble": true, "Type": surname, "PrimaryKey": false, "DefaultValue": "", "Length": 100, "Autoincrement": false})

		age := sqlquery.NewInt("int(11)")
		age.Min = -2147483648
		age.Max = 2147483647
		columns = append(columns, map[string]interface{}{"Name": "age", "Position": 4, "NullAble": true, "Type": age, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		birthday := sqlquery.NewDate("date")
		columns = append(columns, map[string]interface{}{"Name": "birthday", "Position": 5, "NullAble": true, "Type": birthday, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		deletedAt := sqlquery.NewDateTime("datetime")
		columns = append(columns, map[string]interface{}{"Name": "deleted_at", "Position": 6, "NullAble": true, "Type": deletedAt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Integers
		typeInt := sqlquery.NewInt("int(11)")
		typeInt.Min = -2147483648
		typeInt.Max = 2147483647
		columns = append(columns, map[string]interface{}{"Name": "type_int", "Position": 7, "NullAble": true, "Type": typeInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeUInt := sqlquery.NewInt("int(10) unsigned")
		typeUInt.Min = 0
		typeUInt.Max = 4294967295
		columns = append(columns, map[string]interface{}{"Name": "type_uint", "Position": 8, "NullAble": true, "Type": typeUInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeBigInt := sqlquery.NewInt("bigint(11)")
		typeBigInt.Min = -9223372036854775808
		typeBigInt.Max = 9223372036854775807
		columns = append(columns, map[string]interface{}{"Name": "type_bigint", "Position": 9, "NullAble": true, "Type": typeBigInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeUBigInt := sqlquery.NewInt("bigint(20) unsigned")
		typeUBigInt.Min = 0
		typeUBigInt.Max = 18446744073709551615
		columns = append(columns, map[string]interface{}{"Name": "type_ubigint", "Position": 10, "NullAble": true, "Type": typeUBigInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeMediumInt := sqlquery.NewInt("mediumint(9)")
		typeMediumInt.Min = -8388608
		typeMediumInt.Max = 8388607
		columns = append(columns, map[string]interface{}{"Name": "type_mediumint", "Position": 11, "NullAble": true, "Type": typeMediumInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeUMediumInt := sqlquery.NewInt("mediumint(9) unsigned")
		typeUMediumInt.Min = 0
		typeUMediumInt.Max = 16777215
		columns = append(columns, map[string]interface{}{"Name": "type_umediumint", "Position": 12, "NullAble": true, "Type": typeUMediumInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeSmallInt := sqlquery.NewInt("smallint(6)")
		typeSmallInt.Min = -32768
		typeSmallInt.Max = 32767
		columns = append(columns, map[string]interface{}{"Name": "type_smallint", "Position": 13, "NullAble": true, "Type": typeSmallInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeUSmallInt := sqlquery.NewInt("smallint(6) unsigned")
		typeUSmallInt.Min = 0
		typeUSmallInt.Max = 65535
		columns = append(columns, map[string]interface{}{"Name": "type_usmallint", "Position": 14, "NullAble": true, "Type": typeUSmallInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		typeTinyInt := sqlquery.NewInt("tinyint(4)")
		typeTinyInt.Min = -128
		typeTinyInt.Max = 127
		columns = append(columns, map[string]interface{}{"Name": "type_tinyint", "Position": 15, "NullAble": true, "Type": typeTinyInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeUTinyInt := sqlquery.NewInt("tinyint(4) unsigned")
		typeUTinyInt.Min = 0
		typeUTinyInt.Max = 255
		columns = append(columns, map[string]interface{}{"Name": "type_utinyint", "Position": 16, "NullAble": true, "Type": typeUTinyInt, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Floats
		typeFloat := sqlquery.NewFloat("float")
		columns = append(columns, map[string]interface{}{"Name": "type_float", "Position": 17, "NullAble": true, "Type": typeFloat, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeDouble := sqlquery.NewFloat("double")
		typeDouble.Name = "Float"
		columns = append(columns, map[string]interface{}{"Name": "type_double", "Position": 18, "NullAble": true, "Type": typeDouble, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Texts
		typeVarChar := sqlquery.NewText("varchar(10)")
		typeVarChar.Size = 10
		columns = append(columns, map[string]interface{}{"Name": "type_varchar", "Position": 19, "NullAble": true, "Type": typeVarChar, "PrimaryKey": false, "DefaultValue": "", "Length": 10, "Autoincrement": false})
		typeChar := sqlquery.NewText("char(10)")
		typeChar.Size = 10
		columns = append(columns, map[string]interface{}{"Name": "type_char", "Position": 20, "NullAble": true, "Type": typeChar, "PrimaryKey": false, "DefaultValue": "", "Length": 10, "Autoincrement": false})
		typeTinyText := sqlquery.NewTextArea("tinytext")
		typeTinyText.Size = 255
		columns = append(columns, map[string]interface{}{"Name": "type_tinytext", "Position": 21, "NullAble": true, "Type": typeTinyText, "PrimaryKey": false, "DefaultValue": "", "Length": typeTinyText.Size, "Autoincrement": false})
		typeText := sqlquery.NewTextArea("text")
		typeText.Size = 65535
		columns = append(columns, map[string]interface{}{"Name": "type_text", "Position": 22, "NullAble": true, "Type": typeText, "PrimaryKey": false, "DefaultValue": "", "Length": typeText.Size, "Autoincrement": false})
		typeMediumText := sqlquery.NewTextArea("mediumtext")
		typeMediumText.Size = 16777215
		columns = append(columns, map[string]interface{}{"Name": "type_mediumtext", "Position": 23, "NullAble": true, "Type": typeMediumText, "PrimaryKey": false, "DefaultValue": "", "Length": typeMediumText.Size, "Autoincrement": false})
		typeLongText := sqlquery.NewTextArea("longtext")
		typeLongText.Size = 4294967295
		columns = append(columns, map[string]interface{}{"Name": "type_longtext", "Position": 24, "NullAble": true, "Type": typeLongText, "PrimaryKey": false, "DefaultValue": "", "Length": typeLongText.Size, "Autoincrement": false})

		// Date and Time
		typeTime := sqlquery.NewTime("time")
		columns = append(columns, map[string]interface{}{"Name": "type_time", "Position": 25, "NullAble": true, "Type": typeTime, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeDate := sqlquery.NewDate("date")
		columns = append(columns, map[string]interface{}{"Name": "type_date", "Position": 26, "NullAble": true, "Type": typeDate, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeTimeStamp := sqlquery.NewDateTime("timestamp")
		columns = append(columns, map[string]interface{}{"Name": "type_timestamp", "Position": 27, "NullAble": true, "Type": typeTimeStamp, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})
		typeDateTime := sqlquery.NewDateTime("datetime")
		columns = append(columns, map[string]interface{}{"Name": "type_datetime", "Position": 28, "NullAble": true, "Type": typeDateTime, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		// Type is not defined Yet (Geometry)
		columns = append(columns, map[string]interface{}{"Name": "type_geometry", "Position": 29, "NullAble": true, "Type": nil, "PrimaryKey": false, "DefaultValue": "", "Length": 0, "Autoincrement": false})

		for index, col := range cols {
			assert.Equal(t, columns[index]["Name"], col.Name)
			assert.Equal(t, columns[index]["Position"], col.Position)
			assert.Equal(t, columns[index]["NullAble"], col.NullAble)
			assert.Equal(t, columns[index]["PrimaryKey"], col.PrimaryKey)
			assert.Equal(t, columns[index]["Type"], col.Type, col.Name, columns[index]["Type"])
			assert.Equal(t, null.NewString(columns[index]["DefaultValue"].(string), columns[index]["DefaultValue"].(string) != ""), col.DefaultValue)
			assert.Equal(t, null.NewInt(int64(columns[index]["Length"].(int)), columns[index]["Length"].(int) > 0), col.Length, col.Name)
			assert.Equal(t, columns[index]["Autoincrement"], col.Autoincrement)
		}
	}
}

func TestMysql_ForeignKeys(t *testing.T) {
	// skip test on travis if database is not mysql
	if os.Getenv("TRAVIS") != "" && os.Getenv("DB") != "mysql" {
		return
	}

	var cfg sqlquery.Config
	var err error
	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "../tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "../tests/db.json"})
	}

	if assert.NoError(t, err) {
		b, err := sqlquery.NewBuilderFromConfig(&cfg)
		assert.NoError(t, err)

		for index, table := range []string{"users", "user_posts", "robots", "posts", "histories", "addresses"} {
			driver := mysql.Mysql{}
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
