package kv

import (
	"strings"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

func Bucket(tx *bolt.Tx, bucket []string) *bolt.Bucket {
	b := tx.Bucket([]byte(bucket[0]))
	if b == nil {
		return nil
	}
	for _, part := range bucket[1:] {
		b = b.Bucket([]byte(part))
		if b == nil {
			return nil
		}
	}
	return b
}

func CreateBucket(db *bolt.DB, bucket string) (*bolt.Bucket, error) {
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b, err := tx.CreateBucketIfNotExists([]byte(parts[0]))
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
	}
	for _, part := range parts[1:] {
		b, err = b.CreateBucketIfNotExists([]byte(part))
		if err != nil {
			return nil, err
		}
		if b == nil {
			return nil, nil
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return b, nil
}

var delimiter = ":"

func Get(db *bolt.DB, bucket, key string) (*string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := Bucket(tx, parts)
	if b == nil {
		return nil, nil
	}
	res := string(b.Get([]byte(key)))
	return &res, nil
}

func Exists(db *bolt.DB, bucket string) (bool, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := Bucket(tx, parts)
	if b == nil {
		return false, nil
	}
	return true, nil
}

func Keys(db *bolt.DB, bucket string) ([]string, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := Bucket(tx, parts)
	res := make([]string, 0)
	if b == nil {
		return nil, nil
	}
	err = b.ForEach(func(k, v []byte) error {
		res = append(res, string(k))
		return nil
	})
	return res, err
}

func Put(db *bolt.DB, bucket, key, value string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := Bucket(tx, parts)
	if b == nil {
		return nil
	}
	err = b.Put([]byte(key), []byte(value))
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func Del(db *bolt.DB, bucket, key string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	b := Bucket(tx, parts)
	if b == nil {
		return nil
	}
	err = b.Delete([]byte(key))
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func Drop(db *bolt.DB, bucket string) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	parts := strings.Split(bucket, delimiter)
	if len(parts) == 1 {
		if err = tx.DeleteBucket([]byte(bucket)); err != nil {
			return err
		}
	} else {
		b := Bucket(tx, parts[:len(parts)-2])
		if b == nil {
			return errors.New("key does not exist")
		}
		if err = b.DeleteBucket([]byte(parts[len(parts)-1])); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
