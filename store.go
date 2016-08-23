package main

import (
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/btlike/repository"
	"github.com/xgfone/gobt/conf"
)

type Files []repository.File

func (a Files) Len() int           { return len(a) }
func (a Files) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Files) Less(i, j int) bool { return a[i].Length > a[j].Length }

func storeTorrent(data interface{}, infohash []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e
		}
	}()

	if info, ok := data.(map[string]interface{}); ok {
		var t repository.Torrent
		t.CreateTime = time.Now()

		// get name
		if name, ok := info["name"].(string); ok {
			t.Name = name
			if t.Name == "" {
				return fmt.Errorf("store name len is 0")
			}
		}

		// get infohash
		t.Infohash = hex.EncodeToString(infohash)
		if len(t.Infohash) != 40 {
			return fmt.Errorf("store infohash len is not 40")
		}

		// get files
		if v, ok := info["files"]; !ok {
			t.Length = int64(info["length"].(int))
			t.FileCount = 1
			t.Files = append(t.Files, repository.File{Name: t.Name, Length: t.Length})
		} else {
			var tmpFiles Files
			files := v.([]interface{})
			tmpFiles = make(Files, len(files))
			for i, item := range files {
				fl := item.(map[string]interface{})
				flName := fl["path"].([]interface{})
				tmpFiles[i] = repository.File{
					Name:   flName[0].(string),
					Length: int64(fl["length"].(int)),
				}
			}
			sort.Sort(tmpFiles)

			for k, v := range tmpFiles {
				if len(v.Name) > 0 {
					t.Length += v.Length
					t.FileCount++
					if k < 5 {
						t.Files = append(t.Files, repository.File{
							Name:   v.Name,
							Length: v.Length,
						})
					}
				}
			}
		}

		err = conf.Repository.CreateTorrent(t)
	}

	return
}

func checkTorrent(infohash []byte) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			ok = false
		}
	}()

	ok = true
	Infohash := hex.EncodeToString(infohash)
	if t, err := GetTorrentByInfohash(Infohash); err != nil || t.Infohash != Infohash {
		ok = false
	}
}