package filesys

import (
	"time"

	"appengine/memcache"
)

func (fso *Directory) MemCacheSet() {

	miPut := &memcache.Item{
		Key:        fso.SKey,
		Value:      []byte("anything"), // sadly - value is ignored
		Object:     &fso,
		Expiration: 3600 * time.Second,
	}
	err := memcache.JSON.Set(fso.Fs.c, miPut)
	if err != nil {
		fso.Fs.c.Errorf("fso memcachd %v - key %v", err, fso.SKey)
	} else {
		// fso.Fs.c.Infof("fso memcachd - key %v", fso.SKey)
	}
}
func (fso *File) MemCacheSet() {

	miPut := &memcache.Item{
		Key:        fso.SKey,
		Value:      []byte("anything"), // sadly - value is ignored
		Object:     &fso,
		Expiration: 3600 * time.Second,
	}
	err := memcache.JSON.Set(fso.Fs.c, miPut)
	if err != nil {
		fso.Fs.c.Errorf("fso memcachd %v - key %v", err, fso.SKey)
	} else {
		// fso.Fs.c.Infof("fso memcachd - key %v", fso.SKey)
	}
}
