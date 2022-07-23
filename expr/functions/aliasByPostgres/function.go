package aliasByPostgres

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-graphite/carbonapi/expr/helper"
	"github.com/go-graphite/carbonapi/expr/helper/metric"
	"github.com/go-graphite/carbonapi/expr/interfaces"
	"github.com/go-graphite/carbonapi/expr/types"
	"github.com/go-graphite/carbonapi/pkg/parser"

	_ "github.com/lib/pq" // Needed for proper work of postgresql requests
	"github.com/lomik/zapwriter"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type aliasByPostgres struct {
	interfaces.FunctionBase
	Enabled  bool
	Database map[string]Database
}

// KeyString structure
type KeyString struct {
	VarName     string
	QueryString string
	MatchString string
}

// Database structure
type Database struct {
	URLDB     string
	Username  string
	Password  string
	NameDB    string
	KeyString map[string]KeyString
}

type aliasByPostgresConfig struct {
	Enabled  bool
	Database map[string]Database
}

func (f *aliasByPostgres) SQLConnectDB(databaseName string) (*sql.DB, error) {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "aliasByPostgres"))
	connectString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", f.Database[databaseName].Username, f.Database[databaseName].Password, f.Database[databaseName].URLDB, f.Database[databaseName].NameDB)
	logger.Debug(connectString)
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		logger.Error("Error connect to PostgreSQL Database")
		return nil, err
	}
	return db, nil
}

// SQLQueryDB convenience function to query the database
func (f *aliasByPostgres) SQLQueryDB(query, databaseName string) (res string, err error) {
	var result string
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "aliasByPostgres"))
	db, _ := f.SQLConnectDB(databaseName)
	rows, err := db.Query(query)
	if err != nil {
		logger.Error("Error with query ti database")
	}
	defer func() {
		_ = db.Close()
	}()
	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			logger.Error("Error with scan response")
		}
		logger.Debug(result)
	}
	defer func() {
		_ = rows.Close()
	}()
	return result, nil
}

// GetOrder - standard function
func GetOrder() interfaces.Order {
	return interfaces.Any
}

// New - function for parsing config
func New(configFile string) []interfaces.FunctionMetadata {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "aliasByPostgres"))
	if configFile == "" {
		logger.Debug("no config file specified",
			zap.String("message", "this function requrires config file to work properly"),
		)
		return nil
	}
	v := viper.New()
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	if err != nil {
		logger.Error("failed to read config file",
			zap.Error(err),
		)
		return nil
	}
	key := map[string]KeyString{
		"keyString": {
			VarName:     "var",
			QueryString: "select * from database.table where \"name\" =~ /^var0$/",
			MatchString: ".*",
		},
	}
	database := map[string]Database{
		"postgres": {
			URLDB:     "http://localhost:5432",
			Username:  "User",
			Password:  "Password",
			NameDB:    "databaseName",
			KeyString: key,
		},
	}

	cfg := aliasByPostgresConfig{
		Enabled:  false,
		Database: database,
	}
	err = v.Unmarshal(&cfg)
	if err != nil {
		logger.Error("failed to parse config",
			zap.Error(err),
		)
		return nil
	}
	if !cfg.Enabled {
		logger.Warn("aliasByPostgres config found but aliasByPostgres is disabled")
		return nil
	}
	f := &aliasByPostgres{
		Enabled:  cfg.Enabled,
		Database: cfg.Database,
	}
	res := make([]interfaces.FunctionMetadata, 0)
	for _, n := range []string{"aliasByPostgres"} {
		res = append(res, interfaces.FunctionMetadata{Name: n, F: f})
	}
	return res
}

func (f *aliasByPostgres) Do(ctx context.Context, e parser.Expr, from, until int64, values map[parser.MetricRequest][]*types.MetricData) ([]*types.MetricData, error) {
	logger := zapwriter.Logger("functionInit").With(zap.String("function", "aliasByPostgres"))
	args, err := helper.GetSeriesArg(ctx, e.Args()[0], from, until, values)
	if err != nil {
		return nil, err
	}

	fields, err := e.GetIntArgs(3)
	if err != nil {
		return nil, err
	}

	databaseName, err := e.GetStringArg(1)
	if err != nil {
		return nil, err
	}

	keyString, err := e.GetStringArg(2)
	if err != nil {
		return nil, err
	}

	var results []*types.MetricData
	matchString := regexp.MustCompile(f.Database[databaseName].KeyString[keyString].MatchString)

	for _, a := range args {
		metric := metric.ExtractMetric(a.Name)
		logger.Debug(metric)
		if metric == "" {
			continue
		}
		nodes := strings.Split(metric, ".")
		var name []string
		for _, f := range fields {
			if f < 0 {
				f += len(nodes)
			}
			if f >= len(nodes) || f < 0 {
				continue
			}
			name = append(name, nodes[f])
		}
		tempName := strings.Join(name, ".")
		query := f.Database[databaseName].KeyString[keyString].QueryString
		varName := regexp.MustCompile(f.Database[databaseName].KeyString[keyString].VarName)
		queryFields := len(varName.FindAllString(query, -1))

		for i := 0; i < queryFields; i++ {
			reg := regexp.MustCompile("(" + f.Database[databaseName].KeyString[keyString].VarName + strings.TrimSpace(strconv.Itoa(i)) + ")")
			query = reg.ReplaceAllString(query, name[i])
		}

		res, err := f.SQLQueryDB(query, databaseName)
		if err != nil {
			logger.Error("failed query to Postgresql DB", zap.Error(err))
			return nil, err
		}
		for i := range name {
			if i < queryFields {
				name = append(name[:0], name[0+1:]...)
			}
		}
		if len(res) > 0 {
			if matchString.MatchString(res) {
				r := *a.CopyLink()
				r.Name = strings.Join(name, ".")
				if len(name) > 0 {
					r.Name = res + "." + r.Name
					results = append(results, &r)
				} else {
					r.Name = res
					results = append(results, &r)
				}
				r.Tags["name"] = r.Name
			}
		} else {
			r := *a
			r.Name = tempName
			r.Tags["name"] = r.Name
			results = append(results, &r)
		}
	}
	return results, nil
}

// Description is auto-generated description, based on output of https://github.com/graphite-project/graphite-web
func (f *aliasByPostgres) Description() map[string]types.FunctionDescription {
	return map[string]types.FunctionDescription{
		"aliasByPostgres": {
			Description: "Takes a seriesList and applies an alias derived from database for one or more \"node\"\n portion/s of the target name or tags. Node indices are 0 indexed.\n\n.. code-block:: none\n\n  &target=aliasByPostgres(ganglia.*.cpu.load5,'database','key-string',1)\n\nEach node may be an integer referencing a node in the series name or a string identifying a tag.\n\n.. code-block :: none\n\n aliasByPostgres(\"datacenter\", \"server\", 1)\n\n  # will produce output series like\n  # dc1.server1.load5, dc1.server2.load5, dc1.server1.load10, dc1.server2.load10",
			Function:    "aliasByPostgres(seriesList, *nodes)",
			Group:       "Alias",
			Module:      "graphite.render.functions",
			Name:        "aliasByPostgres",
			Params: []types.FunctionParam{
				{
					Name:     "seriesList",
					Required: true,
					Type:     types.SeriesList,
				},
				{
					Name:     "databaseName",
					Required: true,
					Type:     types.String,
				},
				{
					Name:     "keyString",
					Required: true,
					Type:     types.String,
				},
				{
					Multiple: true,
					Name:     "nodes",
					Required: true,
					Type:     types.NodeOrTag,
				},
			},
		},
	}
}
