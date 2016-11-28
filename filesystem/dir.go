package filesystem

import (
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fntlnz/gridfsmount/datastore"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type Dir struct {
	ds   *datastore.GridFSDataStore
	name string
}

func NewDir(ds *datastore.GridFSDataStore, name string) (*Dir, error) {
	return &Dir{
		ds:   ds,
		name: name,
	}, nil
}

func (dir *Dir) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	logrus.Debug("Getxattr dir: " + dir.name)
	return nil
}

func (dir *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	logrus.Warn("DIR ATTR IST:"+dir.name, a)
	a.Inode = 2
	a.Mode = os.ModeDir | 0555
	return nil
}

func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	file, err := dir.ds.FindByName(name)

	if err != nil {
		return nil, fuse.ENOENT
	}

	defer file.Close()

	logrus.Info("FILE IST:" + name)
	node, err := NewFile(dir.ds, file.Name())

	if err != nil {
		logrus.Errorf("Error creating file: %s", err.Error())
		return nil, fuse.EIO
	}

	return node, nil
}

func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	logrus.Info("nun im listing drin")
	files, err := dir.ds.ListFileNames()

	if err != nil {
		return nil, fuse.ENOENT
	}

	var de []fuse.Dirent
	for _, file := range files {
		logrus.Info("nun im listing drin: " + file)
		if strings.HasSuffix(file, "/") {

			file = file[:len(file)-1]
			de = append(de, fuse.Dirent{
				Inode: 2,
				Name:  file,
				Type:  fuse.DT_Dir,
			})
		} else {
			de = append(de, fuse.Dirent{
				Inode: 2,
				Name:  file,
				Type:  fuse.DT_File,
			})
		}
	}
	return de, nil
}

func (dir *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {

	logrus.Error("drin")
	file, err := dir.ds.Create(req.Name)

	if err != nil {
		logrus.Errorf("An error occurred creating file into the datastore: %s", err.Error())
		return nil, nil, fuse.EIO
	}

	defer file.Close()

	node, err := NewFile(dir.ds, file.Name())

	if err != nil {
		logrus.Errorf("An error occurred creating the file: %s", err.Error())
		return nil, nil, fuse.EIO
	}

	return node, node, nil

}

func (dir *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	logrus.Error("nun drin ne?")

	file, err := dir.ds.Create(req.Name + "/")

	node, err := NewDir(dir.ds, file.Name())

	if err != nil {
		logrus.Errorf("An error occurred creating the file: %s", err.Error())
		return nil, fuse.EIO
	}

	defer file.Close()

	logrus.Debug("NEW FOLDER: " + node.name)
	return node, nil
}
