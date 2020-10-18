package repo

import (
	"database/sql"
	"io"

	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

type mysqlDriver struct {
	db *sql.DB
}

func newMysql(db *sql.DB) Driver {
	return &mysqlDriver{
		db: db,
	}
}

func (d *mysqlDriver) Init() error {
	db.SetMaxOpenConns(16)
	var err error
	_, err = db.Exec(`
	create table if not exists timeseries (
		id bigint unsigned auto_increment not null,
		primary key (id)
	) ENGINE=InnoDB;`)
	if err != nil {
		log.Info("Failed to initialize database", zap.String("driver", "mysql"), zap.Error(err))
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
		log.Info("Failed to initialize database", zap.String("driver", "mysql"), zap.Error(err))
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
		log.Info("Failed to initialize database", zap.String("driver", "mysql"), zap.Error(err))
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
		log.Info("Failed to initialize database", zap.String("driver", "mysql"), zap.Error(err))
		return err
	}

	if err == nil {
		log.Info("Database Initialized", zap.String("driver", "mysql"))
	}

	return err
}

func (d *mysqlDriver) Write(req *prompb.WriteRequest) error {
	var err error
	var result sql.Result
	numLiteralsTotal := 0
	numLiteralsInserted := int64(0)
	for _, timeseries := range req.Timeseries {
		for _, label := range timeseries.Labels {
			for _, literal := range []string{label.Name, label.Value} {
				result, err = db.Exec("insert ignore into `literals` (`value`) values (?)", literal)
				if err != nil {
					log.Error("Failed to append literals", zap.Error(err))
					return err
				}
				var affected int64
				affected, err = result.RowsAffected()
				if err != nil {
					log.Error("Failed to read rows affected", zap.Error(err))
					return err
				}
				numLiteralsInserted += affected
				numLiteralsTotal++
			}
		}
	}
	numLabelsTotal := 0
	numLabelsInserted := int64(0)
	labelSQL := ""
	labelValue := make([]interface{}, 0)
	numSamplesTotal := 0
	numSamplesInserted := int64(0)
	sampleSQL := ""
	sampleValue := make([]interface{}, 0)
	for _, ts := range req.Timeseries {
		var id int64
		result, err = db.Exec("insert into timeseries () values ()")
		if err != nil {
			log.Error("Failed to create new timeseries", zap.Error(err))
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			log.Error("Failed to select last_insert_id()", zap.Error(err))
			return err
		}
		for _, label := range ts.Labels {
			labelSQL += `,(?, (select id from literals where value = ?), (select id from literals where value = ?))`
			labelValue = append(labelValue, id, label.Name, label.Value)
			numLabelsTotal++
		}
		for _, sample := range ts.Samples {
			sampleSQL += `,(?,?,?)`
			sampleValue = append(sampleValue, id, sample.Timestamp, sample.Value)
			numSamplesTotal++
		}
	}

	// Label batch insert
	labelSQL = `insert into labels (timeseries_id, name, value) values ` + labelSQL[1:]
	result, err = db.Exec(labelSQL, labelValue...)
	if err != nil {
		log.Error("Failed to write labels to database", zap.Error(err))
		return err
	}
	numLabelsInserted, err = result.RowsAffected()
	if err != nil {
		log.Error("Failed to read rows affected", zap.Error(err))
		return err
	}

	// Sample batch insert
	sampleSQL = `insert into samples (timeseries_id, timestamp, value) values ` + sampleSQL[1:]
	_, err = db.Exec(sampleSQL, sampleValue...)
	if err != nil {
		log.Error("Failed to write samples to database", zap.Error(err))
		return err
	}
	numSamplesInserted, err = result.RowsAffected()
	if err != nil {
		log.Error("Failed to read rows affected", zap.Error(err))
		return err
	}
	log.Info("Write Done",
		zap.String("driver", "mysql"),
		zap.Int("timeseries", len(req.Timeseries)),
		zap.Int("literals-total", numLiteralsTotal),
		zap.Int64("literals-inserted", numLiteralsInserted),
		zap.Int("labels-total", numLabelsTotal),
		zap.Int64("labels-inserted", numLabelsInserted),
		zap.Int("samples-total", numSamplesTotal),
		zap.Int64("samples-inserted", numSamplesInserted))
	return nil
}

func (d *mysqlDriver) Read(req *prompb.ReadRequest, w io.Writer) error {
	return nil
}

func (d *mysqlDriver) Close() {

}