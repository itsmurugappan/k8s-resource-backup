package storage

import ()

type StorageInterface interface {
	Store(data []byte, ns string, snapid string, name string)
	Retrieve(ns string, snapid string, name string) []byte
}

func GetDBType(db string, connString string) StorageInterface {

	// use the db type to pull the correct database object
	// current supports CRDB only
	// for additional datasoure make changes here
	switch db {
	case "crdb":
		return &CRDB{
			ConnectionString: connString,
		}
	case "files":
		return &Files{
			Path: connString,
		}
	// case "mysql"
	// case "pg"
	default:
		return &Files{
			Path: connString,
		}
	}
}
