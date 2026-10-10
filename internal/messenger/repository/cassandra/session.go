package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
)

func NewSession(hosts []string, keyspace string, disableHostLookup bool) (gocqlx.Session, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.DisableInitialHostLookup = disableHostLookup

	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		return gocqlx.Session{}, fmt.Errorf("cassandra session: %w", err)
	}

	return session, nil
}
