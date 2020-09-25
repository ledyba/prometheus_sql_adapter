package repo

import (
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

func mysqlInit() error {
	var err error
	_, err = db.Exec(`
	create table if not exists timeseries (
		id bigint unsigned auto_increment not null,
		primary key (id)
	) ENGINE=InnoDB;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	create table if not exists labels (
		timeseries_id bigint unsigned not null,
		name bigint unsigned not null,
		value bigint unsigned not null,
		index (timeseries_id)
	) ENGINE=InnoDB;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	create table if not exists literals (
		id bigint unsigned auto_increment not null,
		value varchar(256) unique not null,
		primary key (id),
		index (value)
	) ENGINE=InnoDB;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	create table if not exists samples(
		timeseries_id bigint unsigned not null,
		timestamp bigint unsigned not null,
		value double not null,
		index (timeseries_id),
		index (timestamp)
	) ENGINE=InnoDB;`)
	if err != nil {
		return err
	}

	if err == nil {
		log.Info("Database Initialized", zap.String("driver", "sqlite"))
	}
	return err
}

func mysqlWrite(req *prompb.WriteRequest) error {
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
		labelSQL = "insert ignore into `literals` (`value`) values " + labelSQL[1:]
		_, err = db.Exec(labelSQL, labelValue...)
		if err != nil {
			return err
		}
	}
	labelSQL := ""
	labelValue := make([]interface{}, 0)
	sampleSQL := ""
	sampleValue := make([]interface{}, 0)
	for _, ts := range req.Timeseries {
		_, err = db.Exec("insert into timeseries () values ()")
		if err != nil {
			return err
		}
		row := db.QueryRow(`select last_insert_id()`)
		if row.Err() != nil {
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
		return err
	}
	sampleSQL = `insert into samples (timeseries_id, timestamp, value) values ` + sampleSQL[1:]
	_, err = db.Exec(sampleSQL, sampleValue...)
	if err != nil {
		return err
	}
	return nil
}
