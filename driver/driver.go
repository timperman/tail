package driver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

  "gopkg.in/fsnotify.v1"
  "github.com/timperman/tail/util"
	"github.com/timperman/tail/tailcmd"
)

const (
	DriverName         = "tail"
	VolumeDataPathName = "_data"
	volumesPathName    = "volumes"
)

type VolumeDriver struct {
	name    string
	base    string
	path    string
	events  chan<- []byte
  volumes map[string]*volume
	tailcmds map[string]*tailcmd.TailCmd
  watcher *fsnotify.Watcher
}

type volume struct {
	name string
	path string
}

func New(base string, events chan<- []byte) (*VolumeDriver, error) {
	root := filepath.Join(base, volumesPathName)

	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}

  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    return nil, err
  }

	driver := &VolumeDriver{
		name:    DriverName,
		base:    base,
		path:    root,
    volumes: make(map[string]*volume),
    tailcmds: make(map[string]*tailcmd.TailCmd),
    watcher: watcher,
	}

	dirs, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		name := filepath.Base(dir.Name())
    path := driver.getPath(name)
		driver.volumes[name] = &volume{
			name: name,
			path: path,
		}
    watcher.Add(path)
	}

  go func() {
    for {
      select {
      case event := <-watcher.Events:
        log.Printf("fsnotify event- op: %s name: %s", event.Op, event.Name)
        switch event.Op {
        case fsnotify.Create:
          if tc, err := tailcmd.TailPipe(event.Name, events); err != nil {
            log.Printf("error starting tail command: %v\n", err)
          } else {
            driver.tailcmds[event.Name] = tc
          }
        case fsnotify.Remove:
          if tc, ok := driver.tailcmds[event.Name]; ok {
            tc.Stop()
          }
          delete(driver.tailcmds, event.Name)
        }
      case err := <-watcher.Errors:
        log.Printf("fsnotify error: %v\n", err)
      }
    }
  }()

	return driver, nil
}

func (d *VolumeDriver) getPath(name string) string {
	return filepath.Join(d.path, name, VolumeDataPathName)
}

func (d *VolumeDriver) Create(w http.ResponseWriter, r *http.Request) {
	req, err := util.JSONDecode(r)
	if err != nil {
		util.JSONResponse(w, map[string]interface{}{"Err": err})
		return
	}

	log.Printf("Create request: %v\n", req)

	name := req["Name"].(string)

	if _, found := d.volumes[name]; found {
		util.JSONResponse(w, map[string]interface{}{"Err": nil})
		return
	}

	path := d.getPath(name)
	if err := os.MkdirAll(path, 0755); err != nil {
		if os.IsExist(err) {
			util.JSONResponse(w, map[string]interface{}{"Err": fmt.Errorf("volume already exists under %s", filepath.Dir(path))})
			return
		}
		util.JSONResponse(w, map[string]interface{}{"Err": err})
		return
	}

	d.volumes[name] = &volume{
		name: name,
		path: path,
	}

	util.JSONResponse(w, map[string]interface{}{"Err": nil})
}

func (d *VolumeDriver) Remove(w http.ResponseWriter, r *http.Request) {
	req, err := util.JSONDecode(r)
	if err != nil {
		util.JSONResponse(w, map[string]interface{}{"Err": err})
		return
	}

	log.Printf("Remove request: %v\n", req)

	name := req["Name"].(string)
	v, found := d.volumes[name]
	if !found {
		util.JSONResponse(w, map[string]interface{}{"Err": fmt.Errorf("Volume %s not found", name)})
		return
	}

	realPath, err := filepath.EvalSymlinks(v.path)
	if err != nil {
		if !os.IsNotExist(err) {
			util.JSONResponse(w, map[string]interface{}{"Err": err})
			return
		}
		realPath = filepath.Dir(v.path)
	}

	if !d.scopedPath(realPath) {
		util.JSONResponse(w, map[string]interface{}{"Err": fmt.Errorf("Unable to remove a directory of out the Docker root %s: %s", d.base, realPath)})
		return
	}

	if err := removePath(realPath); err != nil {
		util.JSONResponse(w, map[string]interface{}{"Err": err})
		return
	}

	delete(d.volumes, v.name)
	if err = removePath(filepath.Dir(v.path)); err != nil {
		util.JSONResponse(w, map[string]interface{}{"Err": err})
	} else {
		util.JSONResponse(w, map[string]interface{}{"Err": nil})
	}
}

var oldVfsDir = filepath.Join("vfs", "dir")

func (d *VolumeDriver) scopedPath(realPath string) bool {
	// Volumes path for Docker version >= 1.7
	if strings.HasPrefix(realPath, filepath.Join(d.base, volumesPathName)) && realPath != filepath.Join(d.base, volumesPathName) {
		return true
	}

	// Volumes path for Docker version < 1.7
	if strings.HasPrefix(realPath, filepath.Join(d.base, oldVfsDir)) {
		return true
	}

	return false
}

func removePath(path string) error {
	if err := os.RemoveAll(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (d *VolumeDriver) Mount(w http.ResponseWriter, r *http.Request) {
	req, err := util.JSONDecode(r)
	log.Printf("Mount request: %v\n", req)
	util.JSONResponse(w, map[string]interface{}{"Mountpoint": req["Name"], "Err": err})
}

func (d *VolumeDriver) Unmount(w http.ResponseWriter, r *http.Request) {
	req, err := util.JSONDecode(r)
	log.Printf("Unmount request: %v\n", req)
	util.JSONResponse(w, map[string]interface{}{"Err": err})
}

func (d *VolumeDriver) Path(w http.ResponseWriter, r *http.Request) {
	req, err := util.JSONDecode(r)
	log.Printf("Path request: %v\n", req)
	util.JSONResponse(w, map[string]interface{}{"Mountpoint": req["Name"], "Err": err})
}
