# Backup and Restore k8s Resources

This is a utility to back up and restore k8s resources

I wrote it mainly to backup knative resources but can be extended to other k8s resources.


### Adding new resources

For backing up new resources implement the  ```BackupInterface``` .

### Storage

Currently supports 
1. File
2. crdb

To support additional storages, implement the  ```StorageInterface``` .

### Labels

While running the utility specify the namespace labels you want the utility to look for 

### Knative

To see the details of the knative resource backup. Please take a look at pkg/backupknative.go 

### Running the utility

```
storageinfo="/Developer/misc/files" db=files snapshot_id=20200126kind \
labels="backup=true" action=restore excluded_ns=test go run cmd/main.go
```

| Parameter | Description | Values |
|-----------|-------------|--------|
| storageinfo | place the files will be stored | filepath or db connectioinfo |
| db | storage type | crdb or files |
| snapshot_id | unique id for the backup snapshot | |
| labels | namespace labels the utility needs to look for | eg. backup=true |
| action | restore or backup | restore or backup |
| excluded_ns | namespace you want to exclude from backup | | 