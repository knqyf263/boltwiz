package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"

	"github.com/jhump/protoreflect/desc/protoparse"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/xerrors"

	"github.com/knqyf263/boltwiz/modules/database/model"
)

type Repository struct {
	db        *bolt.DB
	unmarshal func([]byte) string
}

func NewRepository(dbPath, protoType string, protoFiles []string) (*Repository, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, xerrors.Errorf("failed to open db: %w", err)
	}

	unmarshal := func(b []byte) string { return string(b) }

	if protoType != "" && len(protoFiles) > 0 {
		fileDescriptor, err := protoparse.Parser{}.ParseFiles(protoFiles...)
		if err != nil {
			return nil, xerrors.Errorf("failed to parse proto files: %w", err)
		}

		var md *desc.MessageDescriptor
		for _, fd := range fileDescriptor {
			if md = fd.FindMessage(protoType); md != nil {
				break
			}
		}
		if md == nil {
			return nil, xerrors.Errorf("failed to find the specified type (%s): %w", protoType, err)
		}

		unmarshal = func(b []byte) string {
			m := dynamic.NewMessage(md)
			if err := m.Unmarshal(b); err != nil {
				return fmt.Sprintf("failed to unmarshal message: %v", err)
			}
			b, err := json.Marshal(m)
			if err != nil {
				return fmt.Sprintf("failed to marshal message: %v", err)
			}
			return string(b)
		}
	}

	return &Repository{
		db:        db,
		unmarshal: unmarshal,
	}, nil
}
func (r *Repository) Close() error {
	// Skip closing the database if the connection is not established.
	if r.db == nil {
		return nil
	}
	if err := r.db.Close(); err != nil {
		return xerrors.Errorf("failed to close DB: %w", err)
	}
	return nil
}

func (r *Repository) ListElement(input model.ListElemReqBody) (elem model.ListedElem, err error) {
	var resultFullSet []model.Result
	cntOfRecords := 0
	searchkey := strings.ToLower(input.SearchKey)
	err = r.db.View(func(tx *bolt.Tx) error {
		var rootBkt *bolt.Bucket
		if len(input.LevelStack) > 0 {
			rootBkt = tx.Bucket([]byte(input.LevelStack[0]))
			if rootBkt == nil {
				return xerrors.New(fmt.Sprintf("No Root Bucket found by the name : %s", input.LevelStack[0]))
			}
			for i, val := range input.LevelStack[1:] {
				rootBkt = rootBkt.Bucket([]byte(val))
				if rootBkt == nil {
					return xerrors.New(fmt.Sprintf("No Bucket found by the name : %s under the level : %s", val, strings.Join(input.LevelStack[:i+1], "/")))
				}
			}
			_ = rootBkt.ForEach(func(k []byte, v []byte) error {
				if input.SearchKey != "" && !strings.Contains(strings.ToLower(string(k)), searchkey) {
					return nil
				}
				cntOfRecords += 1
				if cntOfRecords > 10000 {
					elem.ExceedsLimit = true
					return nil
				}
				if v == nil {
					elBkt := rootBkt.Bucket(k)
					bktCnt, pairCnt := getInlineBucketandPairCount(elBkt)
					resultFullSet = append(resultFullSet,
						model.Result{
							Name:          string(k),
							IsBucket:      true,
							Value:         "",
							NoOfChildBkts: bktCnt,
							NoOfPairs:     pairCnt,
						})
				} else {
					resultFullSet = append(resultFullSet,
						model.Result{
							Name:     string(k),
							IsBucket: false,
							Value:    r.unmarshal(v),
						})
				}
				return nil
			})
		} else {
			_ = tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				if input.SearchKey != "" && !strings.Contains(strings.ToLower(string(name)), searchkey) {
					return nil
				}
				cntOfRecords += 1
				if cntOfRecords > 10000 {
					elem.ExceedsLimit = true
					return nil
				}
				bktCnt, pairCnt := getInlineBucketandPairCount(b)
				resultFullSet = append(resultFullSet,
					model.Result{
						Name:          string(name),
						IsBucket:      true,
						Value:         "",
						NoOfChildBkts: bktCnt,
						NoOfPairs:     pairCnt,
					})

				return nil
			})
		}

		elem.LevelStack = input.LevelStack
		elem.SearchKey = input.SearchKey
		elem.Results = resultFullSet
		return nil
	})
	if err != nil {
		return model.ListedElem{}, err
	}
	return elem, nil
}

func getInlineBucketandPairCount(b *bolt.Bucket) (bktCnt, pairCnt int) {
	_ = b.ForEach(func(k, v []byte) error {
		if v == nil {
			bktCnt = bktCnt + 1
		} else {
			pairCnt = pairCnt + 1
		}
		return nil
	})
	return bktCnt, pairCnt
}
func (r *Repository) AddBuckets(input model.BucketsToAdd) (err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		var rootBkt *bolt.Bucket
		if len(input.LevelStack) > 0 {
			rootBkt = tx.Bucket([]byte(input.LevelStack[0]))
			if rootBkt == nil {
				return xerrors.New(fmt.Sprintf("No Root Bucket found by the name : %s", input.LevelStack[0]))
			}
			for i, val := range input.LevelStack[1:] {
				rootBkt = rootBkt.Bucket([]byte(val))
				if rootBkt == nil {
					return xerrors.New(fmt.Sprintf("No Bucket found by the name : %s under the level : %s", val, strings.Join(input.LevelStack[:i+1], "/")))
				}
			}
			for _, bkt := range input.Buckets {
				_, err = rootBkt.CreateBucket([]byte(bkt))
				if err != nil {
					return errors.Wrapf(err, "Unable to create bucket %s under stack %s", bkt, input.LevelStack)
				}
			}
		} else {
			for _, bkt := range input.Buckets {
				_, err = tx.CreateBucket([]byte(bkt))
				if err != nil {
					return errors.Wrapf(err, "Unable to create bucket %s under root", bkt)
				}
			}
		}
		return nil
	})
	return err
}

func (r *Repository) AddPairs(input model.PairsToAdd) (err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		var rootBkt *bolt.Bucket
		if len(input.LevelStack) == 0 {
			return errors.New("Cannot create key/value pairs without parent bucket, levelstack missing")
		}
		rootBkt = tx.Bucket([]byte(input.LevelStack[0]))
		if rootBkt == nil {
			return xerrors.New(fmt.Sprintf("No Root Bucket found by the name : %s", input.LevelStack[0]))
		}
		for i, val := range input.LevelStack[1:] {
			rootBkt = rootBkt.Bucket([]byte(val))
			if rootBkt == nil {
				return xerrors.New(fmt.Sprintf("No Bucket found by the name : %s under the level : %s", val, strings.Join(input.LevelStack[:i+1], "/")))
			}
		}
		for _, pair := range input.Pairs {
			val, err1 := json.Marshal(pair.Value)
			if err1 != nil {
				return errors.Wrapf(err1, "Unable to marshal the value %v", pair.Value)
			}
			err = rootBkt.Put([]byte(pair.Key), val)
			if err != nil {
				return errors.Wrapf(err, "Unable to create pair %s under stack %s", pair, input.LevelStack)
			}
		}
		return nil
	})
	return err
}

func (r *Repository) DeleteElement(input model.ItemToDelete) (err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		var rootBkt *bolt.Bucket
		if len(input.LevelStack) > 0 {
			rootBkt = tx.Bucket([]byte(input.LevelStack[0]))
			if rootBkt == nil {
				return xerrors.New(fmt.Sprintf("No Root Bucket found by the name : %s", input.LevelStack[0]))
			}
			for i, val := range input.LevelStack[1:] {
				rootBkt = rootBkt.Bucket([]byte(val))
				if rootBkt == nil {
					return xerrors.New(fmt.Sprintf("No Bucket found by the name : %s under the level : %s", val, strings.Join(input.LevelStack[:i+1], "/")))
				}
			}
			if bkt := rootBkt.Bucket([]byte(input.Key)); bkt != nil {
				err = rootBkt.DeleteBucket([]byte(input.Key))
				if err != nil {
					return errors.Wrapf(err, "Unable to delete bucket %s under stack %s", input.Key, input.LevelStack)
				}
			} else {
				err = rootBkt.Delete([]byte(input.Key))
				if err != nil {
					return errors.Wrapf(err, "Unable to delete key %s under stack %s", input.Key, input.LevelStack)
				}
			}
		} else {
			err = tx.DeleteBucket([]byte(input.Key))
			if err != nil {
				return errors.Wrapf(err, "Unable to delete bucket %s under root", input.Key)
			}
		}
		return nil
	})
	return err
}

func (r *Repository) RenameElement(input model.ItemToRename) (err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		var rootBkt *bolt.Bucket
		if len(input.LevelStack) > 0 {
			rootBkt = tx.Bucket([]byte(input.LevelStack[0]))
			if rootBkt == nil {
				return xerrors.New(fmt.Sprintf("No Root Bucket found by the name : %s", input.LevelStack[0]))
			}

			for i, val := range input.LevelStack[1:] {
				rootBkt = rootBkt.Bucket([]byte(val))
				if rootBkt == nil {
					return xerrors.New(fmt.Sprintf("No Bucket found by the name : %s under the level : %s", val, strings.Join(input.LevelStack[:i+1], "/")))
				}
			}

			val := rootBkt.Get([]byte(input.Key))
			if val != nil {
				err = rootBkt.Put([]byte(input.NewKey), val)
				if err != nil {
					return errors.Wrapf(err, "Unable to put key %s under stack %s", input.NewKey, input.LevelStack)
				}
				err = rootBkt.Delete([]byte(input.Key))
				if err != nil {
					return errors.Wrapf(err, "Unable to delete key %s under stack %s", input.Key, input.LevelStack)
				}
			} else {
				return errors.New("No Key found to be replaced")
			}
		} else {
			return errors.New("Buckets cannot be renamed, please provide level stack")
		}
		return nil
	})
	return err
}

func (r *Repository) UpdatePairValue(input model.ItemToUpdate) (err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		var rootBkt *bolt.Bucket
		if len(input.LevelStack) > 0 {
			rootBkt = tx.Bucket([]byte(input.LevelStack[0]))
			if rootBkt == nil {
				return xerrors.New(fmt.Sprintf("No Root Bucket found by the name : %s", input.LevelStack[0]))
			}
			for i, val := range input.LevelStack[1:] {
				rootBkt = rootBkt.Bucket([]byte(val))
				if rootBkt == nil {
					return xerrors.New(fmt.Sprintf("No Bucket found by the name : %s under the level : %s", val, strings.Join(input.LevelStack[:i+1], "/")))
				}
			}
			if val := rootBkt.Get([]byte(input.Key)); val != nil {
				val1, err1 := json.Marshal(input.NewValue)
				if err1 != nil {
					return errors.Wrapf(err1, "Unable to marshal the value %v", input.NewValue)
				}
				err = rootBkt.Put([]byte(input.Key), val1)
				if err != nil {
					return errors.Wrap(err, "Unable to put new value")
				}
			} else {
				return errors.New("Given key not found")
			}
		} else {
			return errors.New("Please provide level stack")
		}
		return nil
	})
	return err
}
