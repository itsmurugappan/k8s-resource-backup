package storage

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Files struct {
	Path string
}

func (f *Files) Store(data []byte, ns string, snap_id string, name string) {

	fPath := filepath.Join(f.Path, snap_id, ns)

	if _, err := os.Stat(fPath); os.IsNotExist(err) {
		os.MkdirAll(fPath, 0777)
	}

	if err := ioutil.WriteFile(fPath+"/"+name+".json", data, 0777); err != nil {
		log.Printf("Error writing resource %s in ns %s for snapid %s : %s", name, ns, snap_id, err)
	}
}

func (f *Files) Retrieve(ns string, snap_id string, name string) []byte {
	fPath := filepath.Join(f.Path, snap_id, ns)

	result, err := ioutil.ReadFile(fPath + "/" + name + ".json")

	if err != nil {
		log.Printf("error retrieving file from path %s : %s", fPath, err)
		return []byte("")
	}

	return result
}
