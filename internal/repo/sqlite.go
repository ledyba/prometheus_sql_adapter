package repo

import (
	"sort"
	"strings"

	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

func sqliteInit() error {
	_, err := db.Exec(`
create table if not exists timeseries(
  id integer primary key autoincrement
);
create table if not exists labels(
  timeseries_id integer not null,
  name integer not null,
  value integer not null
);
create table if not exists literals(
  id integer primary key autoincrement,
  value text unique not null
);
create table if not exists samples(
  timeseries_id integer not null,
  timestamp integer not null,
  value real not null
);

-- labels
create index if not exists labels_timeseries_index on labels(timeseries_id);

-- samples
create index if not exists samples_timestamp_index on samples(timestamp);
create index if not exists samples_timeseries_index on samples(timeseries_id);

-- literals
create index if not exists literals_value_index on literals(value);
`)
	if err != nil {
		log.Info("Failed to initialize database", zap.String("driver", "sqlite"), zap.Error(err))
	} else {
		log.Info("Database Initialized", zap.String("driver", "sqlite"))
	}
	return err
}

func sqliteWrite(req *prompb.WriteRequest) error {
	var err error
	{ // insert labels
		labelSQL := ""
		labelValue := make([]interface{}, 0)
		for _, timeseries := range req.Timeseries {
			for _, label := range timeseries.Labels {
				labelSQL += ",(?),(?)"
				labelValue = append(labelValue, label.Name, label.Value)
			}
		}
		sort.Slice(labelValue, func(i,j int) bool {
			return strings.Compare(labelValue[i].(string), labelValue[j].(string)) < 0
		})

		labelSQL = `insert or ignore into literals (value) values ` + labelSQL[1:]
		_, err = db.Exec(labelSQL, labelValue...)
		if err != nil {
			log.Error("Failed to append literals", zap.Error(err))
			return err
		}
	}
	labelSQL := ""
	labelValue := make([]interface{}, 0)
	sampleSQL := ""
	sampleValue := make([]interface{}, 0)
	for _, ts := range req.Timeseries {
		_, err = db.Exec(`insert into timeseries default values`)
		if err != nil {
			log.Error("Failed to create new timeseries", zap.Error(err))
			return err
		}
		row := db.QueryRow(`select id from timeseries where rowid = last_insert_rowid()`)
		if row.Err() != nil {
			log.Error("Failed to select id from timeseries where rowid = last_insert_rowid()", zap.Error(row.Err()))
			return row.Err()
		}
		var id uint64
		if err = row.Scan(&id); err != nil {
			return err
		}
		for _, label := range ts.Labels {
			labelSQL += `,(?, (select id from literals where value = ?), (select id from literals where value = ?))`
			labelValue = append(labelValue, id, label.Name, label.Value)
		}
		for _, sample := range ts.Samples {
			sampleSQL += `,(?,?,?)`
			sampleValue = append(sampleValue, id, sample.Timestamp, sample.Value)
		}
	}
	labelSQL = `insert into labels (timeseries_id, name, value) values ` + labelSQL[1:]
	_, err = db.Exec(labelSQL, labelValue...)
	if err != nil {
		log.Error("Failed to write labels to database", zap.Error(err))
		return err
	}
	sampleSQL = `insert into samples (timeseries_id, timestamp, value) values ` + sampleSQL[1:]
	_, err = db.Exec(sampleSQL, sampleValue...)
	if err != nil {
		log.Error("Failed to write samples to database", zap.Error(err))
		return err
	}
	return nil
}
